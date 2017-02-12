package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
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
		resp, err := Get(source)
		if err != nil {
			return "", err
		}
		return resp, nil
	case "https":
		resp, err := Get(source)
		if err != nil {
			return "", err
		}
		return resp, nil
	case "s3":
		resp, err := S3Read(source)
		if err != nil {
			return "", err
		}
		return resp, nil
	default:
		b, err := ioutil.ReadFile(source)
		if err != nil {
			log.Fatal("Error reading the template file: ", err)
		}
		return string(b), nil
	}
}

// configReader parses the config YAML file with Viper
func configReader(conf string) error {
	viper.AddConfigPath(".") // Required - for some reason

	cfg, err := fetchContent(conf)
	if err != nil {
		return err
	}

	err = viper.ReadConfig(bytes.NewBufferString(cfg))
	if err != nil {
		log.Fatal("Fatal error, can't read the config file")
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

	// if stacks is 0 at this point, All Stacks are assumed.
	if len(job.stacks) == 0 {
		job.stacks = make(map[string]string)
		for _, stk := range viper.Sub("stacks").AllKeys() {
			job.stacks[strings.Split(stk, ".")[0]] = ""
		}
	}

	for s := range job.stacks {
		stacks[s] = &stack{}

		for _, v := range viper.Sub("stacks." + s).AllKeys() {

			// Using case statement for setting keyword values, added for scalability later.
			switch v {
			case "depends_on":
				stacks[s].dependsOn = viper.Get(fmt.Sprintf("stacks.%s.%s", s, v)).([]interface{})
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

func templateParser(source string) (string, error) {

	templ, err := fetchContent(source)
	if err != nil {
		return "", err
	}

	// Create template
	t, err := template.New("template").Funcs(templateFunctions).Parse(templ)
	if err != nil {
		log.Fatal("Error parsing the template file: ", err)
	}

	// so that we can write to string
	var doc bytes.Buffer
	t.Execute(&doc, cfvars)
	return doc.String(), nil
}
