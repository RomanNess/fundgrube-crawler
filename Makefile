PI_BINARY?=bin/fundgrube-crawler-pi
PI_SSH_USER_AND_HOST?=pi@pi.local
PI_DEPLOYMENT_PATH?=/home/pi/projects/fundgrube

all: test vet fmt build

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l
	test -z $$(go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l)

build:
	go build -o bin/fundgrube-crawler cmd/fundgrube-crawler/main.go

build-pi:
	GOOS=linux GOARCH=arm GOARM=5 go build -o $(PI_BINARY) cmd/fundgrube-crawler/main.go

deploy-pi: build-pi
	scp $(PI_BINARY) $(PI_SSH_USER_AND_HOST):$(PI_DEPLOYMENT_PATH)

run: build
	./bin/fundgrube-crawler
