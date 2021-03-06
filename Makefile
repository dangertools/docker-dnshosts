NAME=docker-dnshosts

SOURCES=main.go hosts.go

.PHONY: all build docker clean

all: build

build: $(NAME)

$(NAME): $(SOURCES)
	CGO_ENABLED=0 godep go build -a -v -tags netgo -installsuffix netgo -o $@ ./...
	strip $@

docker: build
	docker build --tag=dangertools/$(NAME) .
	docker push dangertools/$(NAME)

clean:
	rm -rf $(NAME)
