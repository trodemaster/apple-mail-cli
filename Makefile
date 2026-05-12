BINARY_MAIL := amail
BINARY_ICAL := aical
GOBIN       := $(shell go env GOPATH)/bin
ENTITLE     := entitlements.plist
PLUGIN_MAIL := plugins/apple-mail/bin
PLUGIN_ICAL := plugins/ical/bin

.PHONY: all build sign install skill skill-mail skill-ical clean fmt vet

## all: build, sign, and install both binaries to GOPATH/bin (default)
all: build sign install

## build: compile amail (current platform) and aical
build:
	go build -o $(BINARY_MAIL) .
	go build -o $(BINARY_ICAL) ./ical/

## sign: ad-hoc codesign both local binaries (required for automation TCC)
sign: build
	codesign --sign - --force --entitlements $(ENTITLE) ./$(BINARY_MAIL)
	codesign --sign - --force --entitlements $(ENTITLE) ./$(BINARY_ICAL)

## install: copy signed binaries to GOPATH/bin and re-sign in place
install: sign
	cp ./$(BINARY_MAIL) $(GOBIN)/$(BINARY_MAIL)
	codesign --sign - --force --entitlements $(ENTITLE) $(GOBIN)/$(BINARY_MAIL)
	cp ./$(BINARY_ICAL) $(GOBIN)/$(BINARY_ICAL)
	codesign --sign - --force --entitlements $(ENTITLE) $(GOBIN)/$(BINARY_ICAL)

## skill: build and sign both plugin binaries
skill: skill-mail skill-ical

## skill-mail: cross-compile universal amail binary into plugins/apple-mail/bin/ and sign it
skill-mail:
	mkdir -p $(PLUGIN_MAIL)
	GOOS=darwin GOARCH=arm64 go build -o $(PLUGIN_MAIL)/$(BINARY_MAIL)_arm64 .
	GOOS=darwin GOARCH=amd64 go build -o $(PLUGIN_MAIL)/$(BINARY_MAIL)_amd64 .
	lipo -create -output $(PLUGIN_MAIL)/$(BINARY_MAIL) \
		$(PLUGIN_MAIL)/$(BINARY_MAIL)_arm64 \
		$(PLUGIN_MAIL)/$(BINARY_MAIL)_amd64
	rm $(PLUGIN_MAIL)/$(BINARY_MAIL)_arm64 $(PLUGIN_MAIL)/$(BINARY_MAIL)_amd64
	cp $(ENTITLE) $(PLUGIN_MAIL)/$(ENTITLE)
	codesign --sign - --force --entitlements $(ENTITLE) $(PLUGIN_MAIL)/$(BINARY_MAIL)

## skill-ical: cross-compile universal aical binary into plugins/ical/bin/ and sign it
skill-ical:
	mkdir -p $(PLUGIN_ICAL)
	GOOS=darwin GOARCH=arm64 go build -o $(PLUGIN_ICAL)/$(BINARY_ICAL)_arm64 ./ical/
	GOOS=darwin GOARCH=amd64 go build -o $(PLUGIN_ICAL)/$(BINARY_ICAL)_amd64 ./ical/
	lipo -create -output $(PLUGIN_ICAL)/$(BINARY_ICAL) \
		$(PLUGIN_ICAL)/$(BINARY_ICAL)_arm64 \
		$(PLUGIN_ICAL)/$(BINARY_ICAL)_amd64
	rm $(PLUGIN_ICAL)/$(BINARY_ICAL)_arm64 $(PLUGIN_ICAL)/$(BINARY_ICAL)_amd64
	cp $(ENTITLE) $(PLUGIN_ICAL)/$(ENTITLE)
	codesign --sign - --force --entitlements $(ENTITLE) $(PLUGIN_ICAL)/$(BINARY_ICAL)

## clean: remove local build artifacts and plugin binaries
clean:
	rm -f ./$(BINARY_MAIL) ./$(BINARY_ICAL)
	rm -f $(PLUGIN_MAIL)/$(BINARY_MAIL) $(PLUGIN_MAIL)/$(ENTITLE)
	rm -f $(PLUGIN_ICAL)/$(BINARY_ICAL) $(PLUGIN_ICAL)/$(ENTITLE)

## fmt: format all Go source files
fmt:
	go fmt ./...

## vet: run go vet
vet:
	go vet ./...
