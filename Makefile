include .project/gomod-project.mk
export GO111MODULE=on
BUILD_FLAGS=
export COVERAGE_EXCLUSIONS="vendor|tests|api/pb/gw|main.go|testsuite.go|mocks.go|.pb.go|.pb.gw.go"

.PHONY: *

.SILENT:

default: help

all: clean tools build proto covtest

#
# clean produced files
#
clean:
	go clean ./...
	rm -rf \
		${COVPATH} \
		${PROJ_BIN}

tools:
	echo "*** Building tools"
	go install github.com/lyft/protoc-gen-star/protoc-gen-debug
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go install github.com/go-phorce/cov-report/cmd/cov-report
	go install golang.org/x/lint/golint
	go install golang.org/x/tools/cmd/goimports

build:
	echo "*** Building plugins"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/protoc-gen-go-json ./cmd/protoc-gen-go-json
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/protoc-gen-go-mock ./cmd/protoc-gen-go-mock
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/protoc-gen-go-proxy ./cmd/protoc-gen-go-proxy
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/protoc-gen-go-allocator ./cmd/protoc-gen-go-allocator

proto-dbg:
	cd ${PROJ_ROOT}/e2e/proto && \
	protoc \
		-I=. \
		-I=../../proto \
		--plugin=protoc-gen-debug=${PROJ_ROOT}/bin/protoc-gen-debug \
		--debug_out=".:." \
		status.proto

proto:
	echo "*** Building proto"
	export PATH=${PROJ_ROOT}/bin:$$PATH && \
	cd ${PROJ_ROOT}/proto/es/api && \
	protoc \
		-I=. \
		-I=../../ \
		--go_out=paths=source_relative:./../../../api \
		--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:./../../../api \
		*.proto && \
    cd ${PROJ_ROOT}/e2e/proto && \
	protoc \
		-I=. \
		-I=../../proto \
		--go_out=paths=source_relative:./.. \
		--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:./.. \
		--go-json_out=logs=true,multiline=true,partial=true:./.. \
		--go-mock_out=logs=true:./.. \
		--go-proxy_out=logs=true:./.. \
		--go-allocator_out=logs=true:./.. \
		*.proto && \
	cd ${PROJ_ROOT}/e2e && \
	find . -name \*.go -exec sh -c "goimports -l -w {} && gofmt -s -l -w {}" \;

