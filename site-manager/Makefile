BINARY=site-manager
GOFLAGS=-ldflags="-s -w"

.PHONY: build linux darwin windows clean deps all

# Standard-Build für die aktuelle Plattform
build: deps
	go build $(GOFLAGS) -o $(BINARY) .

# Cross-Compile für Linux amd64
linux: deps
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) -o $(BINARY)-linux-amd64 .

# Cross-Compile für macOS (Apple Silicon + Intel)
darwin: deps
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(GOFLAGS) -o $(BINARY)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) -o $(BINARY)-darwin-amd64 .

# Cross-Compile für Windows (64-Bit)
windows: deps
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) -o $(BINARY)-windows-amd64.exe .

# Alle Plattformen bauen
all: linux darwin windows

# Dependencies herunterladen
deps:
	go mod tidy
	go mod download

# Aufräumen
clean:
	rm -f $(BINARY) $(BINARY)-linux-amd64 $(BINARY)-darwin-arm64 $(BINARY)-darwin-amd64 $(BINARY)-windows-amd64.exe
