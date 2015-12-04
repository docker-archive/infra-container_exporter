package main

import (
	"log"

	"github.com/fsouza/go-dockerclient"
)

type dockerManager struct {
	addr   string
	parent string
	client *docker.Client
}

func newDockerManager(addr, parent string) *dockerManager {
	client, err := docker.NewClient(addr)
	if err != nil {
		log.Fatalf("Unable to start docker client %v", err.Error())
	}
	return &dockerManager{addr: addr, parent: parent, client: client}
}

// Return a list of all running containers on the system
func (m *dockerManager) Containers() ([]*container, error) {

	// Get all *running* containers
	containers, err := m.client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		return nil, err
	}

	cl := []*container{}
	for _, c := range containers {
		cl = append(cl, &container{
			ID:    c.ID,
			Name:  c.Names[0][1:], // FIXME: This isn't a very good solution but the best I could think of.
			Image: c.Image,
		})
	}
	return cl, nil
}

func (m *dockerManager) Parent() string {
	return m.parent
}
