BINARY     := amail
GOBIN      := $(shell go env GOPATH)/bin
ENTITLE    := entitlements.plist

.PHONY: all build sign install clean fmt vet

## all: build, sign, and install to GOPATH/bin (default)
all: build sign install

## build: compile for the current platform
build:
	go build -o $(BINARY) .

## sign: ad-hoc codesign the local binary (required for Mail.app automation TCC)
sign: build
	codesign --sign - --force --entitlements $(ENTITLE) ./$(BINARY)

## install: copy signed binary to GOPATH/bin and re-sign in place
install: sign
	cp ./$(BINARY) $(GOBIN)/$(BINARY)
	codesign --sign - --force --entitlements $(ENTITLE) $(GOBIN)/$(BINARY)

## clean: remove the local build artifact
clean:
	rm -f ./$(BINARY)

## fmt: format all Go source files
fmt:
	go fmt ./...

## vet: run go vet
vet:
	go vet ./...
