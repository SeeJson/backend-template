#!/bin/sh

API_PATH=../../../../../api/account/; protoc -I $GOPATH/pkg/mod/github.com/envoyproxy/protoc-gen-validate@v0.6.1/ -I . --validate_out=lang=go:$API_PATH  --go_out $API_PATH --go-grpc_out $API_PATH --proto_path .    account.proto