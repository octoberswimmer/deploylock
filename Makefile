VERSION=$(shell git describe --abbrev=0 --always)
LDFLAGS = -ldflags "-X github.com/octoberswimmer/deploylock.Version=${VERSION}"
GCFLAGS = -gcflags="all=-N -l"
EXECUTABLE=deploylock-client
PACKAGE=./cmd/deploylock-client
WINDOWS=$(EXECUTABLE)-windows-amd64.exe
LINUX=$(EXECUTABLE)-linux-amd64
OSX_AMD64=$(EXECUTABLE)-darwin-amd64
OSX_ARM64=$(EXECUTABLE)-darwin-arm64
ALL=$(WINDOWS) $(LINUX) $(OSX_ARM64) $(OSX_AMD64)

default:
	go build -o ${EXECUTABLE} ${LDFLAGS} ${PACKAGE}

install:
	go install ${LDFLAGS} ${PACKAGE}

install-debug:
	go install ${LDFLAGS} ${GCFLAGS} ${PACKAGE}

$(WINDOWS):
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -v -o $(WINDOWS) ${LDFLAGS} ${PACKAGE}

$(LINUX):
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(LINUX) ${LDFLAGS} ${PACKAGE}

$(OSX_AMD64):
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -v -o $(OSX_AMD64) ${LDFLAGS} ${PACKAGE}

$(OSX_ARM64):
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -v -o $(OSX_ARM64) ${LDFLAGS} ${PACKAGE}

$(basename $(WINDOWS)).zip: $(WINDOWS)
	zip $@ $<
	7za rn $@ $< $(EXECUTABLE)$(suffix $<)

%.zip: %
	zip $@ $<
	7za rn $@ $< $(EXECUTABLE)

dist: test $(addsuffix .zip,$(basename $(ALL)))

test:
	test -z "$(go fmt)"
	go vet ./...
	go test ./...
	go test -race ./...

docs:
	go run docs/mkdocs.go

checkcmd-%:
	@hash $(*) > /dev/null 2>&1 || \
		(echo "ERROR: '$(*)' must be installed and available on your PATH."; exit 1)

clean:
	-rm -f $(EXECUTABLE) $(EXECUTABLE)-*

.PHONY: default dist clean docs
