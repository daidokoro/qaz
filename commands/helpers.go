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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ttacon/chalk"
)

// configTemplate - Returns template byte string for init() function
func configTemplate(project string, region string) []byte {
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

// verbose - Takes a stackname and tracks the progress of creating the stack. s - stackname, c - command Type
func verbose(s string, c string, session *session.Session) {
	svc := cloudformation.New(session)

	params := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(s),
	}

	// used to track what lines have already been printed, to prevent dubplicate output
	printed := make(map[string]interface{})

	for {
		Log(fmt.Sprintf("Calling [DescribeStackEvents] with parameters: %s", params), level.debug)
		stackevents, err := svc.DescribeStackEvents(params)
		if err != nil {
			Log(fmt.Sprintln("Error when tailing events: ", err.Error()), level.debug)
		}

		Log(fmt.Sprintln("Response:", stackevents), level.debug)

		for _, event := range stackevents.StackEvents {

			statusReason := ""
			if strings.Contains(*event.ResourceStatus, "FAILED") {
				statusReason = *event.ResourceStatusReason
			}

			line := strings.Join([]string{
				colorMap(*event.ResourceStatus),
				*event.StackName,
				*event.ResourceType,
				*event.LogicalResourceId,
				statusReason,
			}, " - ")

			if _, ok := printed[line]; !ok {
				if strings.Split(*event.ResourceStatus, "_")[0] == c || c == "" {
					Log(strings.Trim(line, "- "), level.info)
				}

				printed[line] = nil
			}
		}

		// Sleep 2 seconds before next check
		time.Sleep(time.Duration(2 * time.Second))
	}
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
