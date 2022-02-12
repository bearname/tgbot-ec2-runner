package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type EC2Service struct {
	svc ec2.EC2
}

func NewEC2Service(svc *ec2.EC2) *EC2Service {
	e := new(EC2Service)
	e.svc = *svc
	return e
}

// StartInstance starts an Amazon EC2 instance.
// Inputs:
//     svc is an Amazon EC2 service client
//     instanceID is the ID of the instance
// Output:
//     If success, nil
//     Otherwise, an error from the call to StartInstances
func (s *EC2Service) StartInstance(instanceID *string) error {
	input := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			instanceID,
		},
		DryRun: aws.Bool(true),
	}
	_, err := s.svc.StartInstances(input)
	awsErr, ok := err.(awserr.Error)

	if ok && awsErr.Code() == "DryRunOperation" {
		// Set DryRun to be false to enable starting the instances
		input.DryRun = aws.Bool(false)
		_, err = s.svc.StartInstances(input)
		if err != nil {
			return err
		}

		return nil
	}

	return err
}

// StopInstance stops an Amazon EC2 instance.
// Inputs:
//     svc is an Amazon EC2 service client
//     instance ID is the ID of the instance
// Output:
//     If success, nil
//     Otherwise, an error from the call to StopInstances
func (s *EC2Service) StopInstance(instanceID *string) error {
	input := &ec2.StopInstancesInput{
		InstanceIds: []*string{
			instanceID,
		},
		DryRun: aws.Bool(true),
	}
	_, err := s.svc.StopInstances(input)
	awsErr, ok := err.(awserr.Error)
	if ok && awsErr.Code() == "DryRunOperation" {
		input.DryRun = aws.Bool(false)
		_, err = s.svc.StopInstances(input)
		if err != nil {
			return err
		}

		return nil
	}

	return err
}
