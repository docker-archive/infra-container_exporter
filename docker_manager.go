package main

import (
	"github.com/docker/libcontainer/cgroups"
	"github.com/fsouza/go-dockerclient"
)

const parentCgroup = "/"

type dockerManager struct {
	addr string
}

func newDockerManager(addr string) *dockerManager {
	return &dockerManager{addr: addr}
}

// Return a list of all running containers on the system
func (m *dockerManager) Containers() ([]*container, error) {
	client, err := docker.NewClient(m.addr)
	if err != nil {
		return nil, err
	}

	// Get all *running* containers
	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		return nil, err
	}

	cl := []*container{}
	for _, c := range containers {
		cl = append(cl, &container{
			ID:   c.ID,
			Name: c.Names[0][1:], // FIXME: This isn't a very good solution but the best I could think of.
			Cgroups: &cgroups.Cgroup{
				Name:   c.ID,
				Parent: parentCgroup,
			},
		})
	}
	return cl, nil
}
