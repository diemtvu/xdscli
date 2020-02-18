all: xdscli
xdscli:
	@go build -o out/xdscli -v -mod=vendor xdscli.go

.PHONY: all