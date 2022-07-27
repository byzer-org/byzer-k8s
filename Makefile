export GO111MODULE=on

all: release

REVISION := $(shell git rev-parse --short HEAD 2>/dev/null)
REVISIONDATE := $(shell git log -1 --pretty=format:'%ad' --date short 2>/dev/null)
PKG := mlsql.tech/allwefantasy/deploy/pkg/version
LDFLAGS = -s -w
ifneq ($(strip $(REVISION)),) # Use git clone
	LDFLAGS += -X $(PKG).revision=$(REVISION) \
		   -X $(PKG).revisionDate=$(REVISIONDATE)
endif

SHELL = /bin/sh

ifdef STATIC
	LDFLAGS += -linkmode external -extldflags '-static'
	CC = /usr/bin/musl-gcc
	export CC
endif


release: linux mac

linux: Makefile cmd/*.go pkg/*/*.go
	env GOOS=linux GOARCH=amd64  go build -ldflags="$(LDFLAGS)"  -o byzerup-linux-amd64 ./cmd

mac:
	env GOOS=darwin GOARCH=amd64  go build -ldflags="$(LDFLAGS)"  -o byzerup-darwin-amd64 ./cmd