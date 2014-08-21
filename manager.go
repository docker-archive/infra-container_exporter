package main

import (
	"github.com/docker/libcontainer/cgroups"
)

// Container manager interface.
type Manager interface {
	// List containers on system
	Containers() ([]*container, error)
}

type container struct {
	Name    string
	ID      string
	Cgroups *cgroups.Cgroup
}
