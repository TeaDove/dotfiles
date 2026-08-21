.PHONY: ckeck install upload cbox-build

PKG_VERSION ?= $(shell cat VERSION)
PKG_OUTPUT ?= build/u
GO ?= GO111MODULE=on CGO_ENABLED=0 go
GOOS ?= $(shell $(GO) version | cut -d' ' -f4 | cut -d'/' -f1)
GOARCH ?= $(shell $(GO) version | cut -d' ' -f4 | cut -d'/' -f2)


test:
	go tool gotestsum --format-hide-empty-pkg -- ./... --race

install:
	$(GO) install u.go
	u install

cbox-build:
	docker build -t claude-box devcontainer
