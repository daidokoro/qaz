# qaze—A CLI tool for Templating & Managing stacks in AWS Cloudformation
[![Build Status](https://travis-ci.org/daidokoro/qaze.svg)](https://travis-ci.org/daidokoro/qaze)

Qaze is a Fork of the Bora project by [@pkazmierczak](https://github.com/pkazmierczak) that aims to focus on simplifying the process of deploying infrastructure on AWS via Cloudformation by utilising the Go Templates Library and custom functions to generate diverse and configurable tables.

Qaze focuses on mininal abstraction from the underlying AWS Cloudformation Platform and instead focuses on dyanamic template generation, using Go Templates and custom templating functions. This allows it to be future proof against failure and allow quick access to new functionality when Amazon inevitably enhance the Cloudformation Platform.


*Features:*
- Advanced templating functionality & custom built-in template functions 

- Support for templates written in JSON & YAML

- Single YAML Configuration file for multiple stack templates per environment

- Utilises Goroutines for Multi-stack concurrent Cloudformation requests for *all* appropriate calls

- Support for AWS Profile selection for Multi-AWS account environments

- *Decoupled* build mechanism. Qaze can manage infrasture by accessing config/templates via S3 or HTTP(S). This way the tool does not need to be stored with the files.

- *Decoupled* stack managemnt. Stacks can be launched individually from different locations and build according to the depency chain as long as the same configuration file is read. 


## Installation

If you have Golang installed:

`go get github.com/daidokoro/qaze`

Pre-build binaries for Darwin and Linux coming soon....

## Requirements
qaze requires:

- AWS credentials, you can read about how to set these up [here](http://blogs.aws.amazon.com/security/post/Tx3D6U6WSFGOK2H/A-New-and-Standardized-Way-to-Manage-Credentials-in-the-AWS-SDKs)


## How It Works!

*qaze* get's uses a main _config.yml_ file as its source-of-truth. This file tells it what stacks it controls and what values to pass to those stacks. 

```yaml

# Specify the AWS region code
# qaze will attempt to get it from AWS configuration
# or from the environment. This setting overrides
# every other.
region: "eu-west-1"

# Required: The project name is prepended to the
# stack names at build time to create unique 
# identifier on the Cloudformation platform
project: "daidokoro"

# Optional: global values, accisible accross
# all stacks can be define under global
global:


# All stack specific values are defined
# under the "stacks" keyword below. 

stacks:
  vpc:
    # Note: "cf" is a required keyword, which tells 
    # qaze when to start reading in template values.
    cf:
      cidr: 10.10.0.0/16
  
  subnet:
    # Note: the "depends_on" keyword is used to list 
    # stack dependencies. Any amount can be listed.
    # This key_word must be defined outside of "cf"
    depends_on:
      - vpc
    
    cf:
      subnets:
        - private: 10.10.0.0/24
        - public: 10.10.2.0/24
```

Note: Config files do not need to be named config.yml

--

## Templates (Getting those values!)

Go has an excellent and expandable templating library which is utilised this in this project to project addition logic in creating templates. To read more on Go Template see [Here](https://golang.org/pkg/text/template)

We'll run through some basic tips and tricks to get you started.

Note that templates must have the same file name (_extension excluded_) as the stack they reference in config when working with local files, however, this does not apply when doing remote calls via S3 or Http for templates.

--

To access the values in our template we need to use templating syntax. In it's most basic form, to fetch say the value for `cidr` in my vpc stack config I would do the following:

```yaml
{{ .vpc.cidr }}
```

That's it! Use the generate command to varify the value 
`$ qaze generate -c path/to/config -t path/to/template`

--

Go Templates are also capable of looping values, for example, to get the values of both _private_ & _public_ in my *subnets* stack, I would do the following.

``` yaml
{{ range $index, $value := .subnets.subnets }} # "range" allows us to loop over items in the template
  {{ range $access, $cidr := $value }} # looping over the key value pairs 
    {{$access}} {{$cidr}} # printing output
  {{ end }}
{{ end }} # Closing loops
```

The above should give you the access level and subnets defined above. More examples as well as the full template implementation of this example can be found in the project examples folder.



#### Deploying Stacks

Stacks can be Deployed/Terminated with a single command.

![Alt text](demo/quick_build.gif?raw=true "Quick Build Demo")

The above however, only works when you are using Qaze in the root of your project directory. Alternatively, Qaze offers a few ways fetching both configuration and template files.

Configuration can be retreived from both Http Get requests & S3.

```
$ qaze deploy -c s3://mybucket/super_config.yml -t vpc::http://someurl/vpc_dev.yml
```

The stack name must be specified using the syntax above. This tells Qaze what values to associate with this stack.

```
$ qaze deploy -c http://mybucket/super_config.yml -t vpc::s3://mybucket/vpc_dev.yml -t subnets::s3://mybucket/subnets.yml
```

You can pass as many `-t` flags as you have stacks, Qaze will deploy all in the correct order and manage the dependency chains as long as the `depends_on` keyword is utilised.

Note that the syntax for specifying stack names with URLs `stackname::url`. The deploy command does not require the stack name syntax when using local files, however the `update` command uses this syntax on *all* `-t --template` arguments. For example:

```
$ qaze deploy -c path/to/config -t path/to/template
$ qaze update -c path/to/config -t vpc::path/to/template
```

Deploy also takes wildcards for local templates. For example:

```
$ qaze deploy -c path/to/config.yml -t "path/*"
```
Quotes are required when using wildcards.

--

 ### Built in Template Functions

Template Functions expand the functionality of Go's Templating library by allowing you to execute external functions to retreive additional information for building your template.

Qaze supports all the Go Template functions as well as some custom ones. Three of these are:

__File:__

A template function for reading values from an external file into a template. For now the file needs to be in the `files` directory in the rood of the project folder.

Example:

```
{{ myfile.txt | File }} # Returns the value of myfile.txt under the files directory
```

__S3Read:__

As the name suggests, this function reads the content of a given s3 key and writes it to the template.

Example:

```
{{ "s3://mybucket/key" | S3Read }} # writes the contents of the object to the template
```

__GET:__

GET implements http GET requests on a given url, and writes the response to the template.

Example
```
{{ "http://localhost" | GET }}
```

--


See `examples` folder for more on usage. More examples to come.

```
$ qaze

qaze is a simple wrapper around Cloudformation.

Usage:
  qaze [flags]
  qaze [command]

Available Commands:
  check       Validates Cloudformation Templates
  deploy      Deploys stack(s) to AWS
  generate    Generates a JSON or YAML template
  init        Creates a basic qaze project
  outputs     Prints stack outputs/exports
  status      Prints status of deployed/un-deployed stacks
  terminate   Terminates stacks
  update      Updates a given stack

Flags:
  -p, --profile string   configured aws profile (default "default")
      --version          print current/running version

Use "qaze [command] --help" for more information about a command.
```


--
## Roadmap and status
qaze is in early development.

*TODO:*

- Implement Change-Set management
- Implement proper logging, log-levels & debug
- Restructure Code for better exception handling
- More Comprehensive Documentation

