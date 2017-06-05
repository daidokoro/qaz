package commands

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	stks "qaz/stacks"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Config type for handling yaml config files
type Config struct {
	Region            string                 `yaml:"region,omitempty" json:"region,omitempty"`
	Project           string                 `yaml:"project" json:"project"`
	GenerateDelimiter string                 `yaml:"gen_time,omitempty" json:"gen_time,omitempty"`
	DeployDelimiter   string                 `yaml:"deploy_time,omitempty" json:"deploy,omitempty"`
	Global            map[string]interface{} `yaml:"global,omitempty" json:"global,omitempty"`
	Stacks            map[string]struct {
		DependsOn  []string               `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
		Parameters []map[string]string    `yaml:"parameters,omitempty" json:"parameters,omitempty"`
		Policy     string                 `yaml:"policy,omitempty" json:"policy,omitempty"`
		Profile    string                 `yaml:"profile,omitempty" json:"profile,omitempty"`
		Source     string                 `yaml:"source,omitempty" json:"source,omitempty"`
		Bucket     string                 `yaml:"bucket,omitempty" json:"bucket,omitempty"`
		Role       string                 `yaml:"role,omitempty" json:"role,omitempty"`
		CF         map[string]interface{} `yaml:"cf,omitempty" json:"cf,omitempty"`
	} `yaml:"stacks" json:"stacks"`
}

// Vars Returns map string of config values
func (c *Config) Vars() map[string]interface{} {
	m := make(map[string]interface{})
	m["global"] = c.Global
	m["region"] = c.Region
	m["project"] = c.Project

	for s, v := range c.Stacks {
		m[s] = v.CF
	}

	return m
}

// Adds parameters to given stack based on config
func (c *Config) parameters(s *stks.Stack) {

	for stk, val := range c.Stacks {
		if s.Name == stk {
			for _, param := range val.Parameters {
				for k, v := range param {
					s.Parameters = append(s.Parameters, &cloudformation.Parameter{
						ParameterKey:   aws.String(k),
						ParameterValue: aws.String(v),
					})
				}

			}
		}
	}
}

// configure parses the config file abd setos stacjs abd ebv
func configure(confSource string, conf string) error {

	if conf == "" {
		cfg, err := fetchContent(confSource)
		if err != nil {
			return err
		}

		conf = cfg
	}

	if err := yaml.Unmarshal([]byte(conf), &config); err != nil {
		return err
	}

	log.Debug(fmt.Sprintln("Config File Read:", config))

	stacks = make(map[string]*stks.Stack)

	// add logging
	stks.Log = &log

	// add repo
	stks.Git = &gitrepo

	// Get Stack Values
	for s, v := range config.Stacks {
		stacks[s] = &stks.Stack{
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
			GenTimeFunc:    &genTimeFunctions,
			DeployTimeFunc: &deployTimeFunctions,
			Project:        &config.Project,
		}

		stacks[s].SetStackName()

		// set session
		sess, err := manager.GetSess(stacks[s].Profile)
		if err != nil {
			return err
		}

		stacks[s].Session = sess

		// set parameters, if any
		config.parameters(stacks[s])
	}

	return nil
}
