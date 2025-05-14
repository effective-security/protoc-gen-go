package api_test

import (
	"bytes"
	"testing"

	"github.com/effective-security/protoc-gen-go/api"
	"github.com/effective-security/protoc-gen-go/e2e"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func Test_Describe(t *testing.T) {
	t.Parallel()
	d := api.NewDescriber(e2e.EnumNameTypes)

	checkEqual := func(val proto.Message, exp string) {
		w := bytes.NewBuffer([]byte{})
		d.Describe(w, val)
		out := w.String()
		assert.Equal(t, exp, out, "%T", val)
	}

	t.Run("CallerStatusResponse", func(t *testing.T) {
		val := &e2e.CallerStatusResponse{
			Subject: "test",
			Role:    "user",
		}
		exp := `Role: user
Subject: test
`
		checkEqual(val, exp)
	})

	t.Run("ServerStatusResponse", func(t *testing.T) {
		val := &e2e.ServerStatusResponse{
			Status: &e2e.ServerStatus{
				Name:       "test",
				ListenUrls: []string{"u1", "u2"},
				Status:     e2e.ServiceStatus_Running,
			},
			Version: &e2e.ServerVersion{
				Build: "2023-01-01",
			},
		}
		exp := `Status:
    Listen Urls:
        - u1
        - u2
    Name: test
    Status: Running
Version:
    Build: "2023-01-01"
`
		checkEqual(val, exp)
	})

	t.Run("Generic", func(t *testing.T) {
		val := &e2e.Generic{
			Messages: []*e2e.Generic_Message{
				{
					Name: "test",
					Id:   "test",
				},
			},
			Name:    "test",
			Id:      1,
			Count:   2,
			Size:    3,
			Enabled: true,
			Value:   4.5,
			Price:   5.5,
			Map: map[string]string{
				"key1": "value1",
			},
			ResourceType: e2e.ResourceType_EC2Instance | e2e.ResourceType_S3Bucket,
		}
		exp := `Resource: EC2 Instance,S3 Bucket
count: 2
enabled: true
id: "1"
messages:
    - id: test
      name: test
name: test
price: 5.5
size: "3"
value: 4.5
`
		checkEqual(val, exp)
	})
}

func Test_DocumentMessage(t *testing.T) {
	t.Parallel()

	checkDoc := func(dscr *api.MessageDescription, indent string, exp string) {
		w := bytes.NewBuffer([]byte{})
		api.DocumentMessage(w, dscr, indent)
		out := w.String()
		assert.Equal(t, exp, out)
	}

	exp1 := `Basic:
  Basic just tests basic fields, including oneofs and so on that don't
  generally work automatically with encoding/json.
  Fields:
    - Field: a
      Type: keyword
    - Field: int
      Type: integer
    - Field: str
      Type: keyword
    - Field: id
      Type: integer
    - Field: map
      Type: flat_object
    - Field: created
      Type: flat_object
    - Field: statuses
      Type: integer
      Enum values: Unknown (0), Scheduled (1), Running (2), Succeeded (4), Failed (16), Cancelled (32), All (2147483647)
    - Field: resource_types
      Type: integer
      Enum values: Unknown (0), EC2 Instance (1), S3 Bucket (2), Lambda Function (4), All (2147483647)

`
	checkDoc(e2e.Basic_MessageDescription, "  ", exp1)
}
