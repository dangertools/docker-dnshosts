package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	dockerapi "github.com/fsouza/go-dockerclient"
)

type HostEntry struct {
	IPAddress         string
	CanonicalHostname string
	Aliases           []string
}

type Hosts struct {
	sync.Mutex
	docker  *dockerapi.Client
	path    string
	domain  string
	entries map[string]HostEntry
	builtin map[string]HostEntry
}

func NewHosts(docker *dockerapi.Client, path, domain string) *Hosts {
	hosts := &Hosts{
		docker: docker,
		path:   path,
		domain: domain,
	}

	FullUpdate(hosts)

	return hosts
}

func (h *Hosts) WriteFile() {
	file, err := os.Create(h.path)

	if err != nil {
		log.WithFields(log.Fields{
			"path": h.path,
			"err":  err,
		}).Error("unable to write hosts file")
		return
	}

	defer file.Close()

	// Write the current time
	currTime := time.Now().UTC().Truncate(time.Millisecond)
	headerText := fmt.Sprintf(`
# Hosts file created by docker-dnshosts
# Number of entries: %d
# Last updated at: %s


# Running containers
`, len(h.entries), currTime)

	file.WriteString(strings.TrimLeft(headerText, "\r\n "))

	for _, entry := range h.entries {
		// <ip>\t<canonical>\t<alias1>\t…\t<aliasN>\n
		file.WriteString(strings.Join(
			append(
				[]string{entry.IPAddress, entry.CanonicalHostname},
				entry.Aliases...,
			),
			"\t",
		) + "\n")
	}

	file.WriteString("\n# Built-in entries\n")

	for _, entry := range h.builtin {
		// <ip>\t<canonical>\t<alias1>\t…\t<aliasN>\n
		file.WriteString(strings.Join(
			append(
				[]string{entry.IPAddress, entry.CanonicalHostname},
				entry.Aliases...,
			),
			"\t",
		) + "\n")
	}
}

func (h *Hosts) ReloadConfiguration() {
    cmd := exec.Command("pkill", "-x", "-HUP", "dnsmasq")
    cmd.Run()
}

func AddContainerEntry(h *Hosts, containerId string) {
	container, err := h.docker.InspectContainer(containerId)
	if err != nil {
		log.WithFields(log.Fields{
			"containerId": containerId,
			"err":         err,
		}).Error("unable to inspect container:")
		return
	}

	entry := HostEntry{
		IPAddress:         container.NetworkSettings.IPAddress,
		CanonicalHostname: container.Config.Hostname,
		Aliases:           []string{
		// container.Name[1:], // could contain "_"
		},
	}

	if h.domain != "" {
		entry.Aliases =
			append(h.entries[containerId].Aliases, container.Config.Hostname+"."+h.domain)
	}

	h.entries[containerId] = entry
}

func (h *Hosts) Add(containerId string) {
	h.Lock()
	defer h.Unlock()

	AddContainerEntry(h, containerId)

	h.WriteFile()
	h.ReloadConfiguration()
}

func (h *Hosts) Remove(containerId string) {
	h.Lock()
	defer h.Unlock()

	delete(h.entries, containerId)

	h.WriteFile()
	h.ReloadConfiguration()
}

func FullUpdate(h *Hosts) {
    h.Lock()
    defer h.Unlock()
    
    containers, err := h.docker.ListContainers(dockerapi.ListContainersOptions{})
    if err != nil {
        log.Println("unable to list containers:", err)
        return
    }
    h.entries = make(map[string]HostEntry)

    for _, container := range containers {
	AddContainerEntry(h, container.ID)
    }
    h.WriteFile()
    h.ReloadConfiguration()
}
