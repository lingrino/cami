package cami

import (
	"errors"
)

var (
	// ErrCreateSession is when we fail to create an AWS session
	ErrCreateSession = errors.New("create session")
	// ErrDesribeImages is when we fail to describe EC2 images
	ErrDesribeImages = errors.New("describe images")
	// ErrDesribeInstances is when we fail to describe EC2 instances
	ErrDesribeInstances = errors.New("describe instances")
	// ErrDeregisterImage is when we fail to deregister an image (AMI)
	ErrDeregisterImage = errors.New("deregister image")
	// ErrDeleteSnapshot is when we fail to delete a snapshot
	ErrDeleteSnapshot = errors.New("delete snapshot")
	// ErrFilterAMIs is when when we fail to filter AMIs and EC2 instances
	ErrFilterAMIs = errors.New("filter AMIs")
)

// ErrDeleteAMIs is when we fail to delete (deregister image + snapshot delete) an image (AMI)
type ErrDeleteAMIs struct {
	// IDs is the list of AMI and snapshot IDs we failed to delete
	IDs []string
}

// Error returns the error string for ErrDeleteAMIs
func (e *ErrDeleteAMIs) Error() string {
	return "delete AMIs"
}

// Append adds a new ID to the list of IDs
func (e *ErrDeleteAMIs) Append(ids ...string) {
	if e.IDs == nil {
		e.IDs = []string{}
	}
	e.IDs = append(e.IDs, ids...)
}

// ErrorOrNil returns nil if IDs is empty and the error otherwise
func (e *ErrDeleteAMIs) ErrorOrNil() error {
	if e == nil || len(e.IDs) == 0 {
		return nil
	}
	return e
}
