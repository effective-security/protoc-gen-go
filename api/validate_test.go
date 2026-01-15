package api_test

import (
	"context"
	"testing"

	"github.com/effective-security/protoc-gen-go/e2e"
	"github.com/stretchr/testify/assert"
)

func TestValidateRequest_ListAnnotationsRequest(t *testing.T) {
	ctx := context.Background()

	tcases := []struct {
		name string
		msg  *e2e.ListAnnotationsRequest
		exp  string
	}{
		{
			name: "nil",
			msg:  nil,
			exp:  "bad_request: List Annotations Request: is not a valid protobuf message",
		},
		{
			name: "empty",
			msg:  &e2e.ListAnnotationsRequest{},
			exp:  "bad_request: Name is required",
		},
		{
			name: "with_name",
			msg: &e2e.ListAnnotationsRequest{
				Name: "test",
			},
			exp: "bad_request: AssetID: at least one of the fields must be set: ResourceID",
		},
		{
			name: "with_asset_id",
			msg: &e2e.ListAnnotationsRequest{
				Name:    "test",
				AssetID: "123456789",
			},
			exp: "bad_request: AssetIDs: minimum count is 1",
		},
		{
			name: "with_asset_ids_display_too_short",
			msg: &e2e.ListAnnotationsRequest{
				Name:     "test",
				AssetID:  "123456789",
				AssetIDs: []string{"123456789"},
				Display:  "test",
			},
			exp: "bad_request: Display: minimum length is 9",
		},
		{
			name: "with_asset_id_display_no_id",
			msg: &e2e.ListAnnotationsRequest{
				Name:     "test",
				AssetIDs: []string{"123456789"},
				Display:  "testaaaaaaaa",
			},
			exp: "bad_request: AssetID: at least one of the fields must be set: ResourceID",
		},
		{
			name: "with_asset_id_display_too_long",
			msg: &e2e.ListAnnotationsRequest{
				Name:       "test",
				ResourceID: "123456789",
				AssetIDs:   []string{"123456789"},
				Display:    "testaaaassssssssssssssssssssssssssssssssssssssaaaa",
			},
			exp: "bad_request: Display: maximum length is 19",
		},
		{
			name: "with_resource_id",
			msg: &e2e.ListAnnotationsRequest{
				Name:       "test",
				ResourceID: "123456789",
			},
			exp: "bad_request: AssetIDs: minimum count is 1",
		},
		{
			name: "with_resource_id",
			msg: &e2e.ListAnnotationsRequest{
				Name:       "test",
				AssetID:    "123456789",
				ResourceID: "123456789",
				AssetIDs:   []string{"123456789"},
				Limit:      10000,
			},
			exp: "bad_request: Limit: maximum value is 1000",
		},
		{
			name: "with_resource_id",
			msg: &e2e.ListAnnotationsRequest{
				Name:       "test",
				AssetID:    "123456789",
				ResourceID: "123456789",
				AssetIDs:   []string{"123456789", "123456789", "123456789", "123456789", "123456789"},
				Limit:      10,
			},
			exp: "bad_request: AssetIDs: maximum count is 3",
		},
		{
			name: "good",
			msg: &e2e.ListAnnotationsRequest{
				Name:       "test",
				AssetID:    "123456789",
				ResourceID: "123456789",
				AssetIDs:   []string{"123456789"},
				Display:    "testaaaaaaaa",
			},
			exp: "",
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate(ctx)
			if tc.exp == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.exp)
			}
		})
	}
}

func TestValidateRequest_Annotation(t *testing.T) {
	ctx := context.Background()

	tcases := []struct {
		name string
		msg  *e2e.Annotation
		exp  string
	}{
		{
			name: "nil",
			msg:  nil,
			exp:  "bad_request: Annotation: is not a valid protobuf message",
		},
		{
			name: "no_id",
			msg:  &e2e.Annotation{},
			exp:  "bad_request: ID is required",
		},
		{
			name: "no_name",
			msg: &e2e.Annotation{
				ID: "123456789",
			},
			exp: "bad_request: Name: minimum length is 2",
		},
		{
			name: "no_map",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
			},
			exp: "bad_request: Map: minimum count is 1",
		},
		{
			name: "map_too_long",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "111111111",
			},
			exp: "bad_request: Map: minimum count is 1",
		},
		{
			name: "no_metadata",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
			},
			exp: "bad_request: Metadata is required",
		},
		{
			name: "metadata_too_long",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map: map[string]string{
					"test":  "test",
					"test2": "test2",
					"test3": "test3",
					"test4": "test4",
				},
			},
			exp: "bad_request: Map: maximum count is 3",
		},
		{
			name: "no_basic",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
			},
			exp: "bad_request: Basic is required",
		},
		{
			name: "basic_map_too_long",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{},
			},
			exp: "bad_request: Basic.map: minimum count is 1",
		},
		{
			name: "basic_name_too_long",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map: map[string]string{"test": "test"},
				},
			},
			exp: "bad_request: Basic.name: minimum length is 8",
		},
		{
			name: "nil",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:  map[string]string{"test": "test"},
					Name: "testaaaaaaaaaaaaaaa",
				},
			},
			exp: "bad_request: Basic.values: minimum count is 1",
		},
		{
			name: "float_value_too_small",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
			},
			exp: "bad_request: FloatValue: minimum value is 1",
		},
		{
			name: "float_value_too_big",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue: 4.5,
			},
			exp: "bad_request: FloatValue: maximum value is 3",
		},
		{
			name: "no_bytes_value",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue: 2.5,
			},
			exp: "bad_request: BytesValue is required",
		},
		{
			name: "bytes_value_too_short",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue: 2.5,
				BytesValue: []byte("1"),
			},
			exp: "bad_request: BytesValue: minimum length is 2",
		},
		{
			name: "bytes_value_too_long",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue: 2.5,
				BytesValue: []byte("taaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaest"),
			},
			exp: "bad_request: BytesValue: maximum length is 10",
		},
		{
			name: "uint64_value_too_small",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue: 2.5,
				BytesValue: []byte("test"),
			},
			exp: "bad_request: Uint64Value: minimum value is 1",
		},
		{
			name: "uint64_value_too_big",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue:  2.5,
				BytesValue:  []byte("test"),
				Uint64Value: 12,
				Int64Value:  10,
			},
			exp: "bad_request: Uint64Value: maximum value is 10",
		},
		{
			name: "int64_value_too_big",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue:  2.5,
				BytesValue:  []byte("test"),
				Uint64Value: 10,
				Int64Value:  12,
			},
			exp: "bad_request: Int64Value: maximum value is 10",
		},
		{
			name: "uint32_value_too_big",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue:  2.5,
				BytesValue:  []byte("test"),
				Uint64Value: 10,
				Int64Value:  10,
				Uint32Value: 11,
			},
			exp: "bad_request: Uint32Value: maximum value is 10",
		},
		{
			name: "int32_value_too_big",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue:  2.5,
				BytesValue:  []byte("test"),
				Uint64Value: 10,
				Int64Value:  10,
				Uint32Value: 10,
				Int32Value:  11,
			},
			exp: "bad_request: Int32Value: maximum value is 10",
		},
		{
			name: "int32_value_too_small",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue:  2.5,
				BytesValue:  []byte("test"),
				Uint64Value: 10,
				Int64Value:  10,
				Uint32Value: 10,
				Int32Value:  1,
			},
			exp: "bad_request: Int32Value: minimum value is 2",
		},
		{
			name: "good",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue:  2.5,
				BytesValue:  []byte("test"),
				Uint64Value: 10,
				Int64Value:  10,
				Uint32Value: 10,
				Int32Value:  10,
			},
			exp: "bad_request: Strings is required",
		},
		{
			name: "good",
			msg: &e2e.Annotation{
				ID:   "123456789",
				Name: "test",
				Map:  map[string]string{"test": "test"},
				Metadata: []*e2e.KVPair{
					{
						Key:   "test",
						Value: "test",
					},
				},
				Basic: &e2e.Basic{
					Map:    map[string]string{"test": "test"},
					Name:   "testaaaaaaaaaaaaaaa",
					Values: []string{"test"},
				},
				FloatValue:  2.5,
				BytesValue:  []byte("test"),
				Uint64Value: 10,
				Int64Value:  10,
				Uint32Value: 10,
				Int32Value:  10,
				Strings:     []string{"test", "test2"},
			},
			exp: "",
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate(ctx)
			if tc.exp == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.exp)
			}
		})
	}
}

func TestValidateRequest_Basic(t *testing.T) {
	ctx := context.Background()

	tcases := []struct {
		name string
		msg  *e2e.Basic
		exp  string
	}{
		{
			name: "nil",
			msg:  nil,
			exp:  "bad_request: Basic: is not a valid protobuf message",
		},
		{
			name: "no_map",
			msg:  &e2e.Basic{},
			exp:  "bad_request: map: minimum count is 1",
		},
	}
	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate(ctx)
			if tc.exp == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.exp)
			}
		})
	}
}
