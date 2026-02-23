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

// RealtimeUpdate is the user-facing representation of a realtime reporting event.
type RealtimeUpdate struct {
	CPU        map[string]RealtimeCPU
	Memory     RealtimeMemory
	Disks      map[string]RealtimeDisk
	Interfaces map[string]RealtimeInterface
}

// RealtimeCPU contains per-CPU usage and temperature metrics.
type RealtimeCPU struct {
	Usage       float64
	Temperature float64
}

// RealtimeMemory contains system memory metrics.
type RealtimeMemory struct {
	PhysicalTotal     int64
	PhysicalAvailable int64
	ArcSize           int64
}

// RealtimeDisk contains per-disk I/O metrics.
type RealtimeDisk struct {
	ReadBytesPerSec  float64
	WriteBytesPerSec float64
	BusyPercent      float64
}

// RealtimeInterface contains per-interface network metrics.
type RealtimeInterface struct {
	ReceivedBytesRate float64
	SentBytesRate     float64
	LinkState         string
	Speed             int
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

// SubscribeRealtime subscribes to real-time system metrics (CPU, memory, disk, network).
func (s *ReportingService) SubscribeRealtime(ctx context.Context) (*Subscription[RealtimeUpdate], error) {
	rawSub, err := s.client.Subscribe(ctx, "reporting.realtime", nil)
	if err != nil {
		return nil, err
	}

	typedCh := make(chan RealtimeUpdate, 100)
	go func() {
		defer close(typedCh)
		for raw := range rawSub.C {
			var resp RealtimeUpdateResponse
			if err := json.Unmarshal(raw, &resp); err != nil {
				continue // skip malformed events
			}
			typedCh <- realtimeUpdateFromResponse(resp)
		}
	}()

	return &Subscription[RealtimeUpdate]{
		C:      typedCh,
		cancel: rawSub.Close,
	}, nil
}

func realtimeUpdateFromResponse(resp RealtimeUpdateResponse) RealtimeUpdate {
	cpus := make(map[string]RealtimeCPU, len(resp.CPU))
	for k, v := range resp.CPU {
		cpus[k] = RealtimeCPU{Usage: v.Usage, Temperature: v.Temperature}
	}

	disks := make(map[string]RealtimeDisk, len(resp.Disks))
	for k, v := range resp.Disks {
		disks[k] = RealtimeDisk{
			ReadBytesPerSec:  v.ReadBytesPerSec,
			WriteBytesPerSec: v.WriteBytesPerSec,
			BusyPercent:      v.BusyPercent,
		}
	}

	ifaces := make(map[string]RealtimeInterface, len(resp.Interfaces))
	for k, v := range resp.Interfaces {
		ifaces[k] = RealtimeInterface{
			ReceivedBytesRate: v.ReceivedBytesRate,
			SentBytesRate:     v.SentBytesRate,
			LinkState:         v.LinkState,
			Speed:             v.Speed,
		}
	}

	return RealtimeUpdate{
		CPU: cpus,
		Memory: RealtimeMemory{
			PhysicalTotal:     resp.Memory.PhysicalMemoryTotal,
			PhysicalAvailable: resp.Memory.PhysicalMemoryAvailable,
			ArcSize:           resp.Memory.ArcSize,
		},
		Disks:      disks,
		Interfaces: ifaces,
	}
}
