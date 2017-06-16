package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/daidokoro/qaz/bucket"
	"github.com/daidokoro/qaz/utils"
	"regexp"
	"strings"
)

// TODO: Come up with a better way to do this
// fetchContent - checks the source type, url/s3/file and calls the corresponding function
func fetchContent(source string) (string, error) {
	switch strings.Split(strings.ToLower(source), ":")[0] {
	case "http", "https":
		log.Debug(fmt.Sprintln("Source Type: [http] Detected, Fetching Source: ", source))
		resp, err := utils.Get(source)
		if err != nil {
			return "", err
		}
		return resp, nil
	case "s3":
		log.Debug(fmt.Sprintln("Source Type: [s3] Detected, Fetching Source: ", source))
		sess, err := manager.GetSess(run.profile)
		utils.HandleError(err)

		resp, err := bucket.S3Read(source, sess)
		if err != nil {
			return "", err
		}
		return resp, nil
	case "lambda":
		log.Debug(fmt.Sprintln("Source Type: [lambda] Detected, Fetching Source: ", source))
		lambdaSrc := strings.Split(strings.Replace(source, "lambda:", "", -1), "@")

		var raw interface{}
		if err := json.Unmarshal([]byte(lambdaSrc[0]), &raw); err != nil {
			return "", err
		}

		event, err := json.Marshal(raw)
		if err != nil {
			return "", err
		}

		reg, err := regexp.Compile("[^A-Za-z0-9_-]+")
		if err != nil {
			return "", err
		}

		lambdaName := reg.ReplaceAllString(lambdaSrc[1], "")

		f := awsLambda{
			name:    lambdaName,
			payload: event,
		}

		// using default profile
		sess := manager.sessions[run.profile]
		if err := f.Invoke(sess); err != nil {
			return "", err
		}

		return f.response, nil

	default:
		if gitrepo.URL != "" {
			log.Debug(fmt.Sprintln("Source Type: [git-repo file] Detected, Fetching Source: ", source))
			out, ok := gitrepo.Files[source]
			if ok {
				return out, nil
			} else if !ok {
				log.Warn(fmt.Sprintf("config [%s] not found in git repo - checking local file system", source))
			}

		}

		log.Debug(fmt.Sprintln("Source Type: [file] Detected, Fetching Source: ", source))
		b, err := ioutil.ReadFile(source)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}
