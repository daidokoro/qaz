package commands

import (
	"bytes"
	"fmt"
	"text/template"

	yaml "gopkg.in/yaml.v2"

	stks "github.com/daidokoro/qaz/stacks"
	"github.com/daidokoro/qaz/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/daidokoro/hcl"
)

// Config type for handling yaml config files
type Config struct {
	String            string                 `yaml:"-" json:"-" hcl:"-"`
	Region            string                 `yaml:"region,omitempty" json:"region,omitempty" hcl:"region,omitempty"`
	Project           string                 `yaml:"project" json:"project" hcl:"project"`
	GenerateDelimiter string                 `yaml:"gen_time,omitempty" json:"gen_time,omitempty" hcl:"gen_time,omitempty"`
	DeployDelimiter   string                 `yaml:"deploy_time,omitempty" json:"deploy_time,omitempty" hcl:"deploy_time,omitempty"`
	Global            map[string]interface{} `yaml:"global,omitempty" json:"global,omitempty" hcl:"global,omitempty"`
	Stacks            map[string]struct {
		DependsOn  []string               `yaml:"depends_on,omitempty" json:"depends_on,omitempty" hcl:"depends_on,omitempty"`
		Parameters []map[string]string    `yaml:"parameters,omitempty" json:"parameters,omitempty" hcl:"parameters,omitempty"`
		Policy     string                 `yaml:"policy,omitempty" json:"policy,omitempty" hcl:"policy,omitempty"`
		Profile    string                 `yaml:"profile,omitempty" json:"profile,omitempty" hcl:"profile,omitempty"`
		Source     string                 `yaml:"source,omitempty" json:"source,omitempty" hcl:"source,omitempty"`
		Bucket     string                 `yaml:"bucket,omitempty" json:"bucket,omitempty" hcl:"bucket,omitempty"`
		Role       string                 `yaml:"role,omitempty" json:"role,omitempty" hcl:"role,omitempty"`
		Tags       []map[string]string    `yaml:"tags,omitempty" json:"tags,omitempty" hcl:"tags,omitempty"`
		Timeout    int64                  `yaml:"timeout,omitempty" json:"timeout,omitempty" hcl:"timeout,omitempty"`
		CF         map[string]interface{} `yaml:"cf,omitempty" json:"cf,omitempty" hcl:"cf,omitempty"`
	} `yaml:"stacks" json:"stacks" hcl:"stacks"`
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

// Adds stack tags to given stack based on config
func (c *Config) tags(s *stks.Stack) {

	for stk, val := range c.Stacks {
		if s.Name == stk {
			for _, param := range val.Tags {
				for k, v := range param {
					s.Tags = append(s.Tags, &cloudformation.Tag{
						Key:   aws.String(k),
						Value: aws.String(v),
					})
				}

			}
		}
	}
}

// execute gentime/deploytime functions in config
func (c *Config) callFunctions() error {

	log.Debug("calling functions in config file")
	// define Delims
	left, right := func() (string, string) {
		if utils.IsJSON(c.String) || utils.IsHCL(c.String) {
			return "${", "}"
		}
		return "!", "\n"
	}()

	// create template
	t, err := template.New("config-template").Delims(left, right).Funcs(GenTimeFunctions).Parse(c.String)
	if err != nil {
		return err
	}

	// so that we can write to string
	var doc bytes.Buffer

	t.Execute(&doc, nil)
	c.String = doc.String()
	log.Debug(fmt.Sprintln("config:", c.String))
	return nil
}

// Configure parses the config file abd setos stacjs abd ebv
func Configure(confSource string, conf string) error {

	if conf == "" {
		cfg, err := fetchContent(confSource)
		if err != nil {
			return err
		}

		config.String = cfg
	}

	// execute Functions
	if err := config.callFunctions(); err != nil {
		return fmt.Errorf("failed to run template functions in config: %s", err)
	}

	log.Debug("checking Config for HCL format...")
	if err := hcl.Unmarshal([]byte(config.String), &config); err != nil {
		// fmt.Println(err)
		log.Debug(fmt.Sprintln("failed to parse hcl... moving to JSON/YAML...", err.Error()))
		if err := yaml.Unmarshal([]byte(config.String), &config); err != nil {
			return err
		}
	}

	log.Debug(fmt.Sprintln("Config File Read:", config.Project))

	stacks = make(map[string]*stks.Stack)

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
			GenTimeFunc:    &GenTimeFunctions,
			DeployTimeFunc: &DeployTimeFunctions,
			Project:        &config.Project,
			Timeout:        v.Timeout,
		}

		stacks[s].SetStackName()

		// set session
		sess, err := manager.GetSess(stacks[s].Profile)
		if err != nil {
			return err
		}

		stacks[s].Session = sess

		// set parameters and tags, if any
		config.parameters(stacks[s])
		config.tags(stacks[s])

	}

	return nil
}
