VERSION = 1.0
COMMIT = $(shell git describe --always)
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

default: build

deps:
	go mod vendor

# build generate binary on './bin' directory.
build: deps
	go build -ldflags "-X main.Version=$(VERSION) -w -s" -o bin/prometheus-nacos-sd .

buildx: deps
	GOGS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -w -s" -o "bin/prometheus-nacos-sd_linux_amd64_v$(VERSION)"  .
	upx bin/prometheus-nacos-sd_linux_amd64_v$(VERSION) || true

lint:
	golint ${GOFILES_NOVENDOR}

vet:
	go vet -v ${GOFILES_NOVENDOR}

test:
	go test -v ${GOFILES_NOVENDOR}

fmt:
	gofmt -l -w ${GOFILES_NOVENDOR}

release: buildx
	git tag v$(VERSION)
	git push origin v$(VERSION)
	ghr -u afghanistanyn v$(VERSION) bin/prometheus-nacos-sd_linux_amd64_v$(VERSION)