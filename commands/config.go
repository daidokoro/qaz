package commands

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	stks "github.com/daidokoro/qaz/stacks"

	"github.com/daidokoro/hcl"
)

// Configure parses the config file abd setos stacjs abd ebv
func Configure(confSource string, conf string) (err error) {

	// set config session
	config.Session, err = manager.GetSess(run.profile)
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
		})

		stacks.MustGet(s).SetStackName()

		// set session
		stacks.MustGet(s).Session, err = manager.GetSess(stacks.MustGet(s).Profile)
		if err != nil {
			return
		}

		// stacks.MustGet(s).Session = sess

		// set parameters and tags, if any
		config.Parameters(stacks.MustGet(s)).Tags(stacks.MustGet(s))

	}

	return
}
