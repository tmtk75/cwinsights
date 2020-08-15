VERSION := $(shell git describe --tags --abbrev=0)
COMMIT:= $(shell git rev-parse HEAD)
VAR_VERSION := main.Version
VAR_COMMIT:= main.Commit

LDFLAGS := -ldflags "-X $(VAR_VERSION)=$(VERSION) \
	-X $(VAR_COMMIT)=$(COMMIT)"

cwinsights: *.go
	go build $(LDFLAGS) -o cwinsights .

install:
	go install

clean:
	rm -rf cwinsights
