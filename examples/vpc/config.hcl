region = "eu-west-1"

project = "daidokoro"

global {
  tags = [
    {
      code = "go"
    },
    {
      service = "example"
    },
    {
      life = "live"
    },
  ]
}

// vpc stack
stacks "vpc" {
  policy = "https://s3-eu-west-1.amazonaws.com/daidokoro-dev/policies/stack.json"
  source = "https://raw.githubusercontent.com/daidokoro/qaz/master/examples/vpc/templates/vpc.yml"

  // Cloudformation values
  cf {
    cidr = "10.10.0.0/16"
  }
}

// subnet stack
stacks "subnet" {
  depends_on = ["vpc"]
  source     = "https://raw.githubusercontent.com/daidokoro/qaz/master/examples/vpc/templates/subnet.yml"

  // Cloudformation values
  cf {
    subnets = [
      {
        private = "10.10.0.0/24"
      },
      {
        public = "10.10.2.0/24"
      },
    ]
  }
}
