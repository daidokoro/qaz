# Simple VPC

This examples demonstrates cross-stack referencing using a Deploy-Time function: `<< stack_output "vpc::vpcid" >>` in `templates/subnets.yml`

Cloudformation Import/Export is still the recommended way to do this, however, this method has it's uses also.
