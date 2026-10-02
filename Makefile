GO ?= GO111MODULE=on CGO_ENABLED=0 go


test:
	go tool gotestsum --format-hide-empty-pkg -- ./... --race

install:
	$(GO) install u.go
	$(GO) run . install
	u ss
	u ss
	u ss
	exec zsh
