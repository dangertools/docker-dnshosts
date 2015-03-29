FROM andrew-d/docker-hosts
MAINTAINER Georg Schild <dangertools@gmail.com>

ADD docker-dnshosts /usr/local/bin/

## should *not* run as root, but needs access to /var/run/docker.sock, which
## should *not* be accessible by nobody. *sigh*
#USER nobody
ENTRYPOINT ["/usr/local/bin/docker-dnshosts"]
