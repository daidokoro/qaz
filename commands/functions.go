package commands

import (
	"fmt"
	"io/ioutil"
	"qaz/bucket"
	stks "qaz/stacks"
	"qaz/utils"
	"strings"
	"text/template"

	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

// Common Functions - Both Deploy/Gen

var kmsEncrypt = func(kid string, text string) (string, error) {
	sess, err := manager.GetSess(run.profile)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	svc := kms.New(sess)

	params := &kms.EncryptInput{
		KeyId:     aws.String(kid),
		Plaintext: []byte(text),
	}

	resp, err := svc.Encrypt(params)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	return base64.StdEncoding.EncodeToString(resp.CiphertextBlob), nil
}

var kmsDecrypt = func(cipher string) (string, error) {
	sess, err := manager.GetSess(run.profile)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	svc := kms.New(sess)

	ciph, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	params := &kms.DecryptInput{
		CiphertextBlob: []byte(ciph),
	}

	resp, err := svc.Decrypt(params)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	return string(resp.Plaintext), nil
}

var httpGet = func(url string) (interface{}, error) {
	log.Debug(fmt.Sprintln("Calling Template Function [GET] with arguments:", url))
	resp, err := utils.Get(url)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	return resp, nil
}

var s3Read = func(url string, profile ...string) (string, error) {
	log.Debug(fmt.Sprintln("Calling Template Function [S3Read] with arguments:", url))

	var p = run.profile
	if len(profile) < 1 {
		log.Warn(fmt.Sprintf("No Profile specified for S3read, using: %s", p))
	} else {
		p = profile[0]
	}

	sess, err := manager.GetSess(p)
	utils.HandleError(err)

	resp, err := bucket.S3Read(url, sess)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	return resp, nil
}

var lambdaInvoke = func(name string, payload string) (interface{}, error) {
	f := awsLambda{name: name}
	if payload != "" {
		f.payload = []byte(payload)
	}

	sess, err := manager.GetSess(run.profile)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	if err := f.Invoke(sess); err != nil {
		log.Error(err.Error())
		return "", err
	}

	return f.response, nil
}

var prefix = func(s string, pre string) bool {
	return strings.HasPrefix(s, pre)
}

var suffix = func(s string, suf string) bool {
	return strings.HasSuffix(s, suf)
}

var contains = func(s string, con string) bool {
	return strings.Contains(s, con)
}

// template function maps

var genTimeFunctions = template.FuncMap{
	// simple additon function useful for counters in loops
	"add": func(a int, b int) int {
		log.Debug(fmt.Sprintln("Calling Template Function [add] with arguments:", a, b))
		return a + b
	},

	// strip function for removing characters from text
	"strip": func(s string, rmv string) string {
		log.Debug(fmt.Sprintln("Calling Template Function [strip] with arguments:", s, rmv))
		return strings.Replace(s, rmv, "", -1)
	},

	// cat function for reading text from a given file under the files folder
	"cat": func(path string) (string, error) {

		log.Debug(fmt.Sprintln("Calling Template Function [cat] with arguments:", path))
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Error(err.Error())
			return "", err
		}
		return string(b), nil
	},

	// suffix - returns true if string starts with given suffix
	"suffix": suffix,

	// prefix - returns true if string starts with given prefix
	"prefix": prefix,

	// contains - returns true if string contains
	"contains": contains,

	// Get get does an HTTP Get request of the given url and returns the output string
	"GET": httpGet,

	// S3Read reads content of file from s3 and returns string contents
	"s3_read": s3Read,

	// invoke - invokes a lambda function
	"invoke": lambdaInvoke,

	// kms-encrypt - Encrypts PlainText using KMS key
	"kms_encrypt": kmsEncrypt,

	// kms-decrypt - Descrypts CipherText
	"kms_decrypt": kmsDecrypt,
}

var deployTimeFunctions = template.FuncMap{
	// Fetching stackoutputs
	"stack_output": func(target string) (string, error) {
		log.Debug(fmt.Sprintf("Deploy-Time function resolving: %s", target))
		req := strings.Split(target, "::")

		s := stacks[req[0]]

		if err := s.Outputs(); err != nil {
			return "", err
		}

		for _, i := range s.Output.Stacks {
			for _, o := range i.Outputs {
				if *o.OutputKey == req[1] {
					return *o.OutputValue, nil
				}
			}
		}

		return "", fmt.Errorf("Stack Output Not found - Stack:%s | Output:%s", req[0], req[1])
	},

	"stack_output_ext": func(target string) (string, error) {
		log.Debug(fmt.Sprintf("Deploy-Time function resolving: %s", target))
		req := strings.Split(target, "::")

		sess, err := manager.GetSess(run.profile)
		if err != nil {
			log.Error(err.Error())
			return "", nil
		}

		s := stks.Stack{
			Stackname: req[0],
			Session:   sess,
		}

		if err := s.Outputs(); err != nil {
			return "", err
		}

		for _, i := range s.Output.Stacks {
			for _, o := range i.Outputs {
				if *o.OutputKey == req[1] {
					return *o.OutputValue, nil
				}
			}
		}

		return "", fmt.Errorf("Stack Output Not found - Stack:%s | Output:%s", req[0], req[1])
	},

	// suffix - returns true if string starts with given suffix
	"suffix": suffix,

	// prefix - returns true if string starts with given prefix
	"prefix": prefix,

	// contains - returns true if string contains
	"contains": contains,

	// Get get does an HTTP Get request of the given url and returns the output string
	"GET": httpGet,

	// S3Read reads content of file from s3 and returns string contents
	"s3_read": s3Read,

	// invoke - invokes a lambda function
	"invoke": lambdaInvoke,

	// kms-encrypt - Encrypts PlainText using KMS key
	"kms_encrypt": kmsEncrypt,

	// kms-decrypt - Descrypts CipherText
	"kms_decrypt": kmsDecrypt,
}
