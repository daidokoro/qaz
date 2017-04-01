# Multi-Account Deployment

This example shows the setup for a multiple AWS account deployment.

This feature explicitly requires AWS Account credentials to be configured correctly.


For this example, here's my aws config:

```toml
[profile default]
aws_secret_access_key = xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
aws_access_key_id = xxxxxxxxxxxxxxxxxxxxxx
region = eu-west-1

[profile lab]
aws_secret_access_key = xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
aws_access_key_id = xxxxxxxxxxxxxxxxxxxxxx
region = eu-west-1
```

--

My Configuration below has 2 stacks, _mainVPC_ which has no __profile__ keyword defined will use the default profile configuration to deploy and _labVPC_ which has _lab_ defined as the profile, will deploy using the __lab__ credentials in my config above.

```yaml
# Stacks
stacks:
  mainVPC:
    depends_on:
      - labVPC

    cf:
      cidr: 10.10.0.0/24

  labVPC:
    profile: lab

```
Note that the _mainVPC_ depends_on on the _labVPC_, so what we have is a cross-account stack dependency.

--

The template below is for the __mainVPC__ stack defined in the config file above.

```yaml
AWSTemplateFormatVersion: '2010-09-09'

Description: |
  This is an example VPC deployed via Qaz

Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: {{ .mainVPC.cidr }}

Outputs:
  vpcid:
    Description: VPC ID
    Value:  << stack_output "labVPC::vpcid" >>
    Export:
      Name: lap-vpc-id

```
Note the Outputs section, the Deploy-Time function `stack_output` is being used to export the Value of the _labVPC_ vpcid, in other words, the output of a stack in another account is being used as an export.
