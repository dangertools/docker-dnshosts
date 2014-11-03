package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	dockerapi "github.com/fsouza/go-dockerclient"
	flags "github.com/jessevdk/go-flags"
)

func getopt(name, def string) string {
	if env := os.Getenv(name); env != "" {
		return env
	}

	return def
}

type Options struct {
	DomainName string `short:"d" long:"domain-name" description:"domain to append"`
	Verbose    bool   `short:"v" long:"verbose" description:"be more verbose"`
	File       struct {
		Filename string
	} `positional-args:"true" required:"true" description:"the hosts file to write"`
}

func main() {
	var opts Options

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	docker, err := dockerapi.NewClient(getopt("DOCKER_HOST", "unix:///var/run/docker.sock"))
	if err != nil {
		log.WithField("err", err).Fatal("could not connect to docker")
	}

	hosts := NewHosts(docker, opts.File.Filename, opts.DomainName)

	// set up to handle events early, so we don't miss anything while doing the
	// initial population
	events := make(chan *dockerapi.APIEvents)
	err = docker.AddEventListener(events)
	if err != nil {
		log.WithField("err", err).Fatal("could not add event listener")
	}

	containers, err := docker.ListContainers(dockerapi.ListContainersOptions{})
	if err != nil {
		log.WithField("err", err).Fatal("could not list containers")
	}

	for _, listing := range containers {
		log.WithField("id", listing.ID).Debug("adding existing container")
		go hosts.Add(listing.ID)
	}

	log.Infoln("listening for Docker events...")
	for msg := range events {
		switch msg.Status {
		case "start":
			log.WithField("id", msg.ID).Debug("adding new container")
			go hosts.Add(msg.ID)

		case "die":
			log.WithField("id", msg.ID).Debug("removing dead container")
			go hosts.Remove(msg.ID)
		}
	}

	log.Fatal("docker event loop closed") // todo: reconnect?
}
