all: xdscli

xdscli:
	@go build -o out/linux_amd64/xdscli -v -mod=vendor xdscli.go

install: xdscli
	cp out/linux_amd64/xdscli ${GOPATH}/bin

.PHONY: all