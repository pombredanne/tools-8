.DEFAULT_GOAL := help

help:
	@echo "make <[version|buildver]|semver>: Update and commit version.go in ./config"

VERSION=0.0.0
STAMP=`date -u '+%e %b %Y %R %Z'`
HASH=`git rev-parse --short HEAD`

.PHONY: buildver
buildver:
	sed -i "s/\(	buildhash.*\)/	buildhash = \"$(HASH)\"/g" config/version.go
	sed -i "s/\(	buildtime.*\)/	buildtime = \"$(STAMP)\"/g" config/version.go

.PHONY: version
version: buildver

.PHONY: clean
clean:
	@rm ogload

.PHONY: build
build: buildver
	go build -v

ogload: buildver
	go build -v

install: buildver
	go install -v

.PHONY: semver
semver:
	@echo "Increment this yourself, dang."
