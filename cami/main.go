package cami

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Config holds the configuration for our AWS struct
type Config struct {
	// Set to true to run non-destructively
	DryRun bool
}

// AWS is the main struct that holds our client and info
type AWS struct {
	cfg *Config

	sess *session.Session
	ec2  *ec2.EC2
}

// NewAWS returns a new AWS struct
func NewAWS(c *Config) (*AWS, error) {
	return &AWS{cfg: c}, nil
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

	pageNum, pageMax := 0, 20
	err = a.ec2.DescribeInstancesPages(ec2I,
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			pageNum++

			for _, res := range ec2O.Reservations {
				output = append(output, res.Instances...)
			}

			// Page 20 times at most
			return !lastPage || pageNum <= pageMax
		},
	)
	if err != nil {
		return output, fmt.Errorf("failed to describe instances: %w", err)
	}

	return output, err
}

// FilterAMIs returns back the list of AMIs with images in ec2s removed
func (a *AWS) FilterAMIs(amis []*ec2.Image, ec2s []*ec2.Instance) ([]*ec2.Image, error) {
	var err error
	var output []*ec2.Image

	hasD := make(map[string]bool)
	for _, ec2 := range ec2s {
		hasD[*ec2.ImageId] = true
	}

	for _, ami := range amis {
		if _, ok := hasD[*ami.ImageId]; !ok {
			output = append(output, ami)
		}
	}

	return output, err
}

// DeleteAMIs deregisters all AMIs in the provided list and deletes the snapshots
// associated with the deregistered AMI. Returns a list of IDs that were successfully
// deleted. If DryDrun == true does not actually delete.
func (a *AWS) DeleteAMIs(amis []*ec2.Image) ([]string, error) {
	var err error
	var output []string

	for _, ami := range amis {
		if !a.cfg.DryRun {
			amiI := &ec2.DeregisterImageInput{
				ImageId: ami.ImageId,
			}
			_, err := a.ec2.DeregisterImage(amiI)
			if err != nil {
				err = fmt.Errorf("failed to deregister ami %v: %w", ami.ImageId, err)
			}
		}
		output = append(output, *ami.ImageId)

		for _, bdm := range ami.BlockDeviceMappings {
			if bdm.Ebs != nil {
				if !a.cfg.DryRun {
					snapID := bdm.Ebs.SnapshotId
					snapI := &ec2.DeleteSnapshotInput{
						SnapshotId: snapID,
					}
					_, err := a.ec2.DeleteSnapshot(snapI)
					if err != nil {
						err = fmt.Errorf("failed to deregister ami %v: %w", ami.ImageId, err)
					}
				}
				output = append(output, *bdm.Ebs.SnapshotId)
			}
		}
	}

	return output, err
}

// DeleteUnusedAMIs finds and deletes all AMIs (and their associated snapshots)
// that are not being used by any current EC2 instances in the same account.
func (a *AWS) DeleteUnusedAMIs() ([]string, error) {
	var err error
	var output []string

	amis, err := a.AMIs()
	if err != nil {
		return output, fmt.Errorf("failed to describe AMIs: %w", err)
	}

	ec2s, err := a.EC2s(amis)
	if err != nil {
		return output, fmt.Errorf("failed to describe EC2s: %w", err)
	}

	amis, err = a.FilterAMIs(amis, ec2s)
	if err != nil {
		return output, fmt.Errorf("failed to filter AMIs: %w", err)
	}

	output, err = a.DeleteAMIs(amis)
	if err != nil {
		return output, fmt.Errorf("failed to delete all AMIs and snapshots: %w", err)
	}

	return output, err
}
