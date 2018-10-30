from troposphere import ec2, Output
from troposphere import Template, Ref, Tags

class VPC():
    def __init__(self):
        self.template = Template()

        # Build Template
        self.add_vpc()
        self.add_outputs()

    def add_vpc(self):
        # Add vpc
        t = self.template

        self.vpc = t.add_resource(ec2.VPC(
            'VPC',
            # note the qaz template bracets here
            CidrBlock='{{ .stack.cidr }}',
            Tags=Tags(
                Name='qaz-tropos-test'
            )
        ))

    def add_outputs(self):
        t = self.template

        t.add_output(Output(
            'vpcid',
            Description="Vpc ID",
            Value=Ref(self.vpc)
        ))

    @staticmethod
    def template():
        return VPC().template.to_json()

# all troposphere templates must print a single
# stack/template to stdout when run
if __name__ == '__main__':
    print(VPC.template())
