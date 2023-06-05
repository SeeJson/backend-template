grpc
	- 安装protobuf
		https://github.com/protocolbuffers/protobuf/releases
		下载protoc-3.17.3-linux-x86_64.zip1.6
		解压之后 文件夹bin下面 proto
		运行 ./protoc -version
	- 安装protoc-gen-go
		go get -u github.com/golang/protobuf/protoc-gen-go

	- 安装grpc
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

	- 安装protoc-gen-validate
		go get -u  github.com/envoyproxy/protoc-gen-validate

	- 生成pd grpc.pd文件
		protoc -I ../ -I . --validate_out=lang=go:.  --go_out . --go-grpc_out . --proto_path .    account.proto

		protoc -I $GOPATH/pkg/mod/github.com/envoyproxy/protoc-gen-validate@v0.6.1/ -I . --validate_out=lang=go:.  --go_out . --go-grpc_out . --proto_path .    account.proto




	