all: xdscli

xdscli:
	@go build -o out/xdscli -v -mod=vendor xdscli.go

install: xdscli
	cp out/xdscli ${GOPATH}/bin

.PHONY: all