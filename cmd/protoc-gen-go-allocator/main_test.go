package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestModule(t *testing.T) {
	data, err := os.ReadFile("testdata/code_generator_request.pb.bin")
	require.NoError(t, err)

	req := new(pluginpb.CodeGeneratorRequest)
	err = proto.Unmarshal(data, req)
	require.NoError(t, err)

	g, err := protogen.Options{}.New(req)
	require.NoError(t, err)
	err = generator(g)
	require.NoError(t, err)
}
