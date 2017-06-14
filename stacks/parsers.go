package stacks

import (
	"bytes"
	"fmt"
	"text/template"
)

// DeployTimeParser - Parses templates during deployment to resolve specfic Dependency functions like stackout...
func (s *Stack) DeployTimeParser() error {

	// define Delims
	left, right := s.delims("deploy")

	// Create template
	t, err := template.New("deploy-template").Delims(left, right).Funcs(*s.DeployTimeFunc).Parse(s.Template)
	if err != nil {
		return err
	}

	// so that we can write to string
	var doc bytes.Buffer

	// Add metadata specific to the stack we're working with to the parser
	s.TemplateValues["stack"] = s.Name
	s.TemplateValues["parameters"] = s.Parameters

	t.Execute(&doc, s.TemplateValues)
	s.Template = doc.String()
	Log.Debug(fmt.Sprintf("Deploy Time Template Generate:\n%s", s.Template))

	return nil
}

// GenTimeParser - Parses templates before deploying them...
func (s *Stack) GenTimeParser() error {

	if err := s.FetchContent(); err != nil {
		return err
	}

	// define Delims
	left, right := s.delims("gen")

	// create template
	t, err := template.New("gen-template").Delims(left, right).Funcs(*s.GenTimeFunc).Parse(s.Template)
	if err != nil {
		return err
	}

	// so that we can write to string
	var doc bytes.Buffer

	// Add metadata specific to the stack we're working with to the parser
	s.TemplateValues["stack"] = s.Name
	s.TemplateValues["parameters"] = s.Parameters

	t.Execute(&doc, s.TemplateValues)
	s.Template = doc.String()
	return nil
}
