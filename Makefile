NAME=docker-hosts

SOURCES=main.go hosts.go

.PHONY: all build docker clean

all: build

build: $(NAME)

$(NAME): $(SOURCES)
	godep go build -v -o $@ ./...

docker: build
	docker build --tag=andrewd/$(NAME) .
	docker push andrewd/$(NAME)

clean:
	rm -rf build
