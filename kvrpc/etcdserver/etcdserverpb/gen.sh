#!/usr/bin/env bash

protoc -I$GOPATH/src -I. \
	-I$GOPATH/src/github.com/gogo/protobuf \
	-I$GOPATH/src/github.com/gogo/protobuf/protobuf \
	-I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--gogo_out=plugins=grpc:. \
	*.proto