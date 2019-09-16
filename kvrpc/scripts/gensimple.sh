#!/usr/bin/env bash

set -e

if ! [[ "$0" =~ scripts/gensimple.sh ]]; then
	echo "must be run from repository root"
	exit 255
fi

DIRS="./etcdserver/etcdserverpb"

ETCD_IO_ROOT="${GOPATH}/src/hank.com"
ETCD_ROOT="${ETCD_IO_ROOT}/etcd-3.3.12-hign"
GOGOPROTO_ROOT="${GOPATH}/src/github.com/gogo/protobuf"
SCHWAG_ROOT="${GOPATH}/src/github.com/hexfusion/schwag"
GOGOPROTO_PATH="${GOGOPROTO_ROOT}:${GOGOPROTO_ROOT}/protobuf"
GRPC_GATEWAY_ROOT="${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway"

#go get -u github.com/gogo/protobuf/{proto,protoc-gen-gogo,gogoproto}

#pushd "${GOGOPROTO_ROOT}"
#	go install ./protoc-gen-gogo
#popd

# generate gateway code
#go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
#go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

#TODO git reset --hard SHA
#pushd "${GRPC_GATEWAY_ROOT}"
#	go install ./protoc-gen-grpc-gateway
#popd

for dir in ${DIRS}; do
	pushd "${dir}"
            protoc -I$GOPATH/src -I. \
            	-I$GOPATH/src/github.com/gogo/protobuf \
            	-I$GOPATH/src/github.com/gogo/protobuf/protobuf \
            	-I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
            	--gogo_out=plugins=grpc:. \
            	*.proto
		popd
done



