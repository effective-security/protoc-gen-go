package api_test

import (
	"bytes"
	"testing"

	"github.com/effective-security/protoc-gen-go/api"
	"github.com/effective-security/protoc-gen-go/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			Map1: map[string]e2e.ResourceType_Enum{
				"key1": e2e.ResourceType_EC2Instance,
			},
			Map2: map[string]*e2e.Generic_Message{
				"key1": {
					Name: "test",
					Id:   "test",
				},
			},
			ResourceType: e2e.ResourceType_EC2Instance | e2e.ResourceType_S3Bucket,
		}
		exp := `Resource: EC2 Instance,S3 Bucket
count: 2
enabled: true
id: "1"
map 1:
    key1: EC2 Instance
map 2:
    key1:
        id: test
        name: test
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

	t.Run("Annotation", func(t *testing.T) {
		val := &e2e.Annotation{
			ID:   "test",
			Name: "test",
			Type: e2e.AnnotationType_Bar,
			Map: map[string]string{
				"test1": "testv",
				"test2": "testv2",
			},
			Metadata: []*e2e.KVPair{
				{
					Key:   "test",
					Value: "test",
				},
			},
			Basic: &e2e.Basic{
				Values: []string{"v1", "v2"},
				Map: map[string]string{
					"k1": "v1",
					"k2": "v2",
				},
			},
			FloatValue:  1.23456,
			BytesValue:  []byte("test"),
			Uint64Value: 1,
			Int64Value:  1,
			Uint32Value: 1,
			Int32Value:  1,
		}
		exp := `Basic:
    map:
        k1: v1
        k2: v2
    values:
        - v1
        - v2
Bytes Value: dGVzdA==
Float Value: 1.23456
ID: test
Int 32 Value: 1
Int 64 Value: "1"
Map:
    test1: testv
    test2: testv2
Metadata:
    - Key: test
      Value: test
Name: test
Type: Bar
Uint 32 Value: 1
Uint 64 Value: "1"
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
    - Field: name
      Type: keyword
    - Field: values
      Type: keyword

`
	checkDoc(e2e.Basic_MessageDescription, "  ", exp1)

	exp2 := `Annotation:
  Fields:
    - Field: ID
      Type: keyword
    - Field: Name
      Type: keyword
    - Field: Type
      Type: integer
      Enum values: Unknown (0), Bar (1), Foo (2)
    - Field: Map
      Type: flat_object
    - Field: Metadata
      Type: flat_object
      Documentation: Metadata is a list of internal metadata associated with the asset
    - Field: Basic
      Type: flat_object
    - Field: FloatValue
      Type: float
    - Field: BytesValue
      Type: keyword
    - Field: Uint64Value
      Type: integer
    - Field: Int64Value
      Type: integer
    - Field: Uint32Value
      Type: integer
    - Field: Int32Value
      Type: integer

`
	checkDoc(e2e.Annotation_MessageDescription, "  ", exp2)

	exp3 := `Annotations Response:
  Fields:
    - Field: Annotations
      Type: flat_object
    - Field: NextOffset
      Type: integer

`
	checkDoc(e2e.AnnotationsResponse_MessageDescription, "  ", exp3)

}

func Test_GetTabularData(t *testing.T) {
	t.Parallel()

	a1 := &e2e.Annotation{
		ID:   "1",
		Name: "test1",
		Type: e2e.AnnotationType_Bar,
		Map: map[string]string{
			"test1": "testv",
			"test2": "testv2",
		},
		Metadata: []*e2e.KVPair{
			{
				Key:   "test",
				Value: "test",
			},
		},
		Basic: &e2e.Basic{
			Values: []string{"v1", "v2"},
			Map: map[string]string{
				"k1": "v1",
				"k2": "v2",
			},
		},
		FloatValue:  1.23456,
		BytesValue:  []byte("test"),
		Uint64Value: 1,
		Int64Value:  1,
		Uint32Value: 1,
		Int32Value:  1,
	}
	a2 := &e2e.Annotation{
		ID:   "2",
		Name: "test2",
		Type: e2e.AnnotationType_Bar,
		Map: map[string]string{
			"test3": "testv3",
			"test2": "testv2",
		},
		Metadata: []*e2e.KVPair{
			{
				Key:   "test5",
				Value: "test6",
			},
		},
		Basic: &e2e.Basic{
			Values: []string{"v1", "v2"},
			Map: map[string]string{
				"k3": "v3",
				"k2": "v2",
			},
		},
		FloatValue:  2.23456,
		BytesValue:  []byte("test2"),
		Uint64Value: 2,
		Int64Value:  3,
		Uint32Value: 4,
		Int32Value:  5,
	}

	ares := &e2e.AnnotationsResponse{
		Annotations: []*e2e.Annotation{a1, a2},
	}

	td, err := api.GetTabularData(ares)
	require.NoError(t, err)
	require.Equal(t, 1, len(td.Tables))
	assert.Equal(t, "Annotations", td.Tables[0].ID)
	assert.Equal(t, 8, len(td.Tables[0].Header))
	assert.Equal(t, 2, len(td.Tables[0].Rows))

	exp := `Annotations:

┌────┬───────┬──────┬─────────────┬───────────────┬──────────────┬───────────────┬──────────────┐
│ ID │ NAME  │ TYPE │ FLOAT VALUE │ UINT 64 VALUE │ INT 64 VALUE │ UINT 32 VALUE │ INT 32 VALUE │
├────┼───────┼──────┼─────────────┼───────────────┼──────────────┼───────────────┼──────────────┤
│ 1  │ test1 │ Bar  │ 1.234560    │ 1             │ 1            │ 1             │ 1            │
│ 2  │ test2 │ Bar  │ 2.234560    │ 2             │ 3            │ 4             │ 5            │
└────┴───────┴──────┴─────────────┴───────────────┴──────────────┴───────────────┴──────────────┘

`
	w := bytes.NewBuffer([]byte{})
	td.Print(w)
	assert.Equal(t, exp, w.String())

	w.Reset()

	td2, err := api.GetTabularData(a1)
	require.NoError(t, err)
	require.Equal(t, 1, len(td2.Tables))
	assert.Equal(t, "Annotation", td2.Tables[0].ID)
	assert.Equal(t, 8, len(td2.Tables[0].Header))
	assert.Equal(t, 1, len(td2.Tables[0].Rows))

	exp2 := `Annotation:

 ID            │ 1        
 Name          │ test1    
 Type          │ Bar      
 Float Value   │ 1.234560 
 Uint 64 Value │ 1        
 Int 64 Value  │ 1        
 Uint 32 Value │ 1        
 Int 32 Value  │ 1        

`
	td2.Print(w)
	assert.Equal(t, exp2, w.String())

}
