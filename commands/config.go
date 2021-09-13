package commands

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/daidokoro/qaz/log"
	"github.com/daidokoro/qaz/stacks"

	yaml "gopkg.in/yaml.v2"

	"github.com/daidokoro/hcl"
)

// Configure parses the config file and string and returns a stacks.Map
func Configure(confSource string, conf string) (stks stacks.Map, err error) {

	// set config session
	config.Session, err = GetSession()
	if err != nil {
		return
	}

	if conf == "" {
		// utilise FetchSource to get sources
		if err = stacks.FetchSource(confSource, &config); err != nil {
			return
		}
	} else {
		config.String = conf
	}

	// execute config Functions
	if err = config.CallFunctions(GenTimeFunctions); err != nil {
		err = fmt.Errorf("failed to run template functions in config: %s", err)
		return
	}

	// decrypt config secrets
	if err = config.SopsDecrypt(); err != nil {
		err = fmt.Errorf("failed to decrypt secrets in config: %s", err)
		return
	}

	log.Debug("checking Config for HCL format...")
	if err = hcl.Unmarshal([]byte(config.String), &config); err != nil {
		// fmt.Println(err)
		log.Debug("failed to parse hcl... moving to JSON/YAML... error: %v", err)

		if err = yaml.Unmarshal([]byte(config.String), &config); err != nil {
			return
		}
	}

	log.Debug("Config File Read: %s", config.Project)

	// stacks = make(map[string]*stks.Stack)

	// Get Stack Values
	for s, v := range config.Stacks {
		stks.Add(s, &stacks.Stack{
			Name:             s,
			Profile:          v.Profile,
			Region:           v.Region,
			DependsOn:        v.DependsOn,
			Policy:           v.Policy,
			Source:           v.Source,
			Stackname:        v.Name,
			Bucket:           v.Bucket,
			Role:             v.Role,
			DeployDelims:     &config.DeployDelimiter,
			GenDelims:        &config.GenerateDelimiter,
			TemplateValues:   config.Vars(),
			GenTimeFunc:      &GenTimeFunctions,
			DeployTimeFunc:   &DeployTimeFunctions,
			Project:          &config.Project,
			Timeout:          v.Timeout,
			NotificationARNs: v.NotificationARNs,
			CFRoleARN:        v.CFRoleArn,
		})

		stks.MustGet(s).SetStackName()

		// set session
		stks.MustGet(s).Session, err = GetSession(func(opts *session.Options) {
			if stks.MustGet(s).Profile != "" {
				opts.Profile = stks.MustGet(s).Profile
			}

			// use config region
			if config.Region != "" {
				opts.Config.Region = aws.String(config.Region)
			}

			// stack region trumps all other regions if-set
			if stks.MustGet(s).Region != "" {
				opts.Config.Region = aws.String(stks.MustGet(s).Region)
			}

			return
		})

		if err != nil {
			return
		}

		// set parameters and tags, if any
		config.
			Parameters(stks.MustGet(s)).
			Tags(stks.MustGet(s))

		// add stack based functions stackoutput
		stks.AddFuncs(DeployTimeFunctions)
	}

	if len(config.S3Package) > 0 {
		err = config.PackageToS3(run.executePackage)
	}

	return
}
