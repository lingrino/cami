// Package cami provides a simple service for keeping AMIs clean.
package cami

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
)

type ec2If interface {
	DescribeImages(context.Context, *ec2.DescribeImagesInput, ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DeregisterImage(context.Context, *ec2.DeregisterImageInput, ...func(*ec2.Options)) (*ec2.DeregisterImageOutput, error)
	DeleteSnapshot(context.Context, *ec2.DeleteSnapshotInput, ...func(*ec2.Options)) (*ec2.DeleteSnapshotOutput, error)
}

// Config holds the configuration for our AWS struct.
type Config struct {
	// Set to true to run non-destructively
	DryRun bool
}

// AWS is the main struct that holds our client and info.
type AWS struct {
	cfg *Config

	// Used for testing
	ec2         ec2If
	filterErr   bool
	newEC2Fn    func(aws.Config, ...func(*ec2.Options)) *ec2.Client
	newConfigFn func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error)
}

// NewAWS returns a new AWS struct.
func NewAWS(c *Config) (*AWS, error) {
	a := &AWS{cfg: c}

	a.newEC2Fn = ec2.NewFromConfig
	a.newConfigFn = config.LoadDefaultConfig

	return a, nil
}

// Auth sets up our AWS session and service clients.
func (a *AWS) Auth() error {
	var err error

	cfg, err := a.newConfigFn(context.TODO())
	if err != nil {
		return fmt.Errorf("%w", ErrCreateSession)
	}

	ec2 := a.newEC2Fn(cfg)
	a.ec2 = ec2

	return err
}

// AMIs returns a list of all our AMIs.
func (a *AWS) AMIs() ([]types.Image, error) {
	var err error
	var output []types.Image

	amiI := &ec2.DescribeImagesInput{
		Owners: []string{"self"},
	}
	amiO, err := a.ec2.DescribeImages(context.TODO(), amiI)
	if err != nil {
		return output, fmt.Errorf("%w", ErrDesribeImages)
	}

	output = amiO.Images
	return output, nil
}

// EC2s returns a list of all our EC2 instances using one of the AMIs in the list provided.
func (a *AWS) EC2s(amis []types.Image) ([]types.Instance, error) {
	var output []types.Instance

	amiIDs := make([]string, len(amis))
	for _, ami := range amis {
		amiIDs = append(amiIDs, *ami.ImageId)
	}

	var nextToken *string
	for {
		ec2I := &ec2.DescribeInstancesInput{
			MaxResults: 1000, // nolint:gomnd
			Filters: []types.Filter{
				{
					Name:   aws.String("image-id"),
					Values: amiIDs,
				},
			},
		}
		if nextToken != nil {
			ec2I.NextToken = nextToken
		}

		out, err := a.ec2.DescribeInstances(context.TODO(), ec2I)
		if err != nil {
			return output, fmt.Errorf("%w", ErrDesribeInstances)
		}
		for _, res := range out.Reservations {
			output = append(output, res.Instances...)
		}
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	return output, nil
}

// FilterAMIs returns back the list of AMIs with images in ec2s removed.
func (a *AWS) FilterAMIs(amis []types.Image, ec2s []types.Instance) ([]types.Image, error) {
	var err error
	var output []types.Image

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
func (a *AWS) DeleteAMIs(amis []types.Image) ([]string, error) {
	var output []string
	eda := &ErrDeleteAMIs{}

	for _, ami := range amis {
		amiI := &ec2.DeregisterImageInput{
			ImageId: ami.ImageId,
			DryRun:  a.cfg.DryRun,
		}
		_, err := a.ec2.DeregisterImage(context.TODO(), amiI)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == "DryRunOperation" {
				output = append(output, *ami.ImageId)
			} else {
				eda.Append(*ami.ImageId)
			}
		} else {
			output = append(output, *ami.ImageId)
		}

		for _, bdm := range ami.BlockDeviceMappings {
			if bdm.Ebs == nil {
				continue
			}

			snapID := bdm.Ebs.SnapshotId
			snapI := &ec2.DeleteSnapshotInput{
				SnapshotId: snapID,
				DryRun:     a.cfg.DryRun,
			}
			_, err := a.ec2.DeleteSnapshot(context.TODO(), snapI)
			if err != nil {
				var apiErr smithy.APIError
				if errors.As(err, &apiErr) && apiErr.ErrorCode() == "DryRunOperation" {
					output = append(output, *bdm.Ebs.SnapshotId)
				} else {
					eda.Append(*snapID)
				}
			} else {
				output = append(output, *bdm.Ebs.SnapshotId)
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
