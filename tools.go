// Package tools for go mod

//go:build tools
// +build tools

package tools

import (
	_ "github.com/go-phorce/cov-report/cmd/cov-report"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "github.com/lyft/protoc-gen-star/protoc-gen-debug"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	_ "google.golang.org/grpc"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/status"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
