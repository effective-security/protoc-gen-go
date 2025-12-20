package api_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/effective-security/protoc-gen-go/api"
	"github.com/effective-security/protoc-gen-go/e2e"
	"github.com/effective-security/x/print"
	"github.com/effective-security/x/values"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
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
    - Field: Strings
      Type: keyword
    - Field: Types
      Type: integer
      Enum values: Unknown (0), Bar (1), Foo (2)
      Documentation: Types are for testing enum types.
    - Field: RefIDs
      Type: integer
      Documentation: RefIDs are for testing reference IDs.
    - Field: Hashes
      Type: integer
    - Field: Limits
      Type: integer
    - Field: Counts
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
			"mapk1": "mapv1",
			"mapk2": "mapv2",
		},
		Metadata: []*e2e.KVPair{
			{
				Key:   "metak1",
				Value: "metav1",
			},
			{
				Key:   "metak2",
				Value: "metav2",
			},
		},
		Basic: &e2e.Basic{
			Values: []string{"v1", "v2"},
			Map: map[string]string{
				"basick1": "basicv1",
				"basick2": "basicv2",
			},
			Statuses: e2e.JobStatus_Scheduled,
		},
		FloatValue:  1.23456,
		BytesValue:  []byte("test"),
		Uint64Value: 1,
		Int64Value:  1,
		Uint32Value: 1,
		Int32Value:  1,
		Types:       []e2e.AnnotationType_Enum{e2e.AnnotationType_Bar, e2e.AnnotationType_Foo},
		RefIDs:      []uint64{1, 2, 3},
		Hashes:      []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Limits:      []uint32{1, 2, 3},
		Counts:      []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	a2 := &e2e.Annotation{
		ID:   "2",
		Name: "test2",
		Type: e2e.AnnotationType_Bar,
		Map: map[string]string{
			"mapk3": "mapv3",
			"mapk4": "mapv4",
		},
		Metadata: []*e2e.KVPair{
			{
				Key:   "metak3",
				Value: "metav3",
			},
			{
				Key:   "metak4",
				Value: "metav4",
			},
		},
		Basic: &e2e.Basic{
			Values: []string{"v1", "v2"},
			Map: map[string]string{
				"basick3": "basicv3",
				"basick4": "basicv4",
			},
		},
		FloatValue:  2.23456,
		BytesValue:  []byte("test2"),
		Uint64Value: 2,
		Int64Value:  3,
		Uint32Value: 4,
		Int32Value:  5,
		Types:       []e2e.AnnotationType_Enum{e2e.AnnotationType_Foo},
		RefIDs:      []uint64{4, 5, 6},
		Hashes:      []int64{4, 5, 6},
		Limits:      []uint32{4, 5, 6},
		Counts:      []int32{4, 5, 6},
	}
	a3 := &e2e.Annotation{
		ID:   "3",
		Name: "test3",
	}

	ares := &e2e.AnnotationsResponse{
		Annotations: []*e2e.Annotation{a1, a2, a3},
	}

	td, err := api.GetTabularData(ares)
	require.NoError(t, err)
	require.Equal(t, 2, len(td.Tables))
	assert.Equal(t, "Annotations Response", td.Tables[0].ID)
	assert.Equal(t, 2, len(td.Tables[0].Header))
	assert.Equal(t, 1, len(td.Tables[0].Rows))
	assert.Equal(t, "Annotations", td.Tables[1].ID)
	assert.Equal(t, 16, len(td.Tables[1].Header))
	assert.Equal(t, 3, len(td.Tables[1].Rows))

	exp := `Annotations Response:

 Annotations │ 3 items 
 Next Offset │ 0       

Annotations:

┌────┬───────┬─────────┬─────────┬──────────┬─────────────┬───────────────┬──────────────┬───────────────┬──────────────┬─────────┬─────────┬──────────┬──────────┬─────────┬──────────┐
│ ID │ NAME  │  TYPE   │   MAP   │ METADATA │ FLOAT VALUE │ UINT 64 VALUE │ INT 64 VALUE │ UINT 32 VALUE │ INT 32 VALUE │ STRINGS │  TYPES  │ REF I DS │  HASHES  │ LIMITS  │  COUNTS  │
├────┼───────┼─────────┼─────────┼──────────┼─────────────┼───────────────┼──────────────┼───────────────┼──────────────┼─────────┼─────────┼──────────┼──────────┼─────────┼──────────┤
│ 1  │ test1 │ Bar     │ 2 items │ 2 items  │ 1.234560    │ 1             │ 1            │ 1             │ 1            │         │ 2 items │ 3 items  │ 10 items │ 3 items │ 10 items │
│ 2  │ test2 │ Bar     │ 2 items │ 2 items  │ 2.234560    │ 2             │ 3            │ 4             │ 5            │         │ 1 items │ 3 items  │ 3 items  │ 3 items │ 3 items  │
│ 3  │ test3 │ Unknown │         │          │ 0.000000    │ 0             │ 0            │ 0             │ 0            │         │         │          │          │         │          │
└────┴───────┴─────────┴─────────┴──────────┴─────────────┴───────────────┴──────────────┴───────────────┴──────────────┴─────────┴─────────┴──────────┴──────────┴─────────┴──────────┘

`
	w := bytes.NewBuffer([]byte{})
	td.Print(w)
	assert.Equal(t, exp, w.String())

	w.Reset()

	td2, err := api.GetTabularData(a1)
	require.NoError(t, err)
	require.Equal(t, 2, len(td2.Tables))
	assert.Equal(t, "Annotation", td2.Tables[0].ID)
	assert.Equal(t, 18, len(td2.Tables[0].Header))
	assert.Equal(t, 1, len(td2.Tables[0].Rows))

	assert.Equal(t, "Metadata", td2.Tables[1].ID)
	assert.Equal(t, 2, len(td2.Tables[1].Header))
	assert.Equal(t, 2, len(td2.Tables[1].Rows))

	exp2 := `Annotation:

 ID            │ 1        
 Name          │ test1    
 Type          │ Bar      
 Map           │ 2 items  
 Metadata      │ 2 items  
 Basic         │ <object> 
 Float Value   │ 1.234560 
 Bytes Value   │ dGVzdA== 
 Uint 64 Value │ 1        
 Int 64 Value  │ 1        
 Uint 32 Value │ 1        
 Int 32 Value  │ 1        
 Types         │ 2 items  
 Ref IDs       │ 3 items  
 Hashes        │ 10 items 
 Limits        │ 3 items  
 Counts        │ 10 items 

Metadata:

┌────────┬────────┐
│  KEY   │ VALUE  │
├────────┼────────┤
│ metak1 │ metav1 │
│ metak2 │ metav2 │
└────────┴────────┘

`
	td2.Print(w)
	assert.Equal(t, exp2, w.String())

}

func Test_ListAnnotationsRequest(t *testing.T) {
	t.Parallel()

	req := &e2e.ListAnnotationsRequest{
		Name:     "test",
		AssetID:  "123456789",
		AssetIDs: []string{"123456789"},
		Display:  "test12345",
		Filter:   &e2e.ListAnnotationsRequest_Category{Category: e2e.AnnotationCategory_Internal},
	}
	assert.NoError(t, req.Validate(context.Background()))

	w := bytes.NewBuffer([]byte{})
	api.DocumentMessage(w, e2e.ListAnnotationsRequest_MessageDescription, "")
	out := w.String()
	exp := `List Annotations Request:
Fields:
- Field: Name
  Type: keyword
- Field: AssetID
  Type: keyword
- Field: ResourceID
  Type: keyword
- Field: AssetIDs
  Type: keyword
- Field: Offset
  Type: integer
- Field: Limit
  Type: integer
- Field: Display
  Type: keyword
- Field: Category
  Type: integer
  Enum values: Unknown (0), Internal (1), Security (2), Compliance (4), All (2147483647)
- Field: Type
  Type: integer
  Enum values: Unknown (0), Bar (1), Foo (2)

`
	assert.Equal(t, exp, out)

	w.Reset()
	api.Describe(w, req)
	out = w.String()
	exp = `Asset ID: "123456789"
Asset IDs: "123456789"
Category: Internal
Display: test12345
Name: test
`
	assert.Equal(t, exp, out)
}

func Test_AnnotationSearchResponse(t *testing.T) {
	t.Parallel()

	print.RegisterType(([]*e2e.Facet)(nil), PrintFacetsCustom)

	a1 := &e2e.Annotation{
		ID:   "1",
		Name: "test1",
		Type: e2e.AnnotationType_Bar,
		Map: map[string]string{
			"mapk1": "mapv1",
			"mapk2": "mapv2",
		},
		Metadata: []*e2e.KVPair{
			{
				Key:   "metak1",
				Value: "metav1",
			},
			{
				Key:   "metak2",
				Value: "metav2",
			},
		},
		Basic: &e2e.Basic{
			Values: []string{"v1", "v2"},
			Map: map[string]string{
				"basick1": "basicv1",
				"basick2": "basicv2",
			},
			Statuses: e2e.JobStatus_Scheduled,
		},
		FloatValue:  1.23456,
		BytesValue:  []byte("test"),
		Uint64Value: 1,
		Int64Value:  1,
		Uint32Value: 1,
		Int32Value:  1,
		Types:       []e2e.AnnotationType_Enum{e2e.AnnotationType_Bar, e2e.AnnotationType_Foo},
		RefIDs:      []uint64{1, 2, 3},
		Hashes:      []int64{1, 2, 3},
		Limits:      []uint32{1, 2, 3},
		Counts:      []int32{1, 2, 3},
	}
	a2 := &e2e.Annotation{
		ID:   "2",
		Name: "test2",
		Type: e2e.AnnotationType_Bar,
		Map: map[string]string{
			"mapk3": "mapv3",
			"mapk4": "mapv4",
		},
		Metadata: []*e2e.KVPair{
			{
				Key:   "metak3",
				Value: "metav3",
			},
			{
				Key:   "metak4",
				Value: "metav4",
			},
		},
		Basic: &e2e.Basic{
			Values: []string{"v1", "v2"},
			Map: map[string]string{
				"basick3": "basicv3",
				"basick4": "basicv4",
			},
		},
		FloatValue:  2.23456,
		BytesValue:  []byte("test2"),
		Uint64Value: 2,
		Int64Value:  3,
		Uint32Value: 4,
		Int32Value:  5,
		Types:       []e2e.AnnotationType_Enum{e2e.AnnotationType_Foo},
		RefIDs:      []uint64{4, 5, 6},
		Hashes:      []int64{4, 5, 6},
		Limits:      []uint32{4, 5, 6},
		Counts:      []int32{4, 5, 6},
	}
	a3 := &e2e.Annotation{
		ID:   "3",
		Name: "test3",
	}

	ares := &e2e.AnnotationSearchResponse{
		Found: 3,
		Facets: []*e2e.Facet{
			{
				Name:        "facet1",
				DisplayName: "facet1",
				Count:       1,
				Facets: []*e2e.Facet{
					{
						Name:        "f1",
						DisplayName: "subfacet1",
						Count:       1,
						Buckets: []*e2e.SearchBucket{
							{
								Value:       "Bucket1",
								Count:       1,
								DisplayName: "Bucket1",
							},
							{
								Value:       "Bucket2",
								Count:       1,
								DisplayName: "Bucket2",
							},
						},
					},
					{
						Name:        "f2",
						DisplayName: "facet2",
						Count:       2,
						Buckets: []*e2e.SearchBucket{
							{
								Value:       "Bucket3",
								Count:       1,
								DisplayName: "Bucket3",
							},
						},
					},
				},
				Buckets: []*e2e.SearchBucket{
					{
						Value:       "test",
						Count:       1,
						DisplayName: "test",
						Facets: []*e2e.Facet{
							{
								Name:        "test",
								DisplayName: "test",
								Count:       1,
								Buckets: []*e2e.SearchBucket{
									{
										Value:       "test",
										Count:       1,
										DisplayName: "test",
									},
									{
										Value:       "test2",
										Count:       1,
										DisplayName: "test2",
									},
								},
							},
						},
					},
				},
			},
		},
		Foo: []*e2e.Annotation{a1, a2, a3},
		Bar: []*e2e.Annotation{a2, a3},
	}

	td, err := api.GetTabularData(ares)
	require.NoError(t, err)
	require.Equal(t, 4, len(td.Tables))

	exp := `Annotation Search Response:

 Found  │ 3       
 Facets │ 1 items 
 Foo    │ 3 items 
 Bar    │ 2 items 

Facets:

Facet: facet1
Count: 1

┌───────┬─────────┬───────┐
│ VALUE │ DISPLAY │ COUNT │
├───────┼─────────┼───────┤
│ test  │ test    │ 1     │
└───────┴─────────┴───────┘

Facet: facet1, test
Filter: test
Count: 1

┌───────┬─────────┬───────┐
│ VALUE │ DISPLAY │ COUNT │
├───────┼─────────┼───────┤
│ test  │ test    │ 1     │
│ test2 │ test2   │ 1     │
└───────┴─────────┴───────┘

Facet: facet1, f1
Count: 1

┌─────────┬─────────┬───────┐
│  VALUE  │ DISPLAY │ COUNT │
├─────────┼─────────┼───────┤
│ Bucket1 │ Bucket1 │ 1     │
│ Bucket2 │ Bucket2 │ 1     │
└─────────┴─────────┴───────┘

Facet: facet1, f2
Count: 2

┌─────────┬─────────┬───────┐
│  VALUE  │ DISPLAY │ COUNT │
├─────────┼─────────┼───────┤
│ Bucket3 │ Bucket3 │ 1     │
└─────────┴─────────┴───────┘

Foo:

┌────┬───────┬─────────┬─────────┬──────────┬─────────────┬───────────────┬──────────────┬───────────────┬──────────────┬─────────┬─────────┬──────────┬─────────┬─────────┬─────────┐
│ ID │ NAME  │  TYPE   │   MAP   │ METADATA │ FLOAT VALUE │ UINT 64 VALUE │ INT 64 VALUE │ UINT 32 VALUE │ INT 32 VALUE │ STRINGS │  TYPES  │ REF I DS │ HASHES  │ LIMITS  │ COUNTS  │
├────┼───────┼─────────┼─────────┼──────────┼─────────────┼───────────────┼──────────────┼───────────────┼──────────────┼─────────┼─────────┼──────────┼─────────┼─────────┼─────────┤
│ 1  │ test1 │ Bar     │ 2 items │ 2 items  │ 1.234560    │ 1             │ 1            │ 1             │ 1            │         │ 2 items │ 3 items  │ 3 items │ 3 items │ 3 items │
│ 2  │ test2 │ Bar     │ 2 items │ 2 items  │ 2.234560    │ 2             │ 3            │ 4             │ 5            │         │ 1 items │ 3 items  │ 3 items │ 3 items │ 3 items │
│ 3  │ test3 │ Unknown │         │          │ 0.000000    │ 0             │ 0            │ 0             │ 0            │         │         │          │         │         │         │
└────┴───────┴─────────┴─────────┴──────────┴─────────────┴───────────────┴──────────────┴───────────────┴──────────────┴─────────┴─────────┴──────────┴─────────┴─────────┴─────────┘

Bar:

┌────┬───────┬─────────┬─────────┬──────────┬─────────────┬───────────────┬──────────────┬───────────────┬──────────────┬─────────┬─────────┬──────────┬─────────┬─────────┬─────────┐
│ ID │ NAME  │  TYPE   │   MAP   │ METADATA │ FLOAT VALUE │ UINT 64 VALUE │ INT 64 VALUE │ UINT 32 VALUE │ INT 32 VALUE │ STRINGS │  TYPES  │ REF I DS │ HASHES  │ LIMITS  │ COUNTS  │
├────┼───────┼─────────┼─────────┼──────────┼─────────────┼───────────────┼──────────────┼───────────────┼──────────────┼─────────┼─────────┼──────────┼─────────┼─────────┼─────────┤
│ 2  │ test2 │ Bar     │ 2 items │ 2 items  │ 2.234560    │ 2             │ 3            │ 4             │ 5            │         │ 1 items │ 3 items  │ 3 items │ 3 items │ 3 items │
│ 3  │ test3 │ Unknown │         │          │ 0.000000    │ 0             │ 0            │ 0             │ 0            │         │         │          │         │         │         │
└────┴───────┴─────────┴─────────┴──────────┴─────────────┴───────────────┴──────────────┴───────────────┴──────────────┴─────────┴─────────┴──────────┴─────────┴─────────┴─────────┘

`
	w := bytes.NewBuffer([]byte{})
	td.Print(w)
	assert.Equal(t, exp, w.String())
}

func PrintFacetsCustom(w io.Writer, val any) {
	res := val.([]*e2e.Facet)
	printFacets(w, "", "", res)
}

func printFacets(w io.Writer, parent string, fullName string, res []*e2e.Facet) {
	for _, f := range res {
		subFacets := len(f.Facets)
		buckets := len(f.Buckets)
		if subFacets == 0 && buckets == 0 && f.Count == 0 {
			continue
		}

		name := f.Name
		if parent != "" {
			name = fmt.Sprintf("%s, %s", parent, f.Name)
		}

		_, _ = fmt.Fprintf(w, "Facet: %s\n", name)
		if fullName != "" {
			_, _ = fmt.Fprintf(w, "Filter: %s\n", fullName)
		}
		if f.Count > 0 {
			_, _ = fmt.Fprintf(w, "Count: %d\n", f.Count)
		}
		if len(f.Buckets) > 0 {
			_, _ = fmt.Fprintln(w)
			//fmt.Fprintf(w, "Buckets: %d\n\n", len(f.Buckets))
			printSearchBuckets(w, name, fullName, f.Buckets)
		}

		if subFacets > 0 {
			printFacets(w, name, fullName, f.Facets)
		}
	}
}

func printSearchBuckets(w io.Writer, parent string, fullName string, res []*e2e.SearchBucket) {
	table := createTable(w)
	table.Header([]string{"Value", "Display", "Count"})

	for _, r := range res {
		_ = table.Append([]string{
			r.Value,
			r.DisplayName,
			fmt.Sprintf("%d", r.Count),
		})
	}
	_ = table.Render()
	_, _ = fmt.Fprintln(w)

	for _, r := range res {
		if len(r.Facets) > 0 {
			fn := values.StringsCoalesce(r.DisplayName, r.Value)
			if fullName != "" {
				fn = fmt.Sprintf("%s, %s", fullName, fn)
			}
			printFacets(w, parent, fn, r.Facets)
		}
	}
}

func createTable(w io.Writer) *tablewriter.Table {
	return tablewriter.NewTable(w,
		tablewriter.WithConfig(
			tablewriter.Config{
				Row: tw.CellConfig{
					Formatting: tw.CellFormatting{
						AutoWrap:  tw.WrapTruncate,
						Alignment: tw.AlignLeft,
					},
					ColMaxWidths: tw.CellWidth{Global: 64},
				},
			},
		))
}
