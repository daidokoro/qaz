package stacks

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/daidokoro/qaz/bucket"
	"github.com/daidokoro/qaz/troposphere"
	"github.com/daidokoro/qaz/utils"
)

// Source - interface for template/config sources
type Source interface {
	// Handle - handle source
	Handle() (string, error)
}

// SourceReceiver - interface for all items that require multisource functionality
type SourceReceiver interface {
	// SourceReceiver uses GetSource to call source and
	// apply results
	GetSource(Source) error

	// SourceReceivers may require sessions if calling aws resources
	// like Lambda & S3
	GetSession() *session.Session
}

// GetSource - takes a Source interface and retrieves source data
func (s *Stack) GetSource(src Source) (err error) {
	raw, err := src.Handle()
	if err != nil {
		return
	}

	if s.Troposphere {
		s.Template, err = troposphere.Execute(raw)
		if err != nil {
			return
		}
		return
	}

	s.Template = raw
	return
}

// GetSession - Returns session to use in all source operations
func (s *Stack) GetSession() *session.Session {
	return s.Session
}

// GetSource - takes a Source interface and retrieves source data
func (c *Config) GetSource(src Source) (err error) {
	c.String, err = src.Handle()
	if err != nil {
		return
	}
	return
}

// GetSession - Returns session to use in all source operations
func (c *Config) GetSession() *session.Session {
	return c.Session
}

// FetchSource - uses interfaces to initiate source retreival
func FetchSource(src string, rcv SourceReceiver) error {
	var source Source

	// get source type from uri Scheme
	uri, err := url.Parse(src)
	if err != nil {
		return err
	}

	switch uri.Scheme {
	case "http", "https":
		source = &HTTPSource{src}
	case "lambda":
		source = &LambdaSource{src, rcv.GetSession()}
	case "s3":
		source = &S3Source{src, rcv.GetSession()}
	default:
		source = &FileSource{src}
	}

	return rcv.GetSource(source)
}

// Source TYpes

// HTTPSource - Http source type
type HTTPSource struct {
	Src string
}

// Handle - Source Handle
func (h HTTPSource) Handle() (string, error) {
	log.Debug("Source Type: [http] Detected, Fetching Source: %s ", h.Src)
	return utils.Get(h.Src)
}

// S3Source - S3 source type
type S3Source struct {
	Src     string
	Session *session.Session
}

// Handle - Source Handle
func (s3 S3Source) Handle() (string, error) {
	log.Debug("Source Type: [s3] Detected, Fetching Source: %s", s3.Src)
	return bucket.S3Read(s3.Src, s3.Session)
}

// LambdaSource - lambda source handle
type LambdaSource struct {
	Src     string
	Session *session.Session
}

// Handle - Source Handle
func (l LambdaSource) Handle() (string, error) {
	log.Debug("Source Type: [lambda] Detected, Fetching Source: %s", l.Src)
	src, err := url.Parse(l.Src)
	if err != nil {
		return "", err
	}

	lambdaSrc := strings.Split(src.Opaque, "@")

	var raw interface{}
	if err = json.Unmarshal([]byte(lambdaSrc[0]), &raw); err != nil {
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

	f := awslambda{
		name:    lambdaName,
		payload: event,
	}

	if err := f.Invoke(l.Session); err != nil {
		return "", err
	}

	return f.response, nil
}

// FileSource - interface type
// NOTE: file is assumed if no other type if matched
type FileSource struct {
	Src string
}

// Handle - Source Handle
func (f FileSource) Handle() (resp string, err error) {
	if Git.URL != "" {
		log.Debug("Source Type: [git-repo file] Detected, Fetching Source: %s", f.Src)
		out, ok := Git.Files[f.Src]
		if ok {
			resp = out
			return
		} else if !ok {
			log.Warn("config [%s] not found in git repo - checking local file system", f.Src)
		}
	}

	log.Debug("Source Type: [file] Detected, Fetching Source: %s", f.Src)
	b, err := ioutil.ReadFile(f.Src)
	if err != nil {
		return
	}
	resp = string(b)
	return
}
