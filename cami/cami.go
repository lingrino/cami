package cami

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// Config holds the configuration for our AWS struct
type Config struct {
	// Set to true to run non-destructively
	DryRun bool
}

// AWS is the main struct that holds our client and info
type AWS struct {
	cfg *Config

	// Used for testing
	ec2          ec2iface.EC2API
	filterErr    bool
	newEC2Fn     func(client.ConfigProvider, ...*aws.Config) *ec2.EC2
	newSessionFn func(...*aws.Config) (*session.Session, error)
}

// NewAWS returns a new AWS struct
func NewAWS(c *Config) (*AWS, error) {
	a := &AWS{cfg: c}

	a.newEC2Fn = ec2.New
	a.newSessionFn = session.NewSession

	return &AWS{cfg: c}, nil
}

// Auth sets up our AWS session and service clients
func (a *AWS) Auth() error {
	var err error

	sess, err := a.newSessionFn()
	if err != nil {
		return fmt.Errorf("%w", ErrCreateSession)
	}

	ec2 := a.newEC2Fn(sess)
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
		return output, fmt.Errorf("%w", ErrDesribeImages)
	}

	output = amiO.Images
	return output, err
}

// EC2s returns a list of all our EC2 instances using one of the AMIs in the list provided
func (a *AWS) EC2s(amis []*ec2.Image) ([]*ec2.Instance, error) {
	var err error
	var output []*ec2.Instance

	amiIDs := make([]*string, len(amis))
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

	pageNum, pageMax := 0, 20
	err = a.ec2.DescribeInstancesPages(ec2I,
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			pageNum++

			for _, res := range page.Reservations {
				output = append(output, res.Instances...)
			}

			// Page 20 times at most
			return !lastPage || pageNum <= pageMax
		},
	)
	if err != nil {
		return output, fmt.Errorf("%w", ErrDesribeInstances)
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

	if a.filterErr {
		return output, ErrFilterAMIs
	}

	return output, err
}

// DeleteAMIs deregisters all AMIs in the provided list and deletes the snapshots
// associated with the deregistered AMI. Returns a list of IDs that were successfully
// deleted. If DryDrun == true does not actually delete.
func (a *AWS) DeleteAMIs(amis []*ec2.Image) ([]string, error) {
	var output []string
	eda := &ErrDeleteAMIs{}

	for _, ami := range amis {
		amiI := &ec2.DeregisterImageInput{
			ImageId: ami.ImageId,
			DryRun:  aws.Bool(a.cfg.DryRun),
		}
		_, err := a.ec2.DeregisterImage(amiI)
		if err != nil {
			var awsErr awserr.Error
			if errors.As(err, &awsErr) && awsErr.Code() == "DryRunOperation" {
				output = append(output, *ami.ImageId)
			} else {
				eda.Append(*ami.ImageId)
			}
		} else {
			output = append(output, *ami.ImageId)
		}

		for _, bdm := range ami.BlockDeviceMappings {
			if bdm.Ebs != nil {
				snapID := bdm.Ebs.SnapshotId
				snapI := &ec2.DeleteSnapshotInput{
					SnapshotId: snapID,
					DryRun:     aws.Bool(a.cfg.DryRun),
				}
				_, err := a.ec2.DeleteSnapshot(snapI)
				if err != nil {
					var awsErr awserr.Error
					if errors.As(err, &awsErr) && awsErr.Code() == "DryRunOperation" {
						output = append(output, *bdm.Ebs.SnapshotId)
					} else {
						eda.Append(*snapID)
					}
				} else {
					output = append(output, *bdm.Ebs.SnapshotId)
				}
			}
		}
	}

	return output, eda.ErrorOrNil()
}

// DeleteUnusedAMIs finds and deletes all AMIs (and their associated snapshots)
// that are not being used by any current EC2 instances in the same account.
func (a *AWS) DeleteUnusedAMIs() ([]string, error) {
	var err error
	var output []string

	amis, err := a.AMIs()
	if err != nil {
		return output, err
	}

	ec2s, err := a.EC2s(amis)
	if err != nil {
		return output, err
	}

	amis, err = a.FilterAMIs(amis, ec2s)
	if err != nil {
		return output, err
	}

	output, err = a.DeleteAMIs(amis)
	if err != nil {
		return output, err
	}

	return output, err
}
