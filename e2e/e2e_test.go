package e2e

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	type basicWrapper struct{ Basic }
	var cases = []struct {
		Name string

		// Value and Expected MUST be pointers to structs. If Expected is
		// nil, then it is expected to be identical to Value.
		Value    any
		Expected any
	}{
		{
			"basic",
			&Basic{
				A: "hello",
				B: &Basic_Int{
					Int: 42,
				},
			},
			nil,
		},

		{
			"basic wrapped in Go struct",
			&basicWrapper{
				Basic: Basic{
					A: "hello",
					B: &Basic_Int{
						Int: 42,
					},
				},
			},
			nil,
		},

		{
			"nested",
			&Nested_Message{
				Basic: &Basic{
					A: "hello",
					B: &Basic_Int{
						Int: 42,
					},
				},
			},
			nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			require := require.New(t)

			// Verify marshaling doesn't error
			bs, err := json.Marshal(tt.Value)
			require.NoError(err)
			require.NotEmpty(bs)

			// Determine what we expect the result to be
			expected := tt.Expected
			if expected == nil {
				expected = tt.Value
			}

			// Unmarshal. We want to do this into a concrete type so we
			// use reflection here (you can't just decode into interface{})
			// and have that work.
			val := reflect.New(reflect.ValueOf(expected).Elem().Type())
			require.NoError(json.Unmarshal(bs, val.Interface()))
			require.Equal(val.Interface(), expected)
		})
	}
}

func TestEnumUnmarshalJSON(t *testing.T) {
	bs, err := json.Marshal(ServiceStatus_Running)
	require.NoError(t, err)
	require.NotEmpty(t, bs)

	var svcStatus ServiceStatus_Enum
	require.NoError(t, json.Unmarshal(bs, &svcStatus))
	require.Equal(t, ServiceStatus_Running, svcStatus)

	bs2, err := json.Marshal("Failed")
	require.NoError(t, err)
	require.NotEmpty(t, bs2)

	require.NoError(t, json.Unmarshal(bs2, &svcStatus))
	require.Equal(t, ServiceStatus_Failed, svcStatus)

	bs3, err := json.Marshal([]string{"Running", "Failed"})
	require.NoError(t, err)
	require.NotEmpty(t, bs3)

	require.NoError(t, json.Unmarshal(bs3, &svcStatus))
	require.Equal(t, ServiceStatus_Running|ServiceStatus_Failed, svcStatus)

	bs4, err := json.Marshal(JobStatus_Cancelled | JobStatus_Failed)
	require.NoError(t, err)
	require.NotEmpty(t, bs4)

	var jobStatus JobStatus_Enum
	require.NoError(t, json.Unmarshal(bs4, &jobStatus))
	require.Equal(t, JobStatus_Cancelled|JobStatus_Failed, jobStatus)

	bs5, err := json.Marshal([]string{"Cancelled", "Failed"})
	require.NoError(t, err)
	require.NotEmpty(t, bs5)

	var jobStatuses JobStatus_EnumSlice
	require.NoError(t, json.Unmarshal(bs5, &jobStatuses))
	require.Equal(t, JobStatus_EnumSlice{JobStatus_Cancelled, JobStatus_Failed}, jobStatuses)
}
