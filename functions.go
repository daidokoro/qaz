package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
)

// Go Template Function Map here

var templateFunctions = template.FuncMap{
	// simple additon function useful for counters in loops
	"Add": func(a int, b int) int {
		Log(fmt.Sprintln("Calling Template Function [Add] with arguments:", a, b), level.debug)
		return a + b
	},

	// strip function for removing characters from text
	"Strip": func(s string, rmv string) string {
		Log(fmt.Sprintln("Calling Template Function [Strip] with arguments:", s, rmv), level.debug)
		return strings.Replace(s, rmv, "", -1)
	},

	// file function for reading text from a given file under the files folder
	"File": func(filename string) (string, error) {

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
	"S3Read": func(url string) (string, error) {
		Log(fmt.Sprintln("Calling Template Function [S3Read] with arguments:", url), level.debug)
		resp, err := S3Read(url)
		if err != nil {
			return "", err
		}
		return resp, nil
	},
}
