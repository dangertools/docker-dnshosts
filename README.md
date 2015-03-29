# Simplified Docker container hostname resolution

`docker-dnshosts` (a fork of docker-hosts with a still terrible name) maintains a file in the format of
`/etc/hosts` that contains IP addresses and hostnames of Docker containers.
Contrary to the original solution in docker-hosts, mounting a file is not the way
to go with docker-dnshosts (although the same file is generated). The idea behind docker-dnshosts
is to feed dnsmasq with the hosts information. 
This way hostname resolution works as expected (if the containers are provided with a dnsmasq dns
server) without loosing the features given by docker to the /etc/hosts file.

The solution was inspired by [Amartynov's dnsmasq and docker solution](https://blog.amartynov.ru/archives/dnsmasq-docker-service-discovery/),
yet I didn't like the fact to always manually call a script after changing containers. This solution
builds on the same principle but reacts to docker events and HUPs dnsmasq automatically after finding a change.


## building

This project uses [godep][godep]. Godep must be available on your
path.

    make

## running

Start the `docker-dnshosts` process and give it the path to the container hosts file,
which is bound to dnsmasq:

    docker-host /path/to/hosts

Optionally specify `DOCKER_HOST` environment variable.

Then start a container:

    docker run -i -t --dns=your.dns.server centos /bin/bash

Within the `centos` container, you can just ping any other container running on the same
docker host.

## running in Docker

Create an empty file at `/var/lib/docker/hosts`, make it mode `0644` and owned
by `nobody:nobody`.

    docker run \
        -d \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v /path/to/hosts:/srv/hosts \
        blalor/docker-dnshosts --domain-name=dev.docker /srv/hosts

[godep]: https://github.com/tools/godep
