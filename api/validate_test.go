package api_test

import (
	"context"
	"testing"

	"github.com/effective-security/protoc-gen-go/e2e"
	"github.com/stretchr/testify/assert"
)

func TestValidateRequest(t *testing.T) {
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
