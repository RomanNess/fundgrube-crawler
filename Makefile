PI_BINARY?=fundgrube-crawler-pi
PI_SSH_USER_AND_HOST?=pi@pi.local
PI_DEPLOYMENT_PATH?=/home/pi/projects/fundgrube
PI_SSH_PROFILE?=pi

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
	go build -o bin/fundgrube-migrate cmd/fundgrube-migrate/main.go

build-pi:
	GOOS=linux GOARCH=arm GOARM=6 go build -o bin_pi/$(PI_BINARY) cmd/fundgrube-crawler/main.go

deploy-pi: build-pi
	scp bin_pi/* $(PI_SSH_USER_AND_HOST):$(PI_DEPLOYMENT_PATH)

run: build
	./bin/fundgrube-crawler

## migrate mongodb schema or clean up entries
migrate: build
	./bin/fundgrube-migrate

## export necessary env vars in env.sh before running
run-pi:
	ssh $(PI_SSH_PROFILE) bash -c "'cd $(PI_DEPLOYMENT_PATH) && source env.sh && ./fundgrube-crawler-pi'"

crontab-pi:
	ssh $(PI_SSH_PROFILE) bash -c 'echo; echo "10 * * * * . $(PI_DEPLOYMENT_PATH)/env.sh && $(PI_DEPLOYMENT_PATH)/$(PI_BINARY)" | crontab -'
