package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"text/template"
)

// Go Template Function Map here

var templateFunctions = template.FuncMap{
	// simple additon function useful for counters in loops
	"Add": func(a int, b int) int {
		return a + b
	},

	// strip function for removing characters from text
	"Strip": func(s string, rmv string) string {
		return strings.Replace(s, rmv, "", -1)
	},

	// file function for reading text from a given file under the files folder
	"File": func(filename string) (string, error) {

		p := job.tplFiles[0]
		f := filepath.Join(filepath.Dir(p), "..", "files", filename)
		fmt.Println(f)
		b, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal("Error reading the template file: ", err)
			return "", err
		}
		return string(b), nil
	},

	// Get get does an HTTP Get request of the given url and returns the output string
	"GET": func(url string) (string, error) {
		resp, err := Get(url)
		if err != nil {
			return "", err
		}

		return resp, nil
	},

	// S3Read reads content of file from s3 and returns string contents
	"S3Read": func(url string) (string, error) {
		resp, err := S3Read(url)
		if err != nil {
			return "", err
		}
		return resp, nil
	},
}
