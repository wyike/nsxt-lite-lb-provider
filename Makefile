VERSION := $(shell date "+%s")
PLATFORMS := darwin/arm64 darwin/amd64 linux/amd64

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

.PHONY: binaries $(PLATFORMS)

binaries: $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) CGO_ENABLED=0 go build -a -ldflags "-w -s" -o nsxt-lite-lb-provider-$(os)-$(arch)-$(VERSION)
