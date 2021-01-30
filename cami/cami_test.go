package cami

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

func TestNewAWS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		give    *Config
		wantAWS *AWS
		wantErr error
	}{
		{
			name: "nil",
			give: nil,
			wantAWS: &AWS{
				cfg: nil,
				ec2: nil,
			},
			wantErr: nil,
		},
		{
			name: "empty",
			give: &Config{},
			wantAWS: &AWS{
				cfg: &Config{DryRun: false},
				ec2: nil,
			},
			wantErr: nil,
		},
		{
			name: "dryrun",
			give: &Config{DryRun: true},
			wantAWS: &AWS{
				cfg: &Config{DryRun: true},
				ec2: nil,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aws, err := NewAWS(tt.give)

			if tt.wantErr == nil {
				assert.Nil(t, err)
			} else {
				assert.True(t, errors.Is(err, tt.wantErr), fmt.Sprintf("expected: %s\ngot: %s", tt.wantErr, err))
			}

			assert.Equal(t, tt.wantAWS.cfg, aws.cfg)
			assert.Equal(t, tt.wantAWS.ec2, aws.ec2)
		})
	}
}

func TestAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		give    *AWS
		wantAWS *AWS
		wantErr error
	}{
		{
			name: "error",
			give: &AWS{
				newConfigFn: func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
					return aws.Config{}, fmt.Errorf("FAIL")
				},
			},
			wantAWS: &AWS{},
			wantErr: ErrCreateSession,
		},
		{
			name: "valid",
			give: &AWS{
				newEC2Fn: func(aws.Config, ...func(*ec2.Options)) *ec2.Client { return &ec2.Client{} },
				newConfigFn: func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
					return aws.Config{}, nil
				},
			},
			wantAWS: &AWS{ec2: &ec2.Client{}},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.give.Auth()

			if tt.wantErr == nil {
				assert.Nil(t, err)
			} else {
				assert.True(t, errors.Is(err, tt.wantErr), fmt.Sprintf("expected: %s\ngot: %s", tt.wantErr, err))
			}

			assert.Equal(t, tt.wantAWS.ec2, tt.give.ec2)
		})
	}
}

func TestAMIs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		giveOutput ec2.DescribeImagesOutput
		giveErr    error
		wantAMIs   []types.Image
		wantErr    error
	}{
		{
			name:       "empty",
			giveOutput: ec2.DescribeImagesOutput{},
			giveErr:    nil,
			wantAMIs:   nil,
			wantErr:    nil,
		},
		{
			name:       "error",
			giveOutput: ec2.DescribeImagesOutput{},
			giveErr:    fmt.Errorf("FAIL"),
			wantAMIs:   nil,
			wantErr:    ErrDesribeImages,
		},
		{
			name: "images",
			giveOutput: ec2.DescribeImagesOutput{
				Images: []types.Image{
					{ImageId: aws.String("ami-123")},
					{ImageId: aws.String("ami-456")},
				},
			},
			giveErr: nil,
			wantAMIs: []types.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			wantErr: nil,
		},
		{
			name: "images with error",
			giveOutput: ec2.DescribeImagesOutput{
				Images: []types.Image{
					{ImageId: aws.String("ami-123")},
					{ImageId: aws.String("ami-456")},
				},
			},
			giveErr:  fmt.Errorf("FAIL"),
			wantAMIs: nil,
			wantErr:  ErrDesribeImages,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aws := AWS{
				ec2: &mockEC2{
					RespDescImages:    tt.giveOutput,
					RespDescImagesErr: tt.giveErr,
				},
			}

			amis, err := aws.AMIs()

			if tt.wantErr == nil {
				assert.Nil(t, err)
			} else {
				assert.True(t, errors.Is(err, tt.wantErr), fmt.Sprintf("expected: %s\ngot: %s", tt.wantErr, err))
			}

			assert.Equal(t, tt.wantAMIs, amis)
		})
	}
}

func TestEC2s(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		giveImages []types.Image
		giveOutput ec2.DescribeInstancesOutput
		giveErr    error
		givePages  int
		wantEC2s   []types.Instance
		wantErr    error
	}{
		{
			name:       "empty",
			giveImages: []types.Image{},
			giveOutput: ec2.DescribeInstancesOutput{},
			giveErr:    nil,
			wantEC2s:   nil,
			wantErr:    nil,
		},
		{
			name:       "error",
			giveImages: []types.Image{},
			giveOutput: ec2.DescribeInstancesOutput{},
			giveErr:    fmt.Errorf("FAIL"),
			wantEC2s:   nil,
			wantErr:    ErrDesribeInstances,
		},
		{
			name: "one reservation",
			giveImages: []types.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveOutput: ec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{{Instances: []types.Instance{
					{ImageId: aws.String("ami-123")},
					{ImageId: aws.String("ami-456")},
				}}},
			},
			giveErr: nil,
			wantEC2s: []types.Instance{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			wantErr: nil,
		},
		{
			name: "two reservations",
			giveImages: []types.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveOutput: ec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{
					{Instances: []types.Instance{{ImageId: aws.String("ami-123")}}},
					{Instances: []types.Instance{{ImageId: aws.String("ami-456")}}},
				},
			},
			giveErr: nil,
			wantEC2s: []types.Instance{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			wantErr: nil,
		},
		{
			name: "two pages",
			giveImages: []types.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveOutput: ec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{
					{Instances: []types.Instance{{ImageId: aws.String("ami-123")}}},
					{Instances: []types.Instance{{ImageId: aws.String("ami-456")}}},
				},
			},
			giveErr: nil,
			wantEC2s: []types.Instance{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aws := AWS{
				ec2: &mockEC2{
					RespDescInstances:    tt.giveOutput,
					RespDescInstancesErr: tt.giveErr,
				},
			}

			ec2s, err := aws.EC2s(tt.giveImages)

			if tt.wantErr == nil {
				assert.Nil(t, err)
			} else {
				assert.True(t, errors.Is(err, tt.wantErr), fmt.Sprintf("expected: %s\ngot: %s", tt.wantErr, err))
			}

			assert.Equal(t, tt.wantEC2s, ec2s)
		})
	}
}

func TestFilterAMIs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		giveAMIs []types.Image
		giveEC2s []types.Instance
		wantAMIs []types.Image
		wantErr  error
	}{
		{
			name:     "nil",
			giveAMIs: nil,
			giveEC2s: nil,
			wantAMIs: nil,
			wantErr:  nil,
		},
		{
			name:     "empty",
			giveAMIs: []types.Image{},
			giveEC2s: []types.Instance{},
			wantAMIs: nil,
			wantErr:  nil,
		},
		{
			name: "filter",
			giveAMIs: []types.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveEC2s: []types.Instance{
				{ImageId: aws.String("ami-456")},
				{ImageId: aws.String("ami-789")},
			},
			wantAMIs: []types.Image{
				{ImageId: aws.String("ami-123")},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aws := AWS{}

			filtered, err := aws.FilterAMIs(tt.giveAMIs, tt.giveEC2s)

			if tt.wantErr == nil {
				assert.Nil(t, err)
			} else {
				assert.True(t, errors.Is(err, tt.wantErr), fmt.Sprintf("expected: %s\ngot: %s", tt.wantErr, err))
			}

			assert.Equal(t, tt.wantAMIs, filtered)
		})
	}
}

func TestDeleteAMIs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		giveAMIs               []types.Image
		giveDeregisterImage    ec2.DeregisterImageOutput
		giveDeregisterImageErr error
		giveDeleteSnapshot     ec2.DeleteSnapshotOutput
		giveDeleteSnapshotErr  error
		wantIDs                []string
		wantErr                error
	}{
		{
			name:                   "nil",
			giveAMIs:               nil,
			giveDeregisterImage:    ec2.DeregisterImageOutput{},
			giveDeregisterImageErr: nil,
			giveDeleteSnapshot:     ec2.DeleteSnapshotOutput{},
			giveDeleteSnapshotErr:  nil,
			wantIDs:                nil,
			wantErr:                nil,
		},
		{
			name:                   "empty",
			giveAMIs:               []types.Image{},
			giveDeregisterImage:    ec2.DeregisterImageOutput{},
			giveDeregisterImageErr: nil,
			giveDeleteSnapshot:     ec2.DeleteSnapshotOutput{},
			giveDeleteSnapshotErr:  nil,
			wantIDs:                nil,
			wantErr:                nil,
		},
		{
			name: "deregister image error",
			giveAMIs: []types.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveDeregisterImage:    ec2.DeregisterImageOutput{},
			giveDeregisterImageErr: fmt.Errorf("FAIL"),
			giveDeleteSnapshot:     ec2.DeleteSnapshotOutput{},
			giveDeleteSnapshotErr:  nil,
			wantIDs:                nil,
			wantErr: &ErrDeleteAMIs{
				IDs: []string{"ami-123", "ami-456"},
			},
		},
		{
			name: "delete snapshot error",
			giveAMIs: []types.Image{
				{
					ImageId: aws.String("ami-123"),
					BlockDeviceMappings: []types.BlockDeviceMapping{
						{
							Ebs: &types.EbsBlockDevice{
								SnapshotId: aws.String("snap-123"),
							},
						},
					},
				},
			},
			giveDeregisterImage:    ec2.DeregisterImageOutput{},
			giveDeregisterImageErr: nil,
			giveDeleteSnapshot:     ec2.DeleteSnapshotOutput{},
			giveDeleteSnapshotErr:  fmt.Errorf("FAIL"),
			wantIDs:                []string{"ami-123"},
			wantErr: &ErrDeleteAMIs{
				IDs: []string{"snap-123"},
			},
		},
		{
			name: "images no snaps",
			giveAMIs: []types.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveDeregisterImage:    ec2.DeregisterImageOutput{},
			giveDeregisterImageErr: nil,
			giveDeleteSnapshot:     ec2.DeleteSnapshotOutput{},
			giveDeleteSnapshotErr:  nil,
			wantIDs: []string{
				"ami-123",
				"ami-456",
			},
			wantErr: nil,
		},
		{
			name: "images no snaps dry run",
			giveAMIs: []types.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveDeregisterImage:    ec2.DeregisterImageOutput{},
			giveDeregisterImageErr: mockErr{ErrCode: "DryRunOperation"},
			giveDeleteSnapshot:     ec2.DeleteSnapshotOutput{},
			giveDeleteSnapshotErr:  nil,
			wantIDs: []string{
				"ami-123",
				"ami-456",
			},
			wantErr: nil,
		},
		{
			name: "images and snaps dry run",
			giveAMIs: []types.Image{
				{
					ImageId: aws.String("ami-123"),
					BlockDeviceMappings: []types.BlockDeviceMapping{
						{
							Ebs: &types.EbsBlockDevice{
								SnapshotId: aws.String("snap-123"),
							},
						},
					},
				},
			},
			giveDeregisterImage:    ec2.DeregisterImageOutput{},
			giveDeregisterImageErr: nil,
			giveDeleteSnapshot:     ec2.DeleteSnapshotOutput{},
			giveDeleteSnapshotErr:  mockErr{ErrCode: "DryRunOperation"},
			wantIDs: []string{
				"ami-123",
				"snap-123",
			},
			wantErr: nil,
		},
		{
			name: "images and snaps",
			giveAMIs: []types.Image{
				{
					ImageId: aws.String("ami-123"),
					BlockDeviceMappings: []types.BlockDeviceMapping{
						{
							Ebs: &types.EbsBlockDevice{
								SnapshotId: aws.String("snap-123"),
							},
						},
					},
				},
				{
					ImageId: aws.String("ami-456"),
					BlockDeviceMappings: []types.BlockDeviceMapping{
						{
							Ebs: &types.EbsBlockDevice{
								SnapshotId: aws.String("snap-456"),
							},
						},
						{
							Ebs: &types.EbsBlockDevice{
								SnapshotId: aws.String("snap-789"),
							},
						},
					},
				},
			},
			giveDeregisterImage:    ec2.DeregisterImageOutput{},
			giveDeregisterImageErr: nil,
			giveDeleteSnapshot:     ec2.DeleteSnapshotOutput{},
			giveDeleteSnapshotErr:  nil,
			wantIDs: []string{
				"ami-123",
				"snap-123",
				"ami-456",
				"snap-456",
				"snap-789",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aws := AWS{
				cfg: &Config{},
				ec2: &mockEC2{
					RespDeregisterImage:    tt.giveDeregisterImage,
					RespDeregisterImageErr: tt.giveDeregisterImageErr,
					RespDeleteSnapshot:     tt.giveDeleteSnapshot,
					RespDeleteSnapshotErr:  tt.giveDeleteSnapshotErr,
				},
			}

			ids, err := aws.DeleteAMIs(tt.giveAMIs)

			if tt.wantErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tt.wantErr, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}

			assert.Equal(t, tt.wantIDs, ids)
		})
	}
}

func TestDeleteUnusedAMIs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		give    *AWS
		wantIDs []string
		wantErr error
	}{
		{
			name: "empty",
			give: &AWS{
				ec2: &mockEC2{
					RespDescImages:    ec2.DescribeImagesOutput{},
					RespDescImagesErr: nil,

					RespDescInstances:    ec2.DescribeInstancesOutput{},
					RespDescInstancesErr: nil,

					RespDeregisterImage:    ec2.DeregisterImageOutput{},
					RespDeregisterImageErr: nil,

					RespDeleteSnapshot:    ec2.DeleteSnapshotOutput{},
					RespDeleteSnapshotErr: nil,
				},
			},
			wantIDs: nil,
			wantErr: nil,
		},
		{
			name: "error describe images",
			give: &AWS{
				ec2: &mockEC2{
					RespDescImages:    ec2.DescribeImagesOutput{},
					RespDescImagesErr: fmt.Errorf("FAIL"),

					RespDescInstances:    ec2.DescribeInstancesOutput{},
					RespDescInstancesErr: nil,

					RespDeregisterImage:    ec2.DeregisterImageOutput{},
					RespDeregisterImageErr: nil,

					RespDeleteSnapshot:    ec2.DeleteSnapshotOutput{},
					RespDeleteSnapshotErr: nil,
				},
			},
			wantIDs: nil,
			wantErr: ErrDesribeImages,
		},
		{
			name: "error filter",
			give: &AWS{
				ec2: &mockEC2{
					RespDescImages:    ec2.DescribeImagesOutput{},
					RespDescImagesErr: nil,

					RespDescInstances:    ec2.DescribeInstancesOutput{},
					RespDescInstancesErr: nil,

					RespDeregisterImage:    ec2.DeregisterImageOutput{},
					RespDeregisterImageErr: nil,

					RespDeleteSnapshot:    ec2.DeleteSnapshotOutput{},
					RespDeleteSnapshotErr: nil,
				},
				filterErr: true,
			},
			wantIDs: nil,
			wantErr: ErrFilterAMIs,
		},
		{
			name: "error describe instances",
			give: &AWS{
				ec2: &mockEC2{
					RespDescImages:    ec2.DescribeImagesOutput{},
					RespDescImagesErr: nil,

					RespDescInstances:    ec2.DescribeInstancesOutput{},
					RespDescInstancesErr: fmt.Errorf("FAIL"),

					RespDeregisterImage:    ec2.DeregisterImageOutput{},
					RespDeregisterImageErr: nil,

					RespDeleteSnapshot:    ec2.DeleteSnapshotOutput{},
					RespDeleteSnapshotErr: nil,
				},
			},
			wantIDs: nil,
			wantErr: ErrDesribeInstances,
		},
		{
			name: "error deregister image",
			give: &AWS{
				ec2: &mockEC2{
					RespDescImages: ec2.DescribeImagesOutput{
						Images: []types.Image{
							{ImageId: aws.String("ami-123")},
							{ImageId: aws.String("ami-456")},
						},
					},
					RespDescImagesErr: nil,

					RespDescInstances:    ec2.DescribeInstancesOutput{},
					RespDescInstancesErr: nil,

					RespDeregisterImage:    ec2.DeregisterImageOutput{},
					RespDeregisterImageErr: fmt.Errorf("FAIL"),

					RespDeleteSnapshot:    ec2.DeleteSnapshotOutput{},
					RespDeleteSnapshotErr: nil,
				},
				cfg: &Config{DryRun: false},
			},
			wantIDs: nil,
			wantErr: &ErrDeleteAMIs{
				IDs: []string{"ami-123", "ami-456"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ids, err := tt.give.DeleteUnusedAMIs()

			var eda *ErrDeleteAMIs

			switch {
			case tt.wantErr == nil:
				assert.Nil(t, err)
			case errors.As(err, &eda):
				assert.Equal(t, tt.wantErr, err)
			default:
				assert.True(t, errors.Is(err, tt.wantErr), fmt.Sprintf("expected: %s\ngot: %s", tt.wantErr, err))
			}

			assert.Equal(t, tt.wantIDs, ids)
		})
	}
}
