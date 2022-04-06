export GO111MODULE=on

all: byzer-k8s-deploy

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

byzer-k8s-deploy: Makefile cmd/*.go pkg/*/*.go
	go build -ldflags="$(LDFLAGS)"  -o byzer-k8s-deploy ./cmd
