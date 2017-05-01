 <p align="center">
  <img src="images/qaz.png">
</p>

[![Join the chat at https://gitter.im/qaz-go/Lobby](https://badges.gitter.im/qaz-go/Lobby.svg)](https://gitter.im/qaz-go/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![GitHub stars](https://img.shields.io/github/stars/daidokoro/qaz.svg)](https://github.com/daidokoro/qaz/stargazers)
[![Build Status](https://travis-ci.org/daidokoro/qaz.svg)](https://travis-ci.org/daidokoro/qaz)
![Go Report Card](https://goreportcard.com/badge/github.com/daidokoro/qaz)


__Qaz__ is a _cloud native_ AWS Cloudformation Template Management CLI tool inspired by the Bora project by [@pkazmierczak](https://github.com/pkazmierczak) that focuses on simplifying the process of deploying infrastructure on AWS via Cloudformation by utilising the Go Templates Library and custom functions to generate diverse and configurable templates.

For Qaz being _cloud native_ means having no explicit local dependencies and utilising resources within the AWS Ecosystem to extend functionality. As a result Qaz supports various methods for dynamically generating infrastructure via Cloudformation.

Qaz emphasizes minimal abstraction from the underlying AWS Cloudformation Platform. It instead enhances customisability and re-usability of templates through dynamic template creation and logic.

--

*Features:*

- Advanced template functionality & custom built-in template functions

- Support for Cloudformation templates written in JSON & YAML

- Dynamic deploy script generation utilising the built-in templating functionality

- Single YAML or JSON Configuration file for multiple stack templates per environment

- Utilises Go-routines for Multi-stack concurrent Cloudformation requests for *all* appropriate calls

- Support for AWS Profile selection for Multi-AWS account environments

- Cross stack referencing with support for Cloudformation Exports(_Preferred_) & dynamically retrieving stack outputs on deploy

- *Decoupled* build mechanism. Qaz can manage infrastructure by accessing config/templates via AWS Lambda, S3, or HTTP(S). The tool does not need to be in the same place as the templates/config.

- *Decoupled* stack management. Stacks can be launched individually from different locations and build consistently according to the dependency chain as long as the same configuration file is read.

- *Encryption* & *Decryption* of template values & deployment of encrypted templates using AWS KMS.

- Simultaneous Cross-Account or Cross-Region Stack Deployments.

- Support for fetching templates and configuration via Lambda Execution allows for dynamically generating Cloudformation using any of the Languages supported in AWS Lambda, (_nodejs, python, java_)
- __Troposphere__ support via Lambda.


## Installation

If you have Golang installed:

`go get github.com/daidokoro/qaz`

On Mac or Linux:

```
curl https://raw.githubusercontent.com/daidokoro/qaz/master/install.sh | sh
```

Or, you may need _sudo_:

```
curl https://raw.githubusercontent.com/daidokoro/qaz/master/install.sh | sudo sh
```

Windows EXE _coming soon_....

## Requirements
qaz requires:

- AWS credentials, you can read about how to set these up [here](http://blogs.aws.amazon.com/security/post/Tx3D6U6WSFGOK2H/A-New-and-Standardized-Way-to-Manage-Credentials-in-the-AWS-SDKs)

## Initialize Project

[![asciicast](https://asciinema.org/a/6d27ij32ev7ztarkfmrq5s0zg.png)](https://asciinema.org/a/6d27ij32ev7ztarkfmrq5s0zg?speed=2)


## Checkout the [Wiki](https://github.com/daidokoro/qaz/wiki) for more on how Qaz works!

### Table of contents

- [Home](https://github.com/daidokoro/qaz/wiki)
- [Installation](https://github.com/daidokoro/qaz/wiki/Install)
- [Handling Configuration Files](https://github.com/daidokoro/qaz/wiki/Config)
- [Custom Template Functions](https://github.com/daidokoro/qaz/wiki/Custom-Function)
- [Templating with Qaz](https://github.com/daidokoro/qaz/wiki/Templates)
- [Troposphere via Lambda](https://github.com/daidokoro/qaz/wiki/Troposphere)


--

See `examples` folder for more examples of usage. More examples to come.

```
$ qaz

Usage:
  qaz [flags]
  qaz [command]

Available Commands:
  change      Change-Set management for AWS Stacks
  check       Validates Cloudformation Templates
  deploy      Deploys stack(s) to AWS
  exports     Prints stack exports
  generate    Generates template from configuration values
  help        Help about any command
  init        Creates a basic qaz project
  invoke      Invoke AWS Lambda Functions
  outputs     Prints stack outputs
  set-policy  Set Stack Policies based on configured value
  status      Prints status of deployed/un-deployed stacks
  terminate   Terminates stacks
  update      Updates a given stack

Flags:
      --debug            Run in debug mode...
  -p, --profile string   configured aws profile (default "default")
      --version          print current/running version

Use "qaz [command] --help" for more information about a command.

```

--
## Roadmap and status
Qaz is now in __beta__, no more breaking changes to come. The focus from this point on is stability.

*TODO:*

- More Comprehensive Documentation
- More Deploy/Gen-Time Functions

--

# Contributing

Fork -> Patch -> Push -> Pull Request

_Pull requests welcomed...._
