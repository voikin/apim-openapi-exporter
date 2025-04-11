package controller

import (
	openapiexporterpb "github.com/voikin/apim-proto/gen/go/apim_openapi_exporter/v1"
)

type Controller struct {
	openapiexporterpb.UnimplementedOpenAPIExporterServiceServer
}

func New() *Controller {
	return &Controller{}
}
