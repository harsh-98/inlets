Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"
PLATFORM := $(shell ./hack/platform-tag.sh)

.PHONY: all
all: docker

.PHONY: dist
dist:
	CGO_ENABLED=0 GOOS=linux go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/tunzalctl ./tunzalctl
	CGO_ENABLED=0 GOOS=darwin go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/tunzalctl-darwin ./tunzalctl
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/tunzalctl-armhf ./tunzalctl
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/tunzalctl-arm64 ./tunzalctl

.PHONY: docker
docker:
	docker build --build-arg Version=$(Version) --build-arg GIT_COMMIT=$(GitCommit) -t alexellis2/inlets:$(Version)$(PLATFORM) .

.PHONY: docker-login
docker-login:
	echo -n "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin

.PHONY: push
push:
	docker push alexellis2/inlets:$(Version)$(PLATFORM)
