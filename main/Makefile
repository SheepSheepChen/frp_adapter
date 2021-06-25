SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o frp-adapter main.go
	@docker build -t frp-adapter .
	@rm frp-adapter

push: build
	@docker tag frp-adapter registry.cn-hangzhou.aliyuncs.com/czymirror/frp-adapter:v3.04
	@docker push registry.cn-hangzhou.aliyuncs.com/czymirror/frp-adapter:v3.04

