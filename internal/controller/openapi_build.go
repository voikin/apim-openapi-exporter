package controller

import (
	"context"
	"encoding/json"
	"fmt"

	openapiexporterpb "github.com/voikin/apim-proto/gen/go/apim_openapi_exporter/v1"
)

func (c *Controller) BuildOpenAPISpec(_ context.Context, req *openapiexporterpb.BuildOpenAPISpecRequest) (*openapiexporterpb.BuildOpenAPISpecResponse, error) {
	spec, err := buildOpenAPI(req.ApiGraph)
	if err != nil {
		return nil, fmt.Errorf("buildOpenAPI: %w", err)
	}

	specJson, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	return &openapiexporterpb.BuildOpenAPISpecResponse{
		SpecJson: string(specJson),
	}, nil
}
