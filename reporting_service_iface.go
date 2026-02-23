package truenas

import "context"

// ReportingServiceAPI defines the interface for reporting operations.
type ReportingServiceAPI interface {
	ListGraphs(ctx context.Context) ([]ReportingGraph, error)
	GetData(ctx context.Context, params ReportingGetDataParams) ([]ReportingData, error)
}

// Compile-time checks.
var _ ReportingServiceAPI = (*ReportingService)(nil)
var _ ReportingServiceAPI = (*MockReportingService)(nil)

// MockReportingService is a test double for ReportingServiceAPI.
type MockReportingService struct {
	ListGraphsFunc func(ctx context.Context) ([]ReportingGraph, error)
	GetDataFunc    func(ctx context.Context, params ReportingGetDataParams) ([]ReportingData, error)
}

func (m *MockReportingService) ListGraphs(ctx context.Context) ([]ReportingGraph, error) {
	if m.ListGraphsFunc != nil {
		return m.ListGraphsFunc(ctx)
	}
	return nil, nil
}

func (m *MockReportingService) GetData(ctx context.Context, params ReportingGetDataParams) ([]ReportingData, error) {
	if m.GetDataFunc != nil {
		return m.GetDataFunc(ctx, params)
	}
	return nil, nil
}
