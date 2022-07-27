VERSION=$(shell git describe --tags --always)

.PHONY: build_all
# build
build_all:
	rm -rf bin && mkdir bin bin/linux-amd64 bin/linux-arm64 bin/darwin-amd64 bin/darwin-arm64 bin/windows-amd64\
	&& CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-w -s" -o ./bin/darwin-arm64/ ./... \
	&& CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-w -s" -o ./bin/darwin-amd64/ ./... \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-w -s" -o ./bin/linux-arm64/ ./... \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o ./bin/linux-amd64/ ./... \
	&& CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-w -s" -o ./bin/windows-amd64/ ./... \
	&& upx -9 ./bin/windows-amd64/crawlergo.exe ./bin/darwin-amd64/crawlergo ./bin/darwin-arm64/crawlergo ./bin/linux-amd64/crawlergo ./bin/linux-arm64/crawlergo

.PHONY: build
# build
build:
	rm -rf bin && mkdir bin && go build -ldflags "-w -s" -o ./bin/ ./...