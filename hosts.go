package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

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

	hosts.entries = make(map[string]HostEntry)
	hosts.builtin = make(map[string]HostEntry)

	// combination of docker, centos
	hosts.builtin["__localhost4"] = HostEntry{
		IPAddress:         "127.0.0.1",
		CanonicalHostname: "localhost",
		Aliases:           []string{"localhost4"},
	}

	hosts.builtin["__localhost6"] = HostEntry{
		IPAddress:         "::1",
		CanonicalHostname: "localhost",
		Aliases:           []string{"localhost6", "ip6-localhost", "ip6-loopback"},
	}

	// docker puts these in
	hosts.builtin["fe00::0"] = HostEntry{"fe00::0", "ip6-localnet", nil}
	hosts.builtin["ff00::0"] = HostEntry{"ff00::0", "ip6-mcastprefix", nil}
	hosts.builtin["ff02::1"] = HostEntry{"ff02::1", "ip6-allnodes", nil}
	hosts.builtin["ff02::2"] = HostEntry{"ff02::2", "ip6-allrouters", nil}

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
# Hosts file created by docker-hosts
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

func (h *Hosts) Add(containerId string) {
	h.Lock()
	defer h.Unlock()

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

	h.WriteFile()
}

func (h *Hosts) Remove(containerId string) {
	h.Lock()
	defer h.Unlock()

	delete(h.entries, containerId)

	h.WriteFile()
}
