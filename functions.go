package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
)

// Go Template Function Map here

var genTimeFunctions = template.FuncMap{
	// simple additon function useful for counters in loops
	"add": func(a int, b int) int {
		Log(fmt.Sprintln("Calling Template Function [Add] with arguments:", a, b), level.debug)
		return a + b
	},

	// strip function for removing characters from text
	"strip": func(s string, rmv string) string {
		Log(fmt.Sprintln("Calling Template Function [Strip] with arguments:", s, rmv), level.debug)
		return strings.Replace(s, rmv, "", -1)
	},

	// file function for reading text from a given file under the files folder
	"file": func(filename string) (string, error) {

		Log(fmt.Sprintln("Calling Template Function [File] with arguments:", filename), level.debug)
		p := job.tplFiles[0]
		f := filepath.Join(filepath.Dir(p), "..", "files", filename)
		fmt.Println(f)
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},

	// Get get does an HTTP Get request of the given url and returns the output string
	"GET": func(url string) (string, error) {
		Log(fmt.Sprintln("Calling Template Function [GET] with arguments:", url), level.debug)
		resp, err := Get(url)
		if err != nil {
			return "", err
		}

		return resp, nil
	},

	// S3Read reads content of file from s3 and returns string contents
	"s3_read": func(url string) (string, error) {
		Log(fmt.Sprintln("Calling Template Function [S3Read] with arguments:", url), level.debug)
		resp, err := S3Read(url)
		if err != nil {
			return "", err
		}
		return resp, nil
	},

	// invoke - invokes a lambda function
	"invoke": func(name string, payload string) (string, error) {
		f := function{name: name}
		if payload != "" {
			f.payload = []byte(payload)
		}

		sess, err := awsSession()
		if err != nil {
			return "", err
		}

		if err := f.Invoke(sess); err != nil {
			return "", err
		}

		return f.response, nil
	},
}

var deployTimeFunctions = template.FuncMap{
	// Fetching stackoutputs
	"stack_output": func(target string) (string, error) {
		Log(fmt.Sprintf("Deploy-Time function resolving: %s", target), level.debug)
		req := strings.Split(target, "::")
		sess, err := awsSession()
		if err != nil {
			return "", nil
		}

		stkname := strings.Join([]string{project, req[0]}, "-")
		outputs, err := StackOutputs(stkname, sess)

		if err != nil {
			return "", err
		}

		for _, i := range outputs.Stacks {
			for _, o := range i.Outputs {
				if *o.OutputKey == req[1] {
					return *o.OutputValue, nil
				}
			}
		}

		return "", fmt.Errorf("Stack Output Not found - Stack:%s | Output:%s", req[0], req[1])
	},

	"stack_output_ext": func(target string) (string, error) {
		Log(fmt.Sprintf("Deploy-Time function resolving: %s", target), level.debug)
		req := strings.Split(target, "::")
		sess, err := awsSession()
		if err != nil {
			return "", nil
		}

		outputs, err := StackOutputs(req[0], sess)

		if err != nil {
			return "", err
		}

		for _, i := range outputs.Stacks {
			for _, o := range i.Outputs {
				if *o.OutputKey == req[1] {
					return *o.OutputValue, nil
				}
			}
		}

		return "", fmt.Errorf("Stack Output Not found - Stack:%s | Output:%s", req[0], req[1])
	},

	// Get get does an HTTP Get request of the given url and returns the output string
	"GET": func(url string) (string, error) {
		Log(fmt.Sprintln("Calling Template Function [GET] with arguments:", url), level.debug)
		resp, err := Get(url)
		if err != nil {
			return "", err
		}

		return resp, nil
	},

	// S3Read reads content of file from s3 and returns string contents
	"s3_read": func(url string) (string, error) {
		Log(fmt.Sprintln("Calling Template Function [S3Read] with arguments:", url), level.debug)
		resp, err := S3Read(url)
		if err != nil {
			return "", err
		}
		return resp, nil
	},

	// invoke - invokes a lambda function
	"invoke": func(name string, payload string) (string, error) {
		f := function{name: name}
		if payload != "" {
			f.payload = []byte(payload)
		}

		sess, err := awsSession()
		if err != nil {
			return "", err
		}

		if err := f.Invoke(sess); err != nil {
			return "", err
		}

		return f.response, nil
	},
}
