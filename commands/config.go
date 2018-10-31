package commands

import (
	"fmt"
	"os"
	"sync"

	"github.com/daidokoro/qaz/troposphere"

	yaml "gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	stks "github.com/daidokoro/qaz/stacks"

	"github.com/daidokoro/hcl"
)

var once sync.Once

// Configure parses the config file abd setos stacjs abd ebv
func Configure(confSource string, conf string) (err error) {

	// set config session
	config.Session, err = GetSession()
	if err != nil {
		return
	}

	if conf == "" {
		// utilise FetchSource to get sources
		if err = stks.FetchSource(confSource, &config); err != nil {
			return err
		}
	} else {
		config.String = conf
	}

	// execute config Functions
	if err = config.CallFunctions(GenTimeFunctions); err != nil {
		return fmt.Errorf("failed to run template functions in config: %s", err)
	}

	log.Debug("checking Config for HCL format...")
	if err = hcl.Unmarshal([]byte(config.String), &config); err != nil {
		// fmt.Println(err)
		log.Debug("failed to parse hcl... moving to JSON/YAML... error: %v", err)
		if err = yaml.Unmarshal([]byte(config.String), &config); err != nil {
			return err
		}
	}

	log.Debug("Config File Read: %s", config.Project)

	// stacks = make(map[string]*stks.Stack)

	// Get Stack Values
	for s, v := range config.Stacks {
		stacks.Add(s, &stks.Stack{
			Name:           s,
			Profile:        v.Profile,
			Region:         v.Region,
			DependsOn:      v.DependsOn,
			Policy:         v.Policy,
			Source:         v.Source,
			Bucket:         v.Bucket,
			Role:           v.Role,
			DeployDelims:   &config.DeployDelimiter,
			GenDelims:      &config.GenerateDelimiter,
			TemplateValues: config.Vars(),
			GenTimeFunc:    &GenTimeFunctions,
			DeployTimeFunc: &DeployTimeFunctions,
			Project:        &config.Project,
			Timeout:        v.Timeout,
			Troposphere:    v.Troposphere,
		})

		if v.Troposphere {
			once.Do(func() {
				log.Debug("troposphere stack(s) detected")
				if err := troposphere.BuildImage(); err != nil {
					log.Error("error building troposphere container: %v", err)
					os.Exit(1)
				}
			})
		}

		stacks.MustGet(s).SetStackName()

		// set session
		stacks.MustGet(s).Session, err = GetSession(func(opts *session.Options) {
			if stacks.MustGet(s).Profile != "" {
				opts.Profile = stacks.MustGet(s).Profile
			}

			// use config region
			if config.Region != "" {
				opts.Config.Region = aws.String(config.Region)
			}

			// stack region trumps all other regions if-set
			if stacks.MustGet(s).Region != "" {
				opts.Config.Region = aws.String(stacks.MustGet(s).Region)
			}

			return
		})

		if err != nil {
			return
		}

		// stacks.MustGet(s).Session = sess

		// set parameters and tags, if any
		config.Parameters(stacks.MustGet(s)).Tags(stacks.MustGet(s))

	}

	return
}
