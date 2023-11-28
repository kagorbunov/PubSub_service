GOPATH:=$(shell go env GOPATH)
GOBIN:=$(shell pwd)/bin

.PHONY: run stop

run:
	docker-compose up -d --build

stop:
	docker-compose down