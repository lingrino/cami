package cami

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
)

var _ ec2If = (*mockEC2)(nil)

type mockEC2 struct {
	RespDescImages    ec2.DescribeImagesOutput
	RespDescImagesErr error

	RespDescInstances    ec2.DescribeInstancesOutput
	RespDescInstancesErr error

	RespDeregisterImage    ec2.DeregisterImageOutput
	RespDeregisterImageErr error

	RespDeleteSnapshot    ec2.DeleteSnapshotOutput
	RespDeleteSnapshotErr error
}

func (m mockEC2) DescribeImages(ctx context.Context, in *ec2.DescribeImagesInput, opts ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	return &m.RespDescImages, m.RespDescImagesErr
}

//nolint:lll
func (m mockEC2) DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &m.RespDescInstances, m.RespDescInstancesErr
}

func (m mockEC2) DeregisterImage(context.Context, *ec2.DeregisterImageInput, ...func(*ec2.Options)) (*ec2.DeregisterImageOutput, error) {
	return &m.RespDeregisterImage, m.RespDeregisterImageErr
}

func (m mockEC2) DeleteSnapshot(context.Context, *ec2.DeleteSnapshotInput, ...func(*ec2.Options)) (*ec2.DeleteSnapshotOutput, error) {
	return &m.RespDeleteSnapshot, m.RespDeleteSnapshotErr
}

type mockErr struct {
	error

	ErrCode string
}

func (m mockErr) Error() string {
	return "FAIL"
}

func (m mockErr) ErrorCode() string {
	return m.ErrCode
}

func (m mockErr) ErrorMessage() string {
	return "FAIL"
}

func (m mockErr) ErrorFault() smithy.ErrorFault {
	return 0
}
