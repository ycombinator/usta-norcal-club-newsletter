.PHONY: run build dist

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

run:
	go run main.go $(if $(ORG_ID),-org=$(ORG_ID)) $(if $(TEAMS),-teams=$(TEAMS)) $(if $(FORMAT),-format=$(FORMAT))

build:
	go build .

dist: build
	mkdir -p dist/$(GOOS)/$(GOARCH)
	mv ./usta-norcal-club-newsletter dist/$(GOOS)/$(GOARCH)/
