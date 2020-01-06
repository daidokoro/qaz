package stacks

import (
	"bytes"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Config type for handling yaml config files
type Config struct {
	Session           *session.Session       `yaml:"-" json:"" hcl:""`
	String            string                 `yaml:"-" json:"-" hcl:"-"`
	Region            string                 `yaml:"region,omitempty" json:"region,omitempty" hcl:"region,omitempty"`
	Project           string                 `yaml:"project" json:"project" hcl:"project"`
	GenerateDelimiter string                 `yaml:"gen_time,omitempty" json:"gen_time,omitempty" hcl:"gen_time,omitempty"`
	DeployDelimiter   string                 `yaml:"deploy_time,omitempty" json:"deploy_time,omitempty" hcl:"deploy_time,omitempty"`
	Global            map[string]interface{} `yaml:"global,omitempty" json:"global,omitempty" hcl:"global,omitempty"`
	Stacks            map[string]struct {
		DependsOn        []string               `yaml:"depends_on,omitempty" json:"depends_on,omitempty" hcl:"depends_on,omitempty"`
		Parameters       []map[string]string    `yaml:"parameters,omitempty" json:"parameters,omitempty" hcl:"parameters,omitempty"`
		Policy           string                 `yaml:"policy,omitempty" json:"policy,omitempty" hcl:"policy,omitempty"`
		Profile          string                 `yaml:"profile,omitempty" json:"profile,omitempty" hcl:"profile,omitempty"`
		Region           string                 `yaml:"region,omitempty" json:"region,omitempty" hcl:"region,omitempty"`
		Source           string                 `yaml:"source,omitempty" json:"source,omitempty" hcl:"source,omitempty"`
		Name             string                 `yaml:"name,omitempty" json:"name,omitempty" hcl:"name,omitempty"`
		Bucket           string                 `yaml:"bucket,omitempty" json:"bucket,omitempty" hcl:"bucket,omitempty"`
		Role             string                 `yaml:"role,omitempty" json:"role,omitempty" hcl:"role,omitempty"`
		Tags             []map[string]string    `yaml:"tags,omitempty" json:"tags,omitempty" hcl:"tags,omitempty"`
		Timeout          int64                  `yaml:"timeout,omitempty" json:"timeout,omitempty" hcl:"timeout,omitempty"`
		NotificationARNs []string               `yaml:"notification-arns" json:"notification-arns" hcl:"notification-arns"`
		CF               map[string]interface{} `yaml:"cf,omitempty" json:"cf,omitempty" hcl:"cf,omitempty"`
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

// Parameters - Adds parameters to given stack based on config
func (c *Config) Parameters(s *Stack) *Config {

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

	return c
}

// Tags Adds stack tags to given stack based on config
func (c *Config) Tags(s *Stack) *Config {

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
	return c
}

// CallFunctions - execute gentime/deploytime functions in config
func (c *Config) CallFunctions(fmap template.FuncMap) error {

	Log.Debug("calling functions in config file")

	// create template
	t, err := template.New("config-template").Delims(`{{`, `}}`).Funcs(fmap).Parse(c.String)
	if err != nil {
		return err
	}

	// so that we can write to string
	var doc bytes.Buffer

	t.Execute(&doc, nil)
	c.String = doc.String()
	Log.Debug("config: %s", c.String)
	return nil
}
