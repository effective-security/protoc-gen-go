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
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go install github.com/go-phorce/cov-report/cmd/cov-report
	go install golang.org/x/lint/golint

build:
	echo "*** Building plugins"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/protoc-gen-go-json ./cmd/protoc-gen-go-json

proto:
	echo "*** Building proto"
	export PATH=${PROJ_ROOT}/bin:$$PATH && \
    cd ${PROJ_ROOT}/e2e && protoc --go_out=. --go-json_out=logtostderr=true,v=10,multiline=true,partial=true:. *.proto

