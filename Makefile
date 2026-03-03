BINARY=site-server
GOFLAGS=-ldflags="-s -w"

.PHONY: build linux linux-arm64 clean all

# Standard-Build für die aktuelle Plattform
build:
	go build $(GOFLAGS) -o $(BINARY) .

# Cross-Compile für Linux amd64
linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) -o $(BINARY)-linux-amd64 .

# Cross-Compile für Linux arm64 (Raspberry Pi, ARM-Server)
linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(GOFLAGS) -o $(BINARY)-linux-arm64 .

# Beide Linux-Architekturen
all: linux linux-arm64

# Aufräumen
clean:
	rm -f $(BINARY) $(BINARY)-linux-amd64 $(BINARY)-linux-arm64
