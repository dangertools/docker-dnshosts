FROM gliderlabs/alpine:3.1
MAINTAINER Andrew Dunham <andrew@du.nham.ca>

ADD docker-hosts /usr/local/bin/

## should *not* run as root, but needs access to /var/run/docker.sock, which
## should *not* be accessible by nobody. *sigh*
#USER nobody
ENTRYPOINT ["/usr/local/bin/docker-hosts"]
