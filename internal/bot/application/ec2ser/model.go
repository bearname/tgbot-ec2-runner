package ec2ser

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Ec2Instance struct {
	Id            string
	Name          string
	PublicDnsName string
	State         ec2.InstanceState
}
