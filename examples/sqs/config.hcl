region = "eu-central-1"

project = "example-dev"

// SQS stack
stacks "sqs" {
  source = "https://raw.githubusercontent.com/daidokoro/qaz/master/examples/sqs/templates/sqs.yml"

  cf = {
    indexdocument = "main.html"

    Queues = [
      {
        QueueName = "daido"
      },
      {
        QueueName = "koro"
      },
    ]
  }
}
