package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spf13/viper"
)

// global variables
var (
	region    string
	project   string
	stackname string
	stacks    map[string]*stack
	cfvars    map[string]interface{}
)

// used for stack keyword referencing
var keyword = struct {
	depends        string
	parameters     string
	cloudformation string
}{
	"depends_on",
	"parameters",
	"cf",
}

// fetchContent - checks the source type, url/s3/file and calls the corresponding function
func fetchContent(source string) (string, error) {
	switch strings.Split(strings.ToLower(source), ":")[0] {
	case "http":
		Log(fmt.Sprintln("Source Type: [http] Detected, Fetching Source: ", source), level.debug)
		resp, err := Get(source)
		if err != nil {
			return "", err
		}
		return resp, nil
	case "https":
		Log(fmt.Sprintln("Source Type: [https] Detected, Fetching Source: ", source), level.debug)
		resp, err := Get(source)
		if err != nil {
			return "", err
		}
		return resp, nil
	case "s3":
		Log(fmt.Sprintln("Source Type: [s3] Detected, Fetching Source: ", source), level.debug)
		resp, err := S3Read(source)
		if err != nil {
			return "", err
		}
		return resp, nil
	default:
		Log(fmt.Sprintln("Source Type: [file] Detected, Fetching Source: ", source), level.debug)
		b, err := ioutil.ReadFile(source)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

// getName  - Checks if arg is url or file and returns stack name and filepath/url
func getSource(s string) (string, string, error) {
	if strings.Contains(s, "::") {
		vals := strings.Split(s, "::")
		if len(vals) < 2 {
			return "", "", errors.New(`Error, invalid url format --> Example: stackname::http://someurl OR stackname::s3://bucket/key`)
		}

		return vals[0], vals[1], nil

	}

	name := filepath.Base(strings.Replace(s, filepath.Ext(s), "", -1))
	return name, s, nil

}

// configReader parses the config YAML file with Viper
func configReader(conf string) error {
	viper.SetConfigType("yaml")

	cfg, err := fetchContent(conf)

	if err != nil {
		return err
	}

	Log(fmt.Sprintln("Reading Config String into viper:", cfg), level.debug)
	err = viper.ReadConfig(bytes.NewBuffer([]byte(cfg)))

	if err != nil {
		return err
	}

	// Get basic settings
	region = viper.GetString("region")
	project = viper.GetString("project")

	// Get all the values of variables under cloudformation and put them into a map
	cfvars = make(map[string]interface{})
	stacks = make(map[string]*stack)

	// Get Global values
	cfvars["global"] = viper.Get("global")

	Log(fmt.Sprintln("Keys identified in Config:", viper.AllKeys()), level.debug)

	// Get Stack Values
	for _, item := range viper.Sub("stacks").AllKeys() {
		s := strings.Split(item, ".")[0]

		if _, ok := stacks[s]; !ok {
			// initialise stack
			stacks[s] = &stack{}
			stacks[s].name = s
			stacks[s].setStackName()
			cfvars[s] = viper.Get(fmt.Sprintf("stacks.%s.cf", s))
		}

		key := "stacks." + s
		Log(fmt.Sprintf("Evaluating: [%s] in config"+"\n", s), level.debug)

		Log(fmt.Sprintf("Checking if [%s] exists in [%s]", key, strings.Join(viper.AllKeys(), ", ")), level.debug)
		if !strings.Contains(strings.Join(viper.AllKeys(), ""), key) {
			return fmt.Errorf("Key not found in config: %s", key)
		}

		Log(fmt.Sprintln("Processing: ", item), level.debug)

		switch item {

		// Handle depends_on
		case fmt.Sprintf("%s.%s", s, keyword.depends):
			// Only needs to be set on the first iteration
			dept := viper.Get(fmt.Sprintf("stacks.%s.%s", s, keyword.depends)).([]interface{})
			Log(fmt.Sprintf("Found Dependency for [%s]: %s", s, dept), level.debug)
			stacks[s].dependsOn = dept

		// Handle parameters
		case fmt.Sprintf("%s.%s", s, keyword.parameters):
			params := viper.Get(fmt.Sprintf("stacks.%s.%s", s, keyword.parameters)).([]interface{})
			if len(params) > 0 {
				for _, p := range params {
					for k, v := range p.(map[interface{}]interface{}) {
						stacks[s].parameters = append(stacks[s].parameters, &cloudformation.Parameter{
							ParameterKey:   aws.String(k.(string)),
							ParameterValue: aws.String(v.(string)),
						})
					}
				}
			}

		default:
			// TODO: no default yet..
		}

	}

	return nil
}

// genTimeParser - Parses templates before deploying them...
func genTimeParser(source string) (string, error) {

	templ, err := fetchContent(source)
	if err != nil {
		return "", err
	}

	// Create template
	t, err := template.New("template").Funcs(genTimeFunctions).Parse(templ)
	if err != nil {
		return "", err
	}

	// so that we can write to string
	var doc bytes.Buffer
	t.Execute(&doc, cfvars)
	return doc.String(), nil
}

// deployTimeParser - Parses templates during deployment to resolve specfic Dependency functions like stackout...
func (s *stack) deployTimeParser() error {

	// Create template
	t, err := template.New("template").Delims("<<", ">>").Funcs(deployTimeFunctions).Parse(s.template)
	if err != nil {
		return err
	}

	// so that we can write to string
	var doc bytes.Buffer
	t.Execute(&doc, cfvars)
	s.template = doc.String()
	Log(fmt.Sprintf("Deploy Time Template Generate:\n%s", s.template), level.debug)

	return nil
}
