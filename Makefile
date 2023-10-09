BINARY := chinadns
MODULE := github.com/0990/chinadns

BUILD_DIR     := bin
BUILD_PACKAGE = $(MODULE)/cmd/server
BUILD_TAGS    :=
BUILD_FLAGS   := -v
BUILD_COMMIT  := $(shell git rev-parse --short HEAD)
BUILD_VERSION := $(shell git describe --abbrev=0 --tags HEAD)

LDFLAGS += -w -s -buildid=
LDFLAGS += -X "$(MODULE)/internal/version.Version=$(BUILD_VERSION)"
LDFLAGS += -X "$(MODULE)/internal/version.GitCommit=$(BUILD_COMMIT)"

GO_BUILD = go build  $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(BUILD_TAGS)' -trimpath

all: linux windows linux-arm64

linux:
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)  $(BUILD_PACKAGE)

windows:
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY).exe $(BUILD_PACKAGE)

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@