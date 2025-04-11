package controller

import (
	"fmt"
	"strings"

	sharedpb "github.com/voikin/apim-proto/gen/go/shared/v1"
)

type OpenAPI struct {
	OpenAPI string                     `json:"openapi"`
	Info    map[string]string          `json:"info"`
	Paths   map[string]map[string]Path `json:"paths"`
}

type Path struct {
	Summary    string                 `json:"summary,omitempty"`
	Parameters []Parameter            `json:"parameters,omitempty"`
	Responses  map[string]interface{} `json:"responses"`
}

type Parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
	Schema   Schema `json:"schema"`
	Example  string `json:"example,omitempty"`
}

type Schema struct {
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`
}

func parameterSchema(param *sharedpb.Parameter) Schema {
	switch param.Type {
	case sharedpb.ParameterType_PARAMETER_TYPE_INTEGER:
		return Schema{Type: "integer"}
	case sharedpb.ParameterType_PARAMETER_TYPE_UUID:
		return Schema{Type: "string", Format: "uuid"}
	default:
		return Schema{Type: "string"}
	}
}

func buildOpenAPI(graph *sharedpb.APIGraph) (*OpenAPI, error) {
	idToSegment := map[string]*sharedpb.PathSegment{}
	for _, s := range graph.Segments {
		switch seg := s.Segment.(type) {
		case *sharedpb.PathSegment_Static:
			idToSegment[seg.Static.Id] = s
		case *sharedpb.PathSegment_Param:
			idToSegment[seg.Param.Id] = s
		}
	}

	// постройка путей
	type Node struct {
		ID     string
		Parent *Node
	}
	graphMap := map[string][]string{}
	for _, e := range graph.Edges {
		graphMap[e.From] = append(graphMap[e.From], e.To)
	}

	// найти все операции
	opsBySegment := map[string][]*sharedpb.Operation{}
	for _, op := range graph.Operations {
		opsBySegment[op.PathSegmentId] = append(opsBySegment[op.PathSegmentId], op)
	}

	// DFS для генерации путей
	var paths = map[string]map[string]Path{}

	var dfs func(n *Node, segments []string, inheritedParams []Parameter)
	dfs = func(n *Node, segments []string, inheritedParams []Parameter) {
		seg := idToSegment[n.ID]
		var part string
		var newParams = append([]Parameter{}, inheritedParams...)

		switch s := seg.Segment.(type) {
		case *sharedpb.PathSegment_Static:
			part = s.Static.Name
		case *sharedpb.PathSegment_Param:
			part = fmt.Sprintf("{%s}", s.Param.Name)
			newParams = append(newParams, Parameter{
				Name:     s.Param.Name,
				In:       "path",
				Required: true,
				Schema:   parameterSchema(s.Param),
				Example:  s.Param.Example,
			})
		}

		segments = append(segments, part)

		if ops, ok := opsBySegment[n.ID]; ok {
			path := "/" + strings.Join(segments, "/")
			if _, exists := paths[path]; !exists {
				paths[path] = map[string]Path{}
			}
			for _, op := range ops {
				allParams := append([]Parameter{}, newParams...) // копируем inherited
				for _, qp := range op.QueryParameters {
					allParams = append(allParams, Parameter{
						Name:     qp.Name,
						In:       "query",
						Required: false,
						Schema:   parameterSchema(qp),
						Example:  qp.Example,
					})
				}

				resp := map[string]interface{}{}
				for _, status := range op.StatusCodes {
					resp[fmt.Sprintf("%d", status)] = map[string]string{"description": "response"}
				}

				paths[path][strings.ToLower(op.Method)] = Path{
					Summary:    op.Id,
					Parameters: allParams,
					Responses:  resp,
				}
			}
		}

		for _, child := range graphMap[n.ID] {
			dfs(&Node{ID: child, Parent: n}, segments, newParams)
		}
	}

	// Запуск DFS с корней
	used := map[string]bool{}
	for _, e := range graph.Edges {
		used[e.To] = true
	}
	for _, s := range graph.Segments {
		var id string
		switch seg := s.Segment.(type) {
		case *sharedpb.PathSegment_Static:
			id = seg.Static.Id
		case *sharedpb.PathSegment_Param:
			id = seg.Param.Id
		}
		if !used[id] {
			dfs(&Node{ID: id}, []string{}, []Parameter{})
		}
	}

	return &OpenAPI{
		OpenAPI: "3.0.0",
		Info: map[string]string{
			"title":   "Generated API",
			"version": "1.0.0",
		},
		Paths: paths,
	}, nil
}
