package cami

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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

			assert.Equal(t, tt.wantAWS, aws)
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
				newSessionFn: func(cfgs ...*aws.Config) (*session.Session, error) { return nil, fmt.Errorf("FAIL") },
			},
			wantAWS: &AWS{},
			wantErr: ErrCreateSession,
		},
		{
			name: "valid",
			give: &AWS{
				newEC2Fn:     func(p client.ConfigProvider, cfgs ...*aws.Config) *ec2.EC2 { return &ec2.EC2{} },
				newSessionFn: func(cfgs ...*aws.Config) (*session.Session, error) { return &session.Session{}, nil },
			},
			wantAWS: &AWS{ec2: &ec2.EC2{}},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
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
		wantAMIs   []*ec2.Image
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
				Images: []*ec2.Image{
					{ImageId: aws.String("ami-123")},
					{ImageId: aws.String("ami-456")},
				},
			},
			giveErr: nil,
			wantAMIs: []*ec2.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			wantErr: nil,
		},
		{
			name: "images with error",
			giveOutput: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
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
		giveImages []*ec2.Image
		giveOutput ec2.DescribeInstancesOutput
		giveErr    error
		givePages  int
		wantEC2s   []*ec2.Instance
		wantErr    error
	}{
		{
			name:       "empty",
			giveImages: []*ec2.Image{},
			giveOutput: ec2.DescribeInstancesOutput{},
			giveErr:    nil,
			givePages:  0,
			wantEC2s:   nil,
			wantErr:    nil,
		},
		{
			name:       "error",
			giveImages: []*ec2.Image{},
			giveOutput: ec2.DescribeInstancesOutput{},
			giveErr:    fmt.Errorf("FAIL"),
			givePages:  0,
			wantEC2s:   nil,
			wantErr:    ErrDesribeInstances,
		},
		{
			name: "one reservation",
			giveImages: []*ec2.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveOutput: ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{{Instances: []*ec2.Instance{
					{ImageId: aws.String("ami-123")},
					{ImageId: aws.String("ami-456")},
				}}},
			},
			giveErr:   nil,
			givePages: 1,
			wantEC2s: []*ec2.Instance{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			wantErr: nil,
		},
		{
			name: "two reservations",
			giveImages: []*ec2.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveOutput: ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{Instances: []*ec2.Instance{{ImageId: aws.String("ami-123")}}},
					{Instances: []*ec2.Instance{{ImageId: aws.String("ami-456")}}},
				},
			},
			giveErr:   nil,
			givePages: 1,
			wantEC2s: []*ec2.Instance{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			wantErr: nil,
		},
		{
			name: "two pages",
			giveImages: []*ec2.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveOutput: ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{Instances: []*ec2.Instance{{ImageId: aws.String("ami-123")}}},
					{Instances: []*ec2.Instance{{ImageId: aws.String("ami-456")}}},
				},
			},
			giveErr:   nil,
			givePages: 2,
			wantEC2s: []*ec2.Instance{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
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
					RespDescInstances:      tt.giveOutput,
					RespDescInstancesErr:   tt.giveErr,
					RespDescInstancesPages: tt.givePages,
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
		giveAMIs []*ec2.Image
		giveEC2s []*ec2.Instance
		wantAMIs []*ec2.Image
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
			giveAMIs: []*ec2.Image{},
			giveEC2s: []*ec2.Instance{},
			wantAMIs: nil,
			wantErr:  nil,
		},
		{
			name: "filter",
			giveAMIs: []*ec2.Image{
				{ImageId: aws.String("ami-123")},
				{ImageId: aws.String("ami-456")},
			},
			giveEC2s: []*ec2.Instance{
				{ImageId: aws.String("ami-456")},
				{ImageId: aws.String("ami-789")},
			},
			wantAMIs: []*ec2.Image{
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
		giveAMIs               []*ec2.Image
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
			giveAMIs:               []*ec2.Image{},
			giveDeregisterImage:    ec2.DeregisterImageOutput{},
			giveDeregisterImageErr: nil,
			giveDeleteSnapshot:     ec2.DeleteSnapshotOutput{},
			giveDeleteSnapshotErr:  nil,
			wantIDs:                nil,
			wantErr:                nil,
		},
		{
			name: "deregister image error",
			giveAMIs: []*ec2.Image{
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
			giveAMIs: []*ec2.Image{
				{
					ImageId: aws.String("ami-123"),
					BlockDeviceMappings: []*ec2.BlockDeviceMapping{
						{
							Ebs: &ec2.EbsBlockDevice{
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
			giveAMIs: []*ec2.Image{
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
			giveAMIs: []*ec2.Image{
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
			giveAMIs: []*ec2.Image{
				{
					ImageId: aws.String("ami-123"),
					BlockDeviceMappings: []*ec2.BlockDeviceMapping{
						{
							Ebs: &ec2.EbsBlockDevice{
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
			giveAMIs: []*ec2.Image{
				{
					ImageId: aws.String("ami-123"),
					BlockDeviceMappings: []*ec2.BlockDeviceMapping{
						{
							Ebs: &ec2.EbsBlockDevice{
								SnapshotId: aws.String("snap-123"),
							},
						},
					},
				},
				{
					ImageId: aws.String("ami-456"),
					BlockDeviceMappings: []*ec2.BlockDeviceMapping{
						{
							Ebs: &ec2.EbsBlockDevice{
								SnapshotId: aws.String("snap-456"),
							},
						},
						{
							Ebs: &ec2.EbsBlockDevice{
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

					RespDescInstances:      ec2.DescribeInstancesOutput{},
					RespDescInstancesErr:   nil,
					RespDescInstancesPages: 1,

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

					RespDescInstances:      ec2.DescribeInstancesOutput{},
					RespDescInstancesErr:   nil,
					RespDescInstancesPages: 1,

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

					RespDescInstances:      ec2.DescribeInstancesOutput{},
					RespDescInstancesErr:   nil,
					RespDescInstancesPages: 1,

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

					RespDescInstances:      ec2.DescribeInstancesOutput{},
					RespDescInstancesErr:   fmt.Errorf("FAIL"),
					RespDescInstancesPages: 1,

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
						Images: []*ec2.Image{
							{ImageId: aws.String("ami-123")},
							{ImageId: aws.String("ami-456")},
						},
					},
					RespDescImagesErr: nil,

					RespDescInstances:      ec2.DescribeInstancesOutput{},
					RespDescInstancesErr:   nil,
					RespDescInstancesPages: 1,

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
