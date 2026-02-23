package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// ReportingGraph is the user-facing representation of a graph definition.
type ReportingGraph struct {
	Name             string
	Title            string
	VerticalLabel    string
	Identifiers      []string
	Stacked          bool
	StackedShowTotal bool
}

// ReportingData is the user-facing representation of graph data.
type ReportingData struct {
	Name         string
	Identifier   string
	Data         [][]json.Number
	Start        int64
	End          int64
	Legend       []string
	Aggregations ReportingAggregations
}

// ReportingAggregations contains statistical aggregations.
type ReportingAggregations struct {
	Min  map[string]json.Number
	Max  map[string]json.Number
	Mean map[string]json.Number
}

// ReportingService provides typed methods for the reporting.* API namespace.
type ReportingService struct {
	client  SubscribeCaller
	version Version
}

// NewReportingService creates a new ReportingService.
func NewReportingService(c SubscribeCaller, v Version) *ReportingService {
	return &ReportingService{client: c, version: v}
}

// ListGraphs returns all available reporting graph definitions.
func (s *ReportingService) ListGraphs(ctx context.Context) ([]ReportingGraph, error) {
	result, err := s.client.Call(ctx, "reporting.netdata_graphs", nil)
	if err != nil {
		return nil, err
	}

	var responses []ReportingGraphResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse reporting.netdata_graphs response: %w", err)
	}

	graphs := make([]ReportingGraph, len(responses))
	for i, resp := range responses {
		graphs[i] = reportingGraphFromResponse(resp)
	}
	return graphs, nil
}

// GetData returns reporting data for the specified graphs.
func (s *ReportingService) GetData(ctx context.Context, params ReportingGetDataParams) ([]ReportingData, error) {
	graphs := make([]map[string]any, len(params.Graphs))
	for i, g := range params.Graphs {
		m := map[string]any{"name": string(g.Name)}
		if g.Identifier != "" {
			m["identifier"] = g.Identifier
		}
		graphs[i] = m
	}

	callParams := []any{graphs, map[string]any{
		"unit": params.Unit,
		"page": params.Page,
	}}

	result, err := s.client.Call(ctx, "reporting.netdata_get_data", callParams)
	if err != nil {
		return nil, err
	}

	var responses []ReportingDataResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse reporting.netdata_get_data response: %w", err)
	}

	data := make([]ReportingData, len(responses))
	for i, resp := range responses {
		data[i] = reportingDataFromResponse(resp)
	}
	return data, nil
}

func reportingGraphFromResponse(resp ReportingGraphResponse) ReportingGraph {
	return ReportingGraph{
		Name:             resp.Name,
		Title:            resp.Title,
		VerticalLabel:    resp.VerticalLabel,
		Identifiers:      resp.Identifiers,
		Stacked:          resp.Stacked,
		StackedShowTotal: resp.StackedShowTotal,
	}
}

func reportingDataFromResponse(resp ReportingDataResponse) ReportingData {
	return ReportingData{
		Name:       resp.Name,
		Identifier: resp.Identifier,
		Data:       resp.Data,
		Start:      resp.Start,
		End:        resp.End,
		Legend:     resp.Legend,
		Aggregations: ReportingAggregations{
			Min:  resp.Aggregations.Min,
			Max:  resp.Aggregations.Max,
			Mean: resp.Aggregations.Mean,
		},
	}
}
