# qaze—a simple AWS Cloudformation wrapper 
[![Build Status](https://travis-ci.org/daidokoro/qaze.svg)](https://travis-ci.org/daidokoro/qaze)

Qaze is a Fork of the Bora project by [@pkazmierczak](https://github.com/pkazmierczak) that aims to focus on simplifying the process of deploying infrastructure on AWS via Cloudformation by utilising the Go Templates Library.  

## Objective

Qaze's goal is to be a tool that's easy to use, with mininal abstraction from the underlying AWS Cloudformation Platform. Minimal abstraction will future proof against failure and allow quick access to new functionality when Amazon inevitably enhance the Cloudformation Platform. In other words, _it keeps working even when they change stuff...._


## Features
- Utilises Go's built in Template functionality to enhance templating logic
- Supports Cloudformation YAML/JSON & has the ability to template other files as well
- No figiting around with additional code for writing templates
- Multi-stack concurrent cloudformation requests utilising goroutines
- Single set of templates, 1 config for all stacks per-environment *OR* 1 config file for ALL stacks for ALL environments - __qaze__ won't stand in your way.


## Installation

If you have Golang installed:

`go get github.com/daidokoro/qaze`

#### Alternatively

Pre-build binaries for Darwin and Linux are [available](https://github.com/daidokoro/qaze/releases).

## Usage
qaze requires:

- AWS credentials, you can read about how to set these up [here](http://blogs.aws.amazon.com/security/post/Tx3D6U6WSFGOK2H/A-New-and-Standardized-Way-to-Manage-Credentials-in-the-AWS-SDKs)


To kicking off your first project!

![Alt text](demo/init.gif?raw=true "Demo Init Command")

#### Deploying Stacks

Stacks can be Deployed/Terminated with a single command (works only in the root directory).

![Alt text](demo/quick_build.gif?raw=true "Quick Build Demo")


To deploy a specific stack or select specific template(s) for deployment:
```
$ qaze deploy -c path/to/config.yml -t path/to/template -t path/to/temple
```


Deploy also takes wildcards for templates. For example:
```
$ qaze deploy -c path/to/config.yml -t path/*
```
_Will deploy all tepmlates under `path`_

--

See `examples` folder for more on usage.

```
qaze is a simple wrapper around cloudformation.

Usage:
  qaze [flags]
  qaze [command]

Available Commands:
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
### Go Templates

__qaze__ utilises the Go Template library for template logic and can take full advantage of the various methods and features of same. For documentation, see [Here](https://golang.org/pkg/text/template)

More Examples and documentation to come....

## Roadmap and status
qaze is in early development.

*TODO:*

- Implement Change-Set management
- Implement Functionality to read-in strings/values from files
- Restructure Code for better exception handling
- Debug flag and proper logging
- More Comprehensive Documentation

