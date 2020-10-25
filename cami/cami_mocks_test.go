package cami

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type mockEC2 struct {
	ec2iface.EC2API

	RespDescImages    ec2.DescribeImagesOutput
	RespDescImagesErr error

	RespDescInstances      ec2.DescribeInstancesOutput
	RespDescInstancesErr   error
	RespDescInstancesPages int

	RespDeregisterImage    ec2.DeregisterImageOutput
	RespDeregisterImageErr error

	RespDeleteSnapshot    ec2.DeleteSnapshotOutput
	RespDeleteSnapshotErr error
}

func (m mockEC2) DescribeImages(*ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	return &m.RespDescImages, m.RespDescImagesErr
}

//nolint:lll
func (m mockEC2) DescribeInstancesPages(in *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	for i := 1; i <= m.RespDescInstancesPages; i++ {
		if i == m.RespDescInstancesPages {
			fn(&m.RespDescInstances, true)
		} else {
			fn(&m.RespDescInstances, false)
		}
	}

	return m.RespDescInstancesErr
}

func (m mockEC2) DeregisterImage(*ec2.DeregisterImageInput) (*ec2.DeregisterImageOutput, error) {
	return &m.RespDeregisterImage, m.RespDeregisterImageErr
}

func (m mockEC2) DeleteSnapshot(*ec2.DeleteSnapshotInput) (*ec2.DeleteSnapshotOutput, error) {
	return &m.RespDeleteSnapshot, m.RespDeleteSnapshotErr
}

type mockErr struct {
	error

	ErrCode string
}

func (m mockErr) Error() string {
	return "FAIL"
}

func (m mockErr) Code() string {
	return m.ErrCode
}

func (m mockErr) Message() string {
	return "FAIL"
}

func (m mockErr) OrigErr() error {
	return nil
}
