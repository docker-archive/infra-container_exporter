package main

// Container manager interface.
type Manager interface {
	// List containers on system
	Containers() ([]*container, error)

	// Get cgroup parent
	Parent() string
}

type container struct {
	Name  string
	ID    string
	Image string
}
