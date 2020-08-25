package stacks

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/daidokoro/qaz/bucket"
	"github.com/daidokoro/qaz/log"
	"github.com/daidokoro/qaz/utils"
)

// Config type for handling yaml config files
type Config struct {
	Session           *session.Session `yaml:"-" json:"" hcl:""`
	String            string           `yaml:"-" json:"-" hcl:"-"`
	Region            string           `yaml:"region,omitempty" json:"region,omitempty" hcl:"region,omitempty"`
	Project           string           `yaml:"project" json:"project" hcl:"project"`
	GenerateDelimiter string           `yaml:"gen_time,omitempty" json:"gen_time,omitempty" hcl:"gen_time,omitempty"`
	DeployDelimiter   string           `yaml:"deploy_time,omitempty" json:"deploy_time,omitempty" hcl:"deploy_time,omitempty"`
	S3Package         map[string]struct {
		Source      string `yaml:"source,omitempty" json:"source,omitempty" hcl:"source,omitempty"`
		Destination string `yaml:"destination,omitempty" json:"destination,omitempty" hcl:"destination,omitempty"`
	} `yaml:"s3_package,omitempty" json:"s3_package,omitempty" hcl:"s3_package,omitempty"`
	Global map[string]interface{} `yaml:"global,omitempty" json:"global,omitempty" hcl:"global,omitempty"`
	Stacks map[string]struct {
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
		CFRoleArn        string                 `yaml:"cf-role-arn" json:"cf-role-arn" hcl:"cf-role-arn"`
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

	log.Debug("calling functions in config file")

	// create template
	t, err := template.New("config-template").Delims(`{{`, `}}`).Funcs(fmap).Parse(c.String)
	if err != nil {
		return err
	}

	// so that we can write to string
	var doc bytes.Buffer

	t.Execute(&doc, nil)
	c.String = doc.String()
	log.Debug("config: %s", c.String)
	return nil
}

// used only for concurrent packaging in the function below
type s3Package struct {
	name string
	src  string
	dest string
}

// PackageToS3 - executes package to s3 if --package/-p flag is given
// and packages are defined.
func (c *Config) PackageToS3(packageFlag bool) (err error) {
	if !packageFlag {
		return
	}

	var wg sync.WaitGroup
	jobs := make(chan *s3Package, 5)
	errChan := make(chan error, 3*len(c.S3Package)) // concurrent functions x the number of packages.... just because

	// 3 concurrecnt packaging functions
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			var e error
			defer func() {
				errChan <- e
				wg.Done()
			}()

			for p := range jobs {

				log.Info("s3_package [%s]", p.name)
				u, e := url.Parse(p.dest)
				if e != nil {
					return
				}

				// check url schme for s3
				if strings.ToLower(u.Scheme) != "s3" {
					e = fmt.Errorf("non s3 url detected for package destination: %s", u.Scheme)
					return
				}

				log.Info("creating ZIP package from : [%s]", p.src)
				buf, e := utils.Zip(p.src)
				if e != nil {
					return
				}

				log.Info("uploading package to s3: [%s]", p.dest)
				if _, e = bucket.S3write(u.Host, u.Path, bytes.NewReader(buf.Bytes()), c.Session); err != nil {
					return
				}
				return
			}

			return
		}()
	}

	for k, v := range c.S3Package {
		jobs <- &s3Package{k, v.Source, v.Destination}
	}

	log.Debug("closing concurrent s3 packaging jobs")
	close(jobs)
	wg.Wait()
	close(errChan)

	// check for errors
	for e := range errChan {
		if e != nil {
			err = e
		}
	}
	return
}
