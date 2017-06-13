package commands

import (
	"fmt"
	"io/ioutil"
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
		Log(err.Error(), level.err)
		return "", err
	}

	svc := kms.New(sess)

	params := &kms.EncryptInput{
		KeyId:     aws.String(kid),
		Plaintext: []byte(text),
	}

	resp, err := svc.Encrypt(params)
	if err != nil {
		Log(err.Error(), level.err)
		return "", err
	}

	return base64.StdEncoding.EncodeToString(resp.CiphertextBlob), nil
}

var kmsDecrypt = func(cipher string) (string, error) {
	sess, err := manager.GetSess(run.profile)
	if err != nil {
		Log(err.Error(), level.err)
		return "", err
	}

	svc := kms.New(sess)

	ciph, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		Log(err.Error(), level.err)
		return "", err
	}

	params := &kms.DecryptInput{
		CiphertextBlob: []byte(ciph),
	}

	resp, err := svc.Decrypt(params)
	if err != nil {
		Log(err.Error(), level.err)
		return "", err
	}

	return string(resp.Plaintext), nil
}

var httpGet = func(url string) (interface{}, error) {
	Log(fmt.Sprintln("Calling Template Function [GET] with arguments:", url), level.debug)
	resp, err := Get(url)
	if err != nil {
		Log(err.Error(), level.err)
		return "", err
	}

	return resp, nil
}

var s3Read = func(url string) (string, error) {
	Log(fmt.Sprintln("Calling Template Function [S3Read] with arguments:", url), level.debug)
	resp, err := S3Read(url)
	if err != nil {
		Log(err.Error(), level.err)
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
		Log(err.Error(), level.err)
		return "", err
	}

	if err := f.Invoke(sess); err != nil {
		Log(err.Error(), level.err)
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
		Log(fmt.Sprintln("Calling Template Function [add] with arguments:", a, b), level.debug)
		return a + b
	},

	// strip function for removing characters from text
	"strip": func(s string, rmv string) string {
		Log(fmt.Sprintln("Calling Template Function [strip] with arguments:", s, rmv), level.debug)
		return strings.Replace(s, rmv, "", -1)
	},

	// cat function for reading text from a given file under the files folder
	"cat": func(path string) (string, error) {

		Log(fmt.Sprintln("Calling Template Function [cat] with arguments:", path), level.debug)
		b, err := ioutil.ReadFile(path)
		if err != nil {
			Log(err.Error(), level.err)
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

	// kms-decrypt - Decrypts CipherText
	"kms_decrypt": kmsDecrypt,
}

var deployTimeFunctions = template.FuncMap{
	// Fetching stackoutputs
	"stack_output": func(target string) (string, error) {
		Log(fmt.Sprintf("Deploy-Time function resolving: %s", target), level.debug)
		req := strings.Split(target, "::")

		s := stacks[req[0]]

		if err := s.outputs(); err != nil {
			return "", err
		}

		for _, i := range s.output.Stacks {
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

		sess, err := manager.GetSess(run.profile)
		if err != nil {
			Log(err.Error(), level.err)
			return "", nil
		}

		s := stack{
			stackname: req[0],
			session:   sess,
		}

		if err := s.outputs(); err != nil {
			return "", err
		}

		for _, i := range s.output.Stacks {
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

	// kms-decrypt - Decrypts CipherText
	"kms_decrypt": kmsDecrypt,
}
