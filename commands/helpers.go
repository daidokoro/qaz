package commands

// -- Contains helper functions

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ttacon/chalk"
)

// configTemplate - Returns template byte string for init() function
func configTemplate(project string, region string) []byte {
	return []byte(fmt.Sprintf(`
# Specify the AWS region code
# qaz will attempt to get it from AWS configuration
# or from the environment. This setting overrides
# every other.

region: %s

# Required: specify the name of the Project
# (qaz will prepend this value to the stack
# names defined below.

project: %s

# Optional: global values accisible accross
# all stacks can be define under global

global:

# Stack-specific variables and
# arbitrary keys cab be defined here,
# Under the [stacks] key word,

stacks:
  # Note that the stack name must match the file name of the template file. The extension does not need to be specified.
  your_stack_name_here:
    cf:
      your_key/value_pairs_here:
`, region, project))
}

// colorMap - Used to map a particular color to a cf status phrase - returns lowercase strings in color.
func colorMap(s string) string {

	// If Windows, disable colorS
	if runtime.GOOS == "windows" {
		return s
	}

	v := strings.Split(s, "_")[len(strings.Split(s, "_"))-1]

	var result string

	switch v {
	case "COMPLETE":
		result = chalk.Green.Color(s)
	case "PROGRESS":
		result = chalk.Yellow.Color(s)
	case "FAILED":
		result = chalk.Red.Color(s)
	case "SKIPPED":
		result = chalk.Blue.Color(s)
	default:
		// Unidentified, just returns the same string
		return strings.ToLower(s)
	}
	return strings.ToLower(result)
}

// colorString - Returns colored string
func colorString(s string, color string) string {

	// If Windows, disable colorS
	if runtime.GOOS == "windows" {
		return s
	}

	var result string
	switch strings.ToLower(color) {
	case "green":
		result = chalk.Green.Color(s)
	case "yellow":
		result = chalk.Yellow.Color(s)
	case "red":
		result = chalk.Red.Color(s)
	case "magenta":
		result = chalk.Magenta.Color(s)
	default:
		// Unidentified, just returns the same string
		return s
	}

	return result
}

// all - returns true if all items in array the same as the given string
func all(a []string, s string) bool {
	for _, str := range a {
		if s != str {
			return false
		}
	}
	return true
}

// stringIn - returns true if string in array
func stringIn(s string, a []string) bool {
	Log(fmt.Sprintf("Checking If [%s] is in: %s", s, a), level.debug)
	for _, str := range a {
		if str == s {
			return true
		}
	}
	return false
}

// getInput - reads input from stdin - request & default (if no input)
func getInput(request string, def string) string {
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

// S3Read - Reads the content of a given s3 url endpoint and returns the content string.
func S3Read(url string) (string, error) {

	sess, err := awsSession()
	if err != nil {
		return "", err
	}

	svc := s3.New(sess)

	// Parse s3 url
	bucket := strings.Split(strings.Replace(strings.ToLower(url), `s3://`, "", -1), `/`)[0]
	key := strings.Replace(strings.ToLower(url), fmt.Sprintf("s3://%s/", bucket), "", -1)

	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	Log(fmt.Sprintln("Calling S3 [GetObject] with parameters:", params), level.debug)
	resp, err := svc.GetObject(params)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	Log("Reading from S3 Response Body", level.debug)
	buf.ReadFrom(resp.Body)
	return buf.String(), nil
}
