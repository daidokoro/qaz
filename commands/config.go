package commands

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var config Config

// Config type for handling yaml config files
type Config struct {
	Region  string                 `yaml:"region,omitempty"`
	Project string                 `yaml:"project"`
	Global  map[string]interface{} `yaml:"global,omitempty"`
	Stacks  map[string]struct {
		DependsOn  []string               `yaml:"depends_on,omitempty"`
		Parameters []map[string]string    `yaml:"parameters,omitempty"`
		Policy     string                 `yaml:"policy,omitempty"`
		Profile    string                 `yaml:"profile,omitempty"`
		Source     string                 `yaml:"source,omitempty"`
		CF         map[string]interface{} `yaml:"cf"`
	} `yaml:"stacks"`
}

// Returns map string of config values
func (c *Config) vars() map[string]interface{} {
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
func (c *Config) parameters(s *stack) {

	for stk, val := range c.Stacks {
		if s.name == stk {
			for _, param := range val.Parameters {
				for k, v := range param {
					s.parameters = append(s.parameters, &cloudformation.Parameter{
						ParameterKey:   aws.String(k),
						ParameterValue: aws.String(v),
					})
				}

			}
		}
	}
}

// Read template source and sets the template value in given stack
func (c *Config) getSource(s *stack) error {
	return nil
}

// configReader parses the config YAML file with Viper
func configReader(conf string) error {

	cfg, err := fetchContent(conf)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal([]byte(cfg), &config); err != nil {
		return err
	}

	Log(fmt.Sprintln("Config File Read:", config), level.debug)

	stacks = make(map[string]*stack)

	// Get Stack Values
	for s, v := range config.Stacks {
		stacks[s] = &stack{}
		stacks[s].name = s
		stacks[s].setStackName()
		stacks[s].dependsOn = v.DependsOn
		stacks[s].policy = v.Policy
		stacks[s].profile = v.Profile
		stacks[s].source = v.Source

		// set session
		sess, err := manager.GetSess(stacks[s].profile)
		if err != nil {
			return err
		}

		stacks[s].session = sess

		// set parameters, if any
		config.parameters(stacks[s])
	}

	return nil
}
