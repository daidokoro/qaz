package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/daidokoro/qaz/bucket"
	"github.com/daidokoro/qaz/utils"
	"github.com/spf13/cobra"

	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/daidokoro/qaz/log"
)

// Common Functions - Both Deploy/Gen
var (
	kmsEncrypt = func(kid string, text string) string {
		log.Debug("running template function: [kms_encrypt]")
		sess, err := GetSession()
		utils.HandleError(err)

		svc := kms.New(sess)

		params := &kms.EncryptInput{
			KeyId:     aws.String(kid),
			Plaintext: []byte(text),
		}

		resp, err := svc.Encrypt(params)
		utils.HandleError(err)

		return base64.StdEncoding.EncodeToString(resp.CiphertextBlob)
	}

	kmsDecrypt = func(cipher string) string {
		log.Debug("running template function: [kms_decrypt]")
		sess, err := GetSession()
		utils.HandleError(err)

		svc := kms.New(sess)

		ciph, err := base64.StdEncoding.DecodeString(cipher)
		utils.HandleError(err)

		params := &kms.DecryptInput{
			CiphertextBlob: []byte(ciph),
		}

		resp, err := svc.Decrypt(params)
		utils.HandleError(err)

		return string(resp.Plaintext)
	}

	httpGet = func(url string) interface{} {
		log.Debug("Calling Template Function [GET] with arguments: %s", url)
		resp, err := utils.Get(url)
		utils.HandleError(err)

		return resp
	}

	pkgfunc = func(path, s3URI string, profile ...string) interface{} {
		u, err := url.Parse(s3URI)
		utils.HandleError(err)

		m := make(map[string]string, 2)
		m["S3Key"] = u.Path[1:]
		m["S3Bucket"] = u.Host

		if !run.executePackage {
			log.Warn("[package] template function detected but package not used")
			log.Warn("please use --package/-P flags to execute package functions")
			return m
		}

		log.Debug("Calling Template Function [package] with arguments: %s:")
		sess, err := GetSession(func(opts *session.Options) {
			if len(profile) < 1 {
				log.Warn("No Profile specified for [package], using: %s", run.profile)
				return
			}
			opts.Profile = profile[0]
			return
		})
		utils.HandleError(err)

		log.Info("creating ZIP package from : [%s]", path)
		buf, err := utils.Zip(path)
		utils.HandleError(err)

		log.Info("uploading package to s3: [%s]", s3URI)
		_, err = bucket.S3write(u.Host, u.Path, bytes.NewReader(buf.Bytes()), sess)
		utils.HandleError(err)

		return m
	}

	s3Read = func(url string, profile ...string) string {
		log.Debug("Calling Template Function [S3Read] with arguments: %s", url)

		sess, err := GetSession(func(opts *session.Options) {
			if len(profile) < 1 {
				log.Warn("No Profile specified for S3read, using: %s", run.profile)
				return
			}
			opts.Profile = profile[0]
			return
		})
		utils.HandleError(err)

		resp, err := bucket.S3Read(url, sess)
		utils.HandleError(err)

		return resp
	}

	lambdaInvoke = func(name string, payload string) interface{} {
		log.Debug("running template function: [invoke]")
		f := awsLambda{name: name}
		var m interface{}

		if payload != "" {
			f.payload = []byte(payload)
		}

		sess, err := GetSession()
		utils.HandleError(err)

		err = f.Invoke(sess)
		utils.HandleError(err)

		log.Debug("Lambda response: %s", f.response)

		// parse json if possible
		if err := json.Unmarshal([]byte(f.response), &m); err != nil {
			log.Debug(err.Error())
			return f.response
		}

		return m
	}

	prefix = func(s string, pre string) bool {
		return strings.HasPrefix(s, pre)
	}

	suffix = func(s string, suf string) bool {
		return strings.HasSuffix(s, suf)
	}

	contains = func(s string, con string) bool {
		return strings.Contains(s, con)
	}

	loop = func(n int) []struct{} {
		return make([]struct{}, n)
	}

	literal = func(str string) string {
		return fmt.Sprintf("%#v", str)
	}

	// gentime function maps
	GenTimeFunctions = template.FuncMap{
		// simple additon function useful for counters in loops
		"add": func(a int, b int) int {
			log.Debug("Calling Template Function [add] with arguments: %s + %s", a, b)
			return a + b
		},

		// strip function for removing characters from text
		"strip": func(s string, rmv string) string {
			log.Debug("Calling Template Function [strip] with arguments: (%s, %s) ", s, rmv)
			return strings.Replace(s, rmv, "", -1)
		},

		// cat function for reading text from a given file under the files folder
		"cat": func(path string) string {
			log.Debug("Calling Template Function [cat] with arguments: %s", path)
			b, err := ioutil.ReadFile(path)
			utils.HandleError(err)
			return string(b)
		},

		// literal - return raw/literal string with special chars and all
		"literal": literal,

		// suffix - returns true if string starts with given suffix
		"suffix": suffix,

		// prefix - returns true if string starts with given prefix
		"prefix": prefix,

		// contains - returns true if string contains
		"contains": contains,

		// loop - useful to range over an int (rather than a slice, map, or channel). see examples/loop
		"loop": loop,

		// Get get does an HTTP Get request of the given url and returns the output string
		"GET": httpGet,

		// S3Read reads content of file from s3 and returns string contents
		"s3_read": s3Read,

		// invoke - invokes a lambda function and returns a raw string/interface{}
		"invoke": lambdaInvoke,

		// kms-encrypt - Encrypts PlainText using KMS key
		"kms_encrypt": kmsEncrypt,

		// kms-decrypt - Descrypts CipherText
		"kms_decrypt": kmsDecrypt,

		// mod - returns remainder after division (modulus)
		"mod": func(a int, b int) int {
			log.Debug("Calling Template Function [mod] with arguments: %s % %s", a, b)
			return a % b
		},

		// seq - returns a sequence of numbers
		"seq": func(from, to int) []int {
			log.Debug("Calling Template Function [seq] with arguments: %s % %s", from, to)
			seq := make([]int, to-from+1)
			for i := range seq {
				seq[i] = from + i
			}
			return seq
		},

		// capitalize first letter
		"title": func(s string) string {
			log.Debug("Calling Template Function [title] with arguments: %s", s)
			return strings.Title(s)
		},

		// package function - create  zip package from local path
		"package": pkgfunc,
	}

	// deploytime function maps
	DeployTimeFunctions = template.FuncMap{

		// suffix - returns true if string starts with given suffix
		"suffix": suffix,

		// prefix - returns true if string starts with given prefix
		"prefix": prefix,

		// contains - returns true if string contains
		"contains": contains,

		// loop - useful to range over an int (rather than a slice, map, or channel). see examples/loop
		"loop": loop,

		// literal - return raw/literal string with special chars and all
		"literal": literal,

		// Get get does an HTTP Get request of the given url and returns the output string
		"GET": httpGet,

		// S3Read reads content of file from s3 and returns string contents
		"s3_read": s3Read,

		// invoke - invokes a lambda function and returns a raw string/interface{}
		"invoke": lambdaInvoke,

		// kms-encrypt - Encrypts PlainText using KMS key
		"kms_encrypt": kmsEncrypt,

		// kms-decrypt - Descrypts CipherText
		"kms_decrypt": kmsDecrypt,
	}
)

var templateFunctionDoc = `
--------------------------------
---- Qaz Template Functions ----
--------------------------------

Custom Template Functions expand the functionality of Go's Templating library by allowing you to execute external functions to retrieve additional information for building your template.

Qaz supports all the Go Template functions as well as some custom ones.

Qaz has two levels of custom template functions, these are Gen-Time functions and Deploy-Time functions.

 - Gen-Time functions are functions that are executed when a template is being generated. These are handy for reading files into a template or making API calls to fetch values.
   Gen-Time functions are delimited by {{` + "`{{ }}`" + `}}

 - Deploy-Time functions are run just before the template is pushed to AWS Cloudformation. These are handy for:

	* Fetching values from dependency stacks
	* Making API calls to pull values from resources built by preceding stacks
	* Triggering events via an API call and adjusting the template based on the response
	* Updating Values in a decrypted template
   
   Deploy-Time functions are delimted by << >>

--

{{- range $_, $f := . }}

%s:

	{{ $f.Name }}

%s:
	
	{{ $f.Desc }}
	{{ $f.Type }}

%s:

	{{ $f.Usage }}

--
{{ end }}

This list contains only custom functions added to Qaz, however, all built-in functionality for Go Templates is supported.
See here for details: https://golang.org/pkg/text/template/
`

// TemplateFunctionDesc -describes template functions
type TemplateFunctionDesc struct {
	Name  string
	Desc  string
	Type  string
	Usage string
}

var templateFunctionsCmd = &cobra.Command{
	Use: "template-functions",
	Example: strings.Join([]string{
		"qaz template-functions [function-name]",
		"qaz template-functions",
	}, "\n"),
	Short:  "prints a list with descriptions and examples of all available custom template functions",
	PreRun: initialise,
	Run: func(cmd *cobra.Command, args []string) {

		// simplifying log.ColorString to f
		// used to color function string
		f := func(s string) string {
			return log.ColorString(s, log.CYAN)
		}

		const (
			gd = "(gen-time|deploy-time)"
			g  = "(gen-time only)"
			d  = "(deploy-time only)"
		)

		data := []*TemplateFunctionDesc{
			&TemplateFunctionDesc{
				f("add"),
				"Simple additon function useful for counters in loops", g,
				`{{ add 1 2 }} --> 3`,
			},

			&TemplateFunctionDesc{
				f("strip"),
				"Removes a given substring from text", g,
				`{{ strip "cat" "c" }} --> at`,
			},

			&TemplateFunctionDesc{
				f("cat"),
				"Reads text from a given filepath to template", g,
				`{{ cat "/path/to/some/file.txt" }}`,
			},

			&TemplateFunctionDesc{
				f("literal"),
				"Prints literal unevaluated version of a given string", g,
				`{{ literal "cat\nhouse" }}`,
			},

			&TemplateFunctionDesc{
				f("suffix"),
				"Returns true if the string ends with given suffix", gd,
				`{{ if suffix "something" "thing" }} value {{ end }} --> value`,
			},

			&TemplateFunctionDesc{
				f("prefix"),
				"Returns true if the string starts with given suffix", gd,
				`{{ if prefix "something" "some" }} value {{ end }} --> value`,
			},

			&TemplateFunctionDesc{
				f("contains"),
				"Returns true if the string contains the given sub-string", gd,
				`{{ if contains "something" "some" }} value {{ end }} --> value`,
			},

			&TemplateFunctionDesc{
				f("loop"),
				"Takes an int n and iterates n times", gd,
				`{{ range $i, $_ := loop 5 }} value {{ end }} --> value  value  value  value  value`,
			},

			&TemplateFunctionDesc{
				f("seq"),
				"Returns an iteratable sequence from x to y", g,
				`{{ $range $_, $v := seq 1 5 }} {{ $v }} {{ end }}  --> `,
			},

			&TemplateFunctionDesc{
				f("GET"),
				"HTTP GET reqest to a given url. Response is then written to the template", gd,
				`{{ GET "https://some.endpoint.app" }} --> {"some":"response"} or "some string response"`,
			},

			&TemplateFunctionDesc{
				f("s3read"),
				"Read s3 object and writes the contents to the template", gd,
				`{{ s3read "s3://bucket/containing/things" }} --> "things"`,
			},

			&TemplateFunctionDesc{
				f("invoke"),
				"Invokes a Lambda function and writes the returned value to the template.", gd,
				"{{ invoke \"function_name\" `{\"some_json\":\"some_value\"}` }}",
			},

			&TemplateFunctionDesc{
				f("kms_encrypt"),
				"Generates an encrypted Cipher Text blob using AWS KMS", gd,
				`{{ kms_encrypt kms.keyid "Text to Encrypt!" }} --> "CipherText"`,
			},

			&TemplateFunctionDesc{
				f("kms_decrypt"),
				"Decrypts a given Cipher Text blob using AWS KMS", gd,
				`{{ kms_decrypt "CipherTextBlob" }} --> "Decrypted CipherText"`,
			},

			&TemplateFunctionDesc{
				f("mod"),
				"Modulus Division within templates. I.e Returns the remainder of an uneven division", g,
				`{{ mod 7 3 }} --> 1`,
			},

			&TemplateFunctionDesc{
				f("title"),
				"Returns a copy of the string s with all Unicode letters that begin words mapped to their title case", g,
				`{{ title "tengen toppa gurren lagan" }} --> Tengen Toppa Gurren Lagan`,
			},

			&TemplateFunctionDesc{
				f("package"),
				"The package function creates a ZIP file from a given path and uploads it to S3. This functionality is useful for AWS Lambda & Layer deployments", g,
				`{{ package "./code" "s3://mybucket/path/to/function.zip" }} --> s3://mybucket/path/to/function.zip`,
			},

			&TemplateFunctionDesc{
				f("stack_output"),
				"Fetches the output value of a given stack and stores the value in your template. This function uses the stack name as defined in your project configuration", d,
				`<< stack_output "vpc::vpcid" >>`,
			},

			&TemplateFunctionDesc{
				f("stack_output_ext"),
				"Fetches the output value of a given stack that exists outside of your project/configuration and stores the value in your template. This function requires the full name of the stack as it appears on the AWS Console.", d,
				`<< stack_output_ext "external-vpc::vpcid" >>`,
			},
		}

		// if specific function is requested
		var singleFunctionData []*TemplateFunctionDesc
		if len(args) > 0 {
			fn := args[0]
			for _, tf := range data {
				if tf.Name == f(fn) {
					singleFunctionData = append(singleFunctionData, tf)
				}
			}

			if len(singleFunctionData) > 0 {
				data = singleFunctionData
			}
		}

		doc := fmt.Sprintf(templateFunctionDoc, log.ColorString("Function", log.YELLOW),
			log.ColorString("Description", log.YELLOW),
			log.ColorString("Usage", log.YELLOW),
		)
		tmpl, err := template.New("function doc").Parse(doc)
		utils.HandleError(err)

		var t bytes.Buffer
		utils.HandleError(tmpl.Execute(&t, data))

		// formatting delimeters
		doc = regexp.MustCompile(`{{|}}|<<|>>`).
			ReplaceAllStringFunc(t.String(), func(s string) string {
				return log.ColorString(s, log.RED)
			})

		pager := exec.Command(os.Getenv("PAGER"))
		pager.Stdin = strings.NewReader(doc)
		pager.Stdout = os.Stdout
		utils.HandleError(pager.Run())
	},
}
