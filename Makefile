BINARY := chinadns
MODULE := github.com/0990/chinadns

BUILD_DIR     := build
BUILD_PACKAGE = $(MODULE)/cmd/server
BUILD_TAGS    :=
BUILD_FLAGS   := -v
BUILD_COMMIT  := $(shell git rev-parse --short HEAD)
BUILD_VERSION := $(shell git describe --abbrev=0 --tags HEAD)
CONFIG_DIR := packing
CONFIG_ZIP_NAME := config.zip

LDFLAGS += -w -s -buildid=
LDFLAGS += -X "$(MODULE)/internal/version.Version=$(BUILD_VERSION)"
LDFLAGS += -X "$(MODULE)/internal/version.GitCommit=$(BUILD_COMMIT)"

GO_BUILD = go build  $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(BUILD_TAGS)' -trimpath


UNIX_ARCH_LIST = \
	linux-amd64 \
	linux-arm64

WINDOWS_ARCH_LIST = \
	windows-amd64

unix_releases := $(addsuffix .zip, $(UNIX_ARCH_LIST))
windows_releases := $(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(unix_releases): %.zip: %
	@zip -qmj $(BUILD_DIR)/$(BINARY)-$(basename $@).zip $(BUILD_DIR)/$(BINARY)-$(basename $@)

$(windows_releases): %.zip: %
	@zip -qmj $(BUILD_DIR)/$(BINARY)-$(basename $@).zip $(BUILD_DIR)/$(BINARY)-$(basename $@).exe

releases: $(unix_releases) $(windows_releases) zip-config

all: linux-amd64 windows-amd64 linux-arm64

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@  $(BUILD_PACKAGE)

windows-amd64:
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@.exe $(BUILD_PACKAGE)

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@ $(BUILD_PACKAGE)

zip-config:
	$(shell mkdir -p $(BUILD_DIR))
	@zip -rqj $(BUILD_DIR)/$(CONFIG_ZIP_NAME) $(CONFIG_DIR)

chinadns:
	$(GO_BUILD) -o $(BUILD_DIR)/$(BINARY) $(BUILD_PACKAGE)