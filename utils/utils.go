package utils

// Helper functions

// -- Contains helper functions

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/daidokoro/hcl"
	"github.com/daidokoro/qaz/logger"
)

// Log defines logger
var Log *logger.Logger

// ConfigTemplate - Returns template byte string for init() function
func ConfigTemplate(project string, region string) []byte {
	return []byte(fmt.Sprintf(`
# AWS Region
region: %s

# Project Name
project: %s

# Global Stack Variables
global:

# Stacks
stacks:

`, region, project))
}

// All - returns true if all items in array the same as the given string
func All(a []string, s string) bool {
	for _, str := range a {
		if s != str {
			return false
		}
	}
	return true
}

// StringIn - returns true if string in array
func StringIn(s string, a []string) bool {
	Log.Debug(fmt.Sprintf("Checking If [%s] is in: %s", s, a))
	for _, str := range a {
		if str == s {
			return true
		}
	}
	return false
}

// GetInput - reads input from stdin - request & default (if no input)
func GetInput(request string, def string) string {
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%s]:", request, def)
	t, _ := r.ReadString('\n')

	// using len as t will always have atleast 1 char, "\n"
	if len(t) > 1 {
		return strings.Trim(t, "\n")
	}
	return def
}

// Get - HTTP Get request of given url and returns string
func Get(url string) (string, error) {
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)

	if resp == nil {
		return "", errors.New(fmt.Sprintln("Error, GET request timeout @:", url))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GET request failed, url: %s - Status:[%s]", url, strconv.Itoa(resp.StatusCode))
	}

	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// GetSource  - Checks if arg is url or file and returns stack name and filepath/url
func GetSource(src string) (string, string, error) {
	vals := strings.Split(src, "::")
	if len(vals) < 2 {
		return "", "", errors.New(`Error, invalid format - Usage: stackname::http://someurl OR stackname::path/to/template`)
	}

	return vals[0], vals[1], nil
}

// HandleError - exits on error
func HandleError(err error) {
	if err != nil {
		Log.Error(err.Error())
		os.Exit(1)
	}
}

// IsJSON - returns true if s is a string in valid json format
func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// IsHCL - returns true if s is a string in valid hcl format
func IsHCL(s string) bool {
	var h map[string]interface{}
	return hcl.Unmarshal([]byte(s), &h) == nil
}
