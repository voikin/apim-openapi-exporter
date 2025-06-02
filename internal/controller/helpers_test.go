package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
	sharedpb "github.com/voikin/apim-proto/gen/go/shared/v1"
)

func TestBuildOpenAPI_SimplePathWithParams(t *testing.T) {
	graph := &sharedpb.APIGraph{
		Segments: []*sharedpb.PathSegment{
			{Segment: &sharedpb.PathSegment_Static{
				Static: &sharedpb.StaticSegment{Id: "1", Name: "api"},
			}},
			{Segment: &sharedpb.PathSegment_Static{
				Static: &sharedpb.StaticSegment{Id: "2", Name: "v1"},
			}},
			{Segment: &sharedpb.PathSegment_Static{
				Static: &sharedpb.StaticSegment{Id: "3", Name: "users"},
			}},
			{Segment: &sharedpb.PathSegment_Param{
				Param: &sharedpb.Parameter{
					Id:      "4",
					Name:    "user_id",
					Type:    sharedpb.ParameterType_PARAMETER_TYPE_UUID,
					Example: "123e4567-e89b-12d3-a456-426614174000",
				},
			}},
		},
		Edges: []*sharedpb.Edge{
			{From: "1", To: "2"},
			{From: "2", To: "3"},
			{From: "3", To: "4"},
		},
		Operations: []*sharedpb.Operation{
			{
				Id:            "getUser",
				Method:        "GET",
				PathSegmentId: "4",
				QueryParameters: []*sharedpb.Parameter{
					{
						Name:    "verbose",
						Type:    sharedpb.ParameterType_PARAMETER_TYPE_INTEGER,
						Example: "1",
					},
				},
				StatusCodes: []int32{200, 404},
			},
		},
	}

	result, err := buildOpenAPI(graph)
	require.NoError(t, err)
	require.NotNil(t, result)

	expectedPath := "/api/v1/users/{user_id}"
	pathItem, exists := result.Paths[expectedPath]
	require.True(t, exists, "expected path not found")

	op, exists := pathItem["get"]
	require.True(t, exists, "GET operation not found")

	require.Equal(t, "getUser", op.Summary)

	require.Len(t, op.Parameters, 2)
	require.Contains(t, op.Responses, "200")
	require.Contains(t, op.Responses, "404")

	// Check parameter details
	paramNames := map[string]bool{}
	for _, p := range op.Parameters {
		paramNames[p.Name] = true
		if p.Name == "user_id" {
			require.Equal(t, "path", p.In)
			require.True(t, p.Required)
			require.Equal(t, "string", p.Schema.Type)
			require.Equal(t, "uuid", p.Schema.Format)
		}
		if p.Name == "verbose" {
			require.Equal(t, "query", p.In)
			require.False(t, p.Required)
			require.Equal(t, "integer", p.Schema.Type)
		}
	}
	require.True(t, paramNames["user_id"])
	require.True(t, paramNames["verbose"])
}
