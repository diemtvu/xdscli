all: xds

xds:
	@go build -o out/linux_amd64/xds -v -mod=vendor xdscli.go

install: xds
	cp out/linux_amd64/xds ${GOPATH}/bin

.PHONY: all