VERSION := $(shell git describe --tags --abbrev=0)
COMMIT:= $(shell git rev-parse HEAD)
VAR_VERSION := main.Version
VAR_COMMIT:= main.Commit

LDFLAGS := -ldflags "-X $(VAR_VERSION)=$(VERSION) \
	-X $(VAR_COMMIT)=$(COMMIT)"

cwinsight: *.go
	go build $(LDFLAGS) -o cwinsight .

release: build
	ghr -u tmtk75 $(VERSION) ./build

build: build/sha256sum.txt

build/sha256sum.txt: build/cwinsight_darwin_amd64.zip build/cwinsight_linux_amd64.zip
	(cd build && sha256sum cwinsight_* > sha256sum.txt)

build/cwinsight_darwin_amd64.zip build/cwinsight_linux_amd64.zip: build/cwinsight_darwin_amd64 build/cwinsight_linux_amd64
	parallel '(cd build && zip -m cwinsight_{1}_amd64.zip cwinsight_{1}_amd64)' \
	  ::: darwin linux

build/cwinsight_darwin_amd64 build/cwinsight_linux_amd64: *.go
	parallel 'GOARCH=amd64 GOOS={1} go build $(LDFLAGS) -o build/cwinsight_{1}_amd64 .' \
	  ::: darwin linux

clean:
	rm -rf cwinsight

distclean: clean
	rm -rf build
