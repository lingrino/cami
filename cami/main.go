package cami

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// AWS is the main struct that holds our client and info
type AWS struct {
	sess *session.Session
	ec2  *ec2.EC2
}

// NewAWS returns a new AWS struct
func NewAWS() *AWS {
	return &AWS{}
}

// Auth sets up our AWS session and service clients
func (a *AWS) Auth() error {
	var err error

	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	a.sess = sess

	ec2 := ec2.New(a.sess)
	a.ec2 = ec2

	return err
}

// AMIs returns a list of all our AMIs
func (a *AWS) AMIs() ([]*ec2.Image, error) {
	var err error
	var output []*ec2.Image

	amiI := &ec2.DescribeImagesInput{
		Owners: []*string{aws.String("self")},
	}
	amiO, err := a.ec2.DescribeImages(amiI)
	if err != nil {
		return output, fmt.Errorf("failed to describe images: %w", err)
	}

	output = amiO.Images
	return output, err
}

// EC2s returns a list of all our EC2 instances using one of the AMIs in the list provided
func (a *AWS) EC2s(amis []*ec2.Image) ([]*ec2.Instance, error) {
	var err error
	var output []*ec2.Instance

	var amiIDs []*string
	for _, ami := range amis {
		amiIDs = append(amiIDs, ami.ImageId)
	}

	ec2I := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("image-id"),
				Values: amiIDs,
			},
		},
	}
	ec2O, err := a.ec2.DescribeInstances(ec2I)
	if err != nil {
		return output, fmt.Errorf("failed to describe instances: %w", err)
	}

	for _, res := range ec2O.Reservations {
		output = append(output, res.Instances...)
	}
	return output, err
}

// FilterAMIs returns back the list of AMIs with images in ec2s removed
func (a *AWS) FilterAMIs(amis []*ec2.Image, ec2s []*ec2.Instance) ([]*ec2.Image, error) {
	var err error
	var output []*ec2.Image

	var has bool
	for _, ami := range amis {
		has = false
		for _, ec2 := range ec2s {
			if ami.ImageId == ec2.ImageId {
				has = true
				break
			}
		}
		if !has {
			output = append(output, ami)
		}
	}

	return output, err
}

// DeleteAMIs deregisters all AMIs in the provided list and deletes the snapshots
// associated with the deregistered AMI
func (a *AWS) DeleteAMIs(amis []*ec2.Image) error {
	var err error

	for _, ami := range amis {
		amiI := &ec2.DeregisterImageInput{
			ImageId: ami.ImageId,
		}
		_, err := a.ec2.DeregisterImage(amiI)
		if err != nil {
			return fmt.Errorf("failed to deregister ami %v: %w", ami.ImageId, err)
		}

		for _, bdm := range ami.BlockDeviceMappings {
			snapID := bdm.Ebs.SnapshotId
			snapI := &ec2.DeleteSnapshotInput{
				SnapshotId: snapID,
			}
			_, err := a.ec2.DeleteSnapshot(snapI)
			if err != nil {
				return fmt.Errorf("failed to delete snapshot %v: %w", snapID, err)
			}
		}
	}

	return err
}
