package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// stack - holds all meaningful information about a particular stack.
type stack struct {
	name         string
	stackname    string
	template     string
	dependsOn    []string
	dependents   []interface{}
	stackoutputs *cloudformation.DescribeStacksOutput
	parameters   []*cloudformation.Parameter
	output       *cloudformation.DescribeStacksOutput
	policy       string
	session      *session.Session
	profile      string
	source       string
	bucket       string
	role         string
}

// setStackName - sets the stackname with struct
func (s *stack) setStackName() {
	s.stackname = fmt.Sprintf("%s-%s", config.Project, s.name)
}

// creds - Returns credentials if role set
func (s *stack) creds() *credentials.Credentials {
	var creds *credentials.Credentials
	if s.role == "" {
		return creds
	}
	return stscreds.NewCredentials(s.session, s.role)
}

func (s *stack) deploy() error {

	err := s.deployTimeParser()
	if err != nil {
		return err
	}

	Log(fmt.Sprintf("Updated Template:\n%s", s.template), level.debug)
	done := make(chan bool)
	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})

	createParams := &cloudformation.CreateStackInput{
		StackName:       aws.String(s.stackname),
		DisableRollback: aws.Bool(!job.rollback),
	}

	if s.policy != "" {
		if strings.HasPrefix(s.policy, "http://") || strings.HasPrefix(s.policy, "https://") {
			createParams.StackPolicyURL = &s.policy
		} else {
			createParams.StackPolicyBody = &s.policy
		}
	}

	// NOTE: Add parameters flag here if params set
	if len(s.parameters) > 0 {
		createParams.Parameters = s.parameters
	}

	// If IAM is being touched, add Capabilities
	if strings.Contains(s.template, "AWS::IAM") {
		createParams.Capabilities = []*string{
			aws.String(cloudformation.CapabilityCapabilityIam),
			aws.String(cloudformation.CapabilityCapabilityNamedIam),
		}
	}

	// If bucket - upload to s3
	if s.bucket != "" {
		exists, err := BucketExists(s.bucket, s.session)
		if err != nil {
			Log(fmt.Sprintf("Received Error when checking if [%s] exists: %s", s.bucket, err.Error()), level.warn)
		}

		if !exists {
			Log(fmt.Sprintf(("Creating Bucket [%s]"), s.bucket), level.info)
			if err = CreateBucket(s.bucket, s.session); err != nil {
				return err
			}
		}
		t := time.Now()
		tStamp := fmt.Sprintf("%d-%d-%d_%d%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
		url, err := S3write(s.bucket, fmt.Sprintf("%s_%s.template", s.stackname, tStamp), s.template, s.session)
		if err != nil {
			return err
		}
		createParams.TemplateURL = &url
	} else {
		createParams.TemplateBody = &s.template
	}

	Log(fmt.Sprintln("Calling [CreateStack] with parameters:", createParams), level.debug)
	if _, err := svc.CreateStack(createParams); err != nil {
		return errors.New(fmt.Sprintln("Deploying failed: ", err.Error()))

	}

	go s.tail("CREATE", done)
	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [WaitUntilStackCreateComplete] with parameters:", describeStacksInput), level.debug)
	if err := svc.WaitUntilStackCreateComplete(describeStacksInput); err != nil {
		return err
	}

	Log(fmt.Sprintf("Deployment successful: [%s]", s.stackname), "info")

	done <- true
	return nil
}

func (s *stack) update() error {

	err := s.deployTimeParser()
	if err != nil {
		return err
	}

	done := make(chan bool)
	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})
	updateParams := &cloudformation.UpdateStackInput{
		StackName:    aws.String(s.stackname),
		TemplateBody: aws.String(s.template),
	}

	// If bucket - upload to s3
	if s.bucket != "" {
		exists, err := BucketExists(s.bucket, s.session)
		if err != nil {
			Log(fmt.Sprintf("Received Error when checking if [%s] exists: %s", s.bucket, err.Error()), level.warn)
		}

		if !exists {
			Log(fmt.Sprintf(("Creating Bucket [%s]"), s.bucket), level.info)
			if err = CreateBucket(s.bucket, s.session); err != nil {
				return err
			}
		}
		t := time.Now()
		tStamp := fmt.Sprintf("%d-%d-%d_%d%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
		url, err := S3write(s.bucket, fmt.Sprintf("%s_%s.template", s.stackname, tStamp), s.template, s.session)
		if err != nil {
			return err
		}
		updateParams.TemplateURL = &url
	} else {
		updateParams.TemplateBody = &s.template
	}

	// NOTE: Add parameters flag here if params set
	if len(s.parameters) > 0 {
		updateParams.Parameters = s.parameters
	}

	// If IAM is being touched, add Capabilities
	if strings.Contains(s.template, "AWS::IAM") {
		updateParams.Capabilities = []*string{
			aws.String(cloudformation.CapabilityCapabilityIam),
			aws.String(cloudformation.CapabilityCapabilityNamedIam),
		}
	}

	if s.stackExists() {
		Log("Stack exists, updating...", "info")

		Log(fmt.Sprintln("Calling [UpdateStack] with parameters:", updateParams), level.debug)
		_, err := svc.UpdateStack(updateParams)

		if err != nil {
			return errors.New(fmt.Sprintln("Update failed: ", err))
		}

		go s.tail("UPDATE", done)

		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(s.stackname),
		}
		Log(fmt.Sprintln("Calling [WaitUntilStackUpdateComplete] with parameters:", describeStacksInput), level.debug)
		if err := svc.WaitUntilStackUpdateComplete(describeStacksInput); err != nil {
			return err
		}

		Log(fmt.Sprintf("Stack update successful: [%s]", s.stackname), "info")

	}
	done <- true
	return nil
}

func (s *stack) terminate() error {

	if !s.stackExists() {
		Log(fmt.Sprintf("%s: does not exist...", s.name), level.info)
		return nil
	}

	done := make(chan bool)
	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.DeleteStackInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [DeleteStack] with parameters:", params), level.debug)
	_, err := svc.DeleteStack(params)

	go s.tail("DELETE", done)

	if err != nil {
		done <- true
		return errors.New(fmt.Sprintln("Deleting failed: ", err))
	}

	// describeStacksInput := &cloudformation.DescribeStacksInput{
	// 	StackName: aws.String(s.stackname),
	// }
	//
	// Log(fmt.Sprintln("Calling [WaitUntilStackDeleteComplete] with parameters:", describeStacksInput), level.debug)
	//
	// if err := svc.WaitUntilStackDeleteComplete(describeStacksInput); err != nil {
	// 	return err
	// }

	// NOTE: The [WaitUntilStackDeleteComplete] api call suddenly stopped playing nice.
	// Implemented this crude loop as a patch fix for now
	for {
		if !s.stackExists() {
			done <- true
			break
		}

		time.Sleep(time.Second * 1)
	}

	Log(fmt.Sprintf("Deletion successful: [%s]", s.stackname), "info")

	return nil
}

func (s *stack) stackExists() bool {
	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [DescribeStacks] with parameters:", describeStacksInput), level.debug)
	_, err := svc.DescribeStacks(describeStacksInput)

	if err == nil {
		return true
	}

	return false
}

func (s *stack) status() error {
	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [DescribeStacks] with parameters:", describeStacksInput), level.debug)
	status, err := svc.DescribeStacks(describeStacksInput)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "exist") {
			fmt.Printf("create_pending -> %s [%s]"+"\n", s.name, s.stackname)
			return nil
		}
		return err
	}

	// Define time flag
	stat := *status.Stacks[0].StackStatus
	var timeflag time.Time
	switch strings.Split(stat, "_")[0] {
	case "UPDATE":
		timeflag = *status.Stacks[0].LastUpdatedTime
	default:
		timeflag = *status.Stacks[0].CreationTime
	}

	// Print Status
	fmt.Printf(
		"%s%s - %s --> %s - [%s]"+"\n",
		colorString(`@`, "magenta"),
		timeflag.Format(time.RFC850),
		strings.ToLower(colorMap(*status.Stacks[0].StackStatus)),
		s.name,
		s.stackname,
	)

	return nil
}

func (s *stack) state() (string, error) {
	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [DescribeStacks] with parameters: ", describeStacksInput), level.debug)
	status, err := svc.DescribeStacks(describeStacksInput)
	if err != nil {
		if strings.Contains(err.Error(), "not exist") {
			return state.pending, nil
		}
		return "", err
	}

	if strings.Contains(strings.ToLower(status.GoString()), "complete") {
		return state.complete, nil
	} else if strings.Contains(strings.ToLower(status.GoString()), "fail") {
		return state.failed, nil
	}
	return "", nil
}

func (s *stack) change(req string) error {
	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})

	switch req {

	case "create":
		// Resolve Deploy-Time functions
		err := s.deployTimeParser()
		if err != nil {
			return err
		}

		params := &cloudformation.CreateChangeSetInput{
			StackName:     aws.String(s.stackname),
			ChangeSetName: aws.String(job.changeName),
		}

		Log(fmt.Sprintf("Updated Template:\n%s", s.template), level.debug)

		// If bucket - upload to s3
		var (
			exists bool
			url    string
		)

		if s.bucket != "" {
			exists, err = BucketExists(s.bucket, s.session)
			if err != nil {
				Log(fmt.Sprintf("Received Error when checking if [%s] exists: %s", s.bucket, err.Error()), level.warn)
			}

			if !exists {
				Log(fmt.Sprintf(("Creating Bucket [%s]"), s.bucket), level.info)
				if err = CreateBucket(s.bucket, s.session); err != nil {
					return err
				}
			}
			t := time.Now()
			tStamp := fmt.Sprintf("%d-%d-%d_%d%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
			url, err = S3write(s.bucket, fmt.Sprintf("%s_%s.template", s.stackname, tStamp), s.template, s.session)
			if err != nil {
				return err
			}
			params.TemplateURL = &url
		} else {
			params.TemplateBody = &s.template
		}

		// If IAM is bening touched, add Capabilities
		if strings.Contains(s.template, "AWS::IAM") {
			params.Capabilities = []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
				aws.String(cloudformation.CapabilityCapabilityNamedIam),
			}
		}

		if _, err = svc.CreateChangeSet(params); err != nil {
			return err
		}

		describeParams := &cloudformation.DescribeChangeSetInput{
			StackName:     aws.String(s.stackname),
			ChangeSetName: aws.String(job.changeName),
		}

		for {
			// Waiting for PENDING state to change
			resp, err := svc.DescribeChangeSet(describeParams)
			if err != nil {
				return err
			}

			Log(fmt.Sprintf("Creating Change-Set: [%s] - %s - %s", job.changeName, colorMap(*resp.Status), s.stackname), level.info)

			if *resp.Status == "CREATE_COMPLETE" || *resp.Status == "FAILED" {
				break
			}

			time.Sleep(time.Second * 1)
		}

	case "rm":
		params := &cloudformation.DeleteChangeSetInput{
			ChangeSetName: aws.String(job.changeName),
			StackName:     aws.String(s.stackname),
		}

		if _, err := svc.DeleteChangeSet(params); err != nil {
			return err
		}

		Log(fmt.Sprintf("Change-Set: [%s] deleted", job.changeName), level.info)

	case "list":
		params := &cloudformation.ListChangeSetsInput{
			StackName: aws.String(s.stackname),
		}

		resp, err := svc.ListChangeSets(params)
		if err != nil {
			return err
		}

		// if strings.Contains(resp.GoString(), "Summaries:") {
		for _, i := range resp.Summaries {
			Log(fmt.Sprintf("%s%s - Change-Set: [%s] - Status: [%s]", colorString("@", "magenta"), i.CreationTime.Format(time.RFC850), *i.ChangeSetName, *i.ExecutionStatus), level.info)
		}
		// }

	case "execute":
		done := make(chan bool)
		params := &cloudformation.ExecuteChangeSetInput{
			StackName:     aws.String(s.stackname),
			ChangeSetName: aws.String(job.changeName),
		}

		if _, err := svc.ExecuteChangeSet(params); err != nil {
			return err
		}

		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(s.stackname),
		}

		go s.tail("UPDATE", done)

		Log(fmt.Sprintln("Calling [WaitUntilStackUpdateComplete] with parameters:", describeStacksInput), level.debug)
		if err := svc.WaitUntilStackUpdateComplete(describeStacksInput); err != nil {
			return err
		}

		done <- true

	case "desc":
		params := &cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(job.changeName),
			StackName:     aws.String(s.stackname),
		}

		resp, err := svc.DescribeChangeSet(params)
		if err != nil {
			return err
		}

		o, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", o)
	}

	return nil
}

func (s *stack) check() error {
	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(s.template),
	}

	Log(fmt.Sprintf("Calling [ValidateTemplate] with parameters:\n%s"+"\n--\n", params), level.debug)
	resp, err := svc.ValidateTemplate(params)
	if err != nil {
		return err
	}

	fmt.Printf(
		"%s\n\n%s"+"\n",
		colorString("Valid!", "green"),
		resp.GoString(),
	)

	return nil
}

func (s *stack) outputs() error {

	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})
	outputParams := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [DescribeStacks] with parameters:", outputParams), level.debug)
	outputs, err := svc.DescribeStacks(outputParams)
	if err != nil {
		return errors.New(fmt.Sprintln("Unable to reach stack", err.Error()))
	}

	// set stack outputs property
	s.output = outputs

	return nil
}

func (s *stack) stackPolicy() error {

	if s.policy == "" {
		return fmt.Errorf("Empty Stack Policy value detected...")
	}

	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.SetStackPolicyInput{
		StackName: &s.stackname,
	}

	// Check if source is a URL
	if strings.HasPrefix(s.policy, `http://`) || strings.HasPrefix(s.policy, `https://`) {
		params.StackPolicyURL = &s.policy
	} else {
		params.StackPolicyBody = &s.policy
	}

	Log(fmt.Sprintln("Calling SetStackPolicy with params: ", params), level.debug)
	resp, err := svc.SetStackPolicy(params)
	if err != nil {
		return err
	}

	Log(fmt.Sprintf("Stack Policy applied: [%s] - %s", s.stackname, resp.GoString()), level.info)

	return nil
}

// deployTimeParser - Parses templates during deployment to resolve specfic Dependency functions like stackout...
func (s *stack) deployTimeParser() error {

	// define Delims
	left, right := config.delims("deploy")

	// Create template
	t, err := template.New("deploy-template").Delims(left, right).Funcs(deployTimeFunctions).Parse(s.template)
	if err != nil {
		return err
	}

	// so that we can write to string
	var doc bytes.Buffer
	values := config.vars()

	// Add metadata specific to the stack we're working with to the parser
	values["stack"] = s.name
	values["parameters"] = s.parameters

	t.Execute(&doc, values)
	s.template = doc.String()
	Log(fmt.Sprintf("Deploy Time Template Generate:\n%s", s.template), level.debug)

	return nil
}

// genTimeParser - Parses templates before deploying them...
func (s *stack) genTimeParser() error {

	templ, err := fetchContent(s.source)
	if err != nil {
		return err
	}

	// define Delims
	left, right := config.delims("gen")

	// create template
	t, err := template.New("gen-template").Delims(left, right).Funcs(genTimeFunctions).Parse(templ)
	if err != nil {
		return err
	}

	// so that we can write to string
	var doc bytes.Buffer
	values := config.vars()

	// Add metadata specific to the stack we're working with to the parser
	values["stack"] = s.name
	values["parameters"] = s.parameters

	t.Execute(&doc, values)
	s.template = doc.String()
	return nil
}

// tail - tracks the progress during stack updates. c - command Type
func (s *stack) tail(c string, done <-chan bool) {
	svc := cloudformation.New(s.session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(s.stackname),
	}

	// used to track what lines have already been printed, to prevent dubplicate output
	printed := make(map[string]interface{})

	for {
		select {
		case <-done:
			Log("Tail Job Completed", level.debug)
			return
		default:
			// If channel is not populated, run verbose cf print
			Log(fmt.Sprintf("Calling [DescribeStackEvents] with parameters: %s", params), level.debug)
			stackevents, err := svc.DescribeStackEvents(params)
			if err != nil {
				Log(fmt.Sprintln("Error when tailing events: ", err.Error()), level.debug)
				// Sleep 2 seconds before next check - eep going until done signal
				time.Sleep(time.Duration(2 * time.Second))
				continue
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
}
