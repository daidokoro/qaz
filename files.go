package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

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

	// if stacks is 0 at this point, All Stacks are assumed.
	if len(job.stacks) == 0 {
		job.stacks = make(map[string]string)
		for _, stk := range viper.Sub("stacks").AllKeys() {
			job.stacks[strings.Split(stk, ".")[0]] = ""
		}
	}

	for s := range job.stacks {
		stacks[s] = &stack{}
		key := "stacks." + s
		Log(fmt.Sprintf("Evaluating: [%s] in config"+"\n", s), level.debug)

		Log(fmt.Sprintf("Checking if [%s] exists in [%s]", key, strings.Join(viper.AllKeys(), ", ")), level.debug)
		if !strings.Contains(strings.Join(viper.AllKeys(), ""), key) {
			return fmt.Errorf("Key not found in config: %s", key)
		}

		for _, v := range viper.Sub(key).AllKeys() {
			Log(fmt.Sprintln("Processing: ", v), level.debug)
			// Using case statement for setting keyword values, added for scalability later.
			switch v {
			case "depends_on":
				dept := viper.Get(fmt.Sprintf("stacks.%s.%s", s, v)).([]interface{})
				Log(fmt.Sprintf("Found Dependency for [%s]: %s", s, dept), level.debug)
				stacks[s].dependsOn = dept
			default:
				// TODO: Nothing for now - more built-in values to come... maybe
			}

			// Get Cloudformation values
			cfvars[s] = viper.Get(fmt.Sprintf("stacks.%s.cf", s))
		}

		stacks[s].name = s
		stacks[s].setStackName()
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
	t, err := template.New("template").Delims("%", "%").Funcs(deployTimeFunctions).Parse(s.template)
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
