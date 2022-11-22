// Package tools for go mod

//go:build tools
// +build tools

package tools

import (
	_ "github.com/go-phorce/cov-report/cmd/cov-report"
	_ "golang.org/x/lint/golint"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
