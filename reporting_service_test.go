package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func sampleReportingGraphsJSON() json.RawMessage {
	return json.RawMessage(`[
		{
			"name": "cpu",
			"title": "CPU Usage",
			"vertical_label": "%CPU",
			"identifiers": null,
			"stacked": false,
			"stacked_show_total": false
		},
		{
			"name": "memory",
			"title": "Physical memory utilization",
			"vertical_label": "Bytes",
			"identifiers": null,
			"stacked": true,
			"stacked_show_total": true
		}
	]`)
}

func sampleReportingDataJSON() json.RawMessage {
	return json.RawMessage(`[{
		"name": "cpu",
		"identifier": "",
		"data": [[1700000000, 42.5, 10.2]],
		"start": 1700000000,
		"end": 1700003600,
		"legend": ["time", "user", "system"],
		"aggregations": {
			"min": {"user": 10.0, "system": 2.0},
			"max": {"user": 90.0, "system": 50.0},
			"mean": {"user": 42.5, "system": 10.2}
		}
	}]`)
}

// --- ListGraphs tests ---

func TestReportingService_ListGraphs(t *testing.T) {
	mock := &mockSubscribeCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					if method != "reporting.netdata_graphs" {
						t.Errorf("expected method reporting.netdata_graphs, got %s", method)
					}
					if params != nil {
						t.Errorf("expected nil params, got %v", params)
					}
					return sampleReportingGraphsJSON(), nil
				},
			},
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	graphs, err := svc.ListGraphs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graphs) != 2 {
		t.Fatalf("expected 2 graphs, got %d", len(graphs))
	}
	if graphs[0].Name != "cpu" {
		t.Errorf("expected name cpu, got %s", graphs[0].Name)
	}
	if graphs[0].Title != "CPU Usage" {
		t.Errorf("expected title 'CPU Usage', got %s", graphs[0].Title)
	}
	if graphs[0].VerticalLabel != "%CPU" {
		t.Errorf("expected vertical_label '%%CPU', got %s", graphs[0].VerticalLabel)
	}
	if graphs[1].Name != "memory" {
		t.Errorf("expected name memory, got %s", graphs[1].Name)
	}
	if !graphs[1].Stacked {
		t.Error("expected memory graph to be stacked")
	}
	if !graphs[1].StackedShowTotal {
		t.Error("expected memory graph to show stacked total")
	}
}

func TestReportingService_ListGraphs_Empty(t *testing.T) {
	mock := &mockSubscribeCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					return json.RawMessage(`[]`), nil
				},
			},
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	graphs, err := svc.ListGraphs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graphs) != 0 {
		t.Errorf("expected 0 graphs, got %d", len(graphs))
	}
}

func TestReportingService_ListGraphs_Error(t *testing.T) {
	mock := &mockSubscribeCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					return nil, errors.New("connection refused")
				},
			},
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListGraphs(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReportingService_ListGraphs_ParseError(t *testing.T) {
	mock := &mockSubscribeCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					return json.RawMessage(`not json`), nil
				},
			},
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.ListGraphs(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
}

// --- GetData tests ---

func TestReportingService_GetData(t *testing.T) {
	mock := &mockSubscribeCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					if method != "reporting.netdata_get_data" {
						t.Errorf("expected method reporting.netdata_get_data, got %s", method)
					}
					// Verify params structure
					p := params.([]any)
					graphs := p[0].([]map[string]any)
					if graphs[0]["name"] != string(ReportingGraphCPU) {
						t.Errorf("expected graph name cpu, got %v", graphs[0]["name"])
					}
					return sampleReportingDataJSON(), nil
				},
			},
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	data, err := svc.GetData(context.Background(), ReportingGetDataParams{
		Graphs: []ReportingGraphQuery{{Name: ReportingGraphCPU}},
		Unit:   "HOUR",
		Page:   1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 data result, got %d", len(data))
	}
	if data[0].Name != "cpu" {
		t.Errorf("expected name cpu, got %s", data[0].Name)
	}
	if data[0].Start != 1700000000 {
		t.Errorf("expected start 1700000000, got %d", data[0].Start)
	}
	if data[0].End != 1700003600 {
		t.Errorf("expected end 1700003600, got %d", data[0].End)
	}
	if len(data[0].Legend) != 3 {
		t.Fatalf("expected 3 legend entries, got %d", len(data[0].Legend))
	}
	if len(data[0].Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(data[0].Data))
	}
}

func TestReportingService_GetData_WithIdentifier(t *testing.T) {
	mock := &mockSubscribeCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					p := params.([]any)
					graphs := p[0].([]map[string]any)
					if graphs[0]["identifier"] != "sda" {
						t.Errorf("expected identifier sda, got %v", graphs[0]["identifier"])
					}
					return json.RawMessage(`[]`), nil
				},
			},
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetData(context.Background(), ReportingGetDataParams{
		Graphs: []ReportingGraphQuery{{Name: ReportingGraphDisk, Identifier: "sda"}},
		Unit:   "HOUR",
		Page:   1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReportingService_GetData_Error(t *testing.T) {
	mock := &mockSubscribeCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					return nil, errors.New("timeout")
				},
			},
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetData(context.Background(), ReportingGetDataParams{
		Graphs: []ReportingGraphQuery{{Name: ReportingGraphCPU}},
		Unit:   "HOUR",
		Page:   1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReportingService_GetData_ParseError(t *testing.T) {
	mock := &mockSubscribeCaller{
		mockAsyncCaller: mockAsyncCaller{
			mockCaller: mockCaller{
				callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
					return json.RawMessage(`not json`), nil
				},
			},
		},
	}

	svc := NewReportingService(mock, Version{Major: 25, Minor: 4})
	_, err := svc.GetData(context.Background(), ReportingGetDataParams{
		Graphs: []ReportingGraphQuery{{Name: ReportingGraphCPU}},
		Unit:   "HOUR",
		Page:   1,
	})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

// --- Conversion tests ---

func TestReportingGraphFromResponse(t *testing.T) {
	resp := ReportingGraphResponse{
		Name:             "cpu",
		Title:            "CPU Usage",
		VerticalLabel:    "%CPU",
		Identifiers:      nil,
		Stacked:          false,
		StackedShowTotal: false,
	}
	graph := reportingGraphFromResponse(resp)
	if graph.Name != "cpu" {
		t.Errorf("expected name cpu, got %s", graph.Name)
	}
	if graph.Title != "CPU Usage" {
		t.Errorf("expected title 'CPU Usage', got %s", graph.Title)
	}
}
