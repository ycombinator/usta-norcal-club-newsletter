.PHONY: run build dist

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

run:
	go run main.go $(ORG_ID)

build:
	go build .

dist: build
	mkdir -p dist/$(GOOS)/$(GOARCH)
	mv ./usta-norcal-club-newsletter dist/$(GOOS)/$(GOARCH)/
