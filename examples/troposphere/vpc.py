from troposphere import ec2, Output
from troposphere import Template, Ref, Tags

class VPC(object):
    def __init__(self):
        self.template = Template()
        
        # using  {{ .stack.cidr }} to fetch comma separated list of 
        # cidr blocks from qaz config
        # and removing spaces
        self.cidrs = [ c.strip(' ') for c in '{{ .stack.cidr }}'.split(',') ]

        # iterating cidr blocks
        for cidr in enumerate(self.cidrs):
            self.add_vpc(cidr[0], cidr[1])
    
    def add_vpc(self, index, cidr):
        # Add vpc
        t = self.template
        uid = 'VPC%s' % index

        vpc = t.add_resource(ec2.VPC(
            uid,
            # note the qaz template braces here
            # this value will be swapped out for the one in config
            CidrBlock=cidr,
            Tags=Tags(
                Name='qaz-tropos-test'
            )
        ))

        self.add_outputs(uid, vpc, desc="output for %s" % uid)

    def add_outputs(self, uid, value, desc=""):
        t = self.template

        t.add_output(Output(
            'output%s' % uid,
            Description=desc,
            Value=Ref(value)
        ))

    @staticmethod
    def template():
        return VPC().template.to_yaml()

# all troposphere templates must print a single
# stack/template to stdout when run
if __name__ == '__main__':
    print(VPC.template())
