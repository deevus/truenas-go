package truenas

import (
	"context"
	"encoding/json"
	"fmt"
)

// CronJob is the user-facing representation of a TrueNAS cron job.
// CaptureStdout/CaptureStderr use intuitive semantics (true = capture),
// unlike the API's inverted stdout/stderr fields.
type CronJob struct {
	ID            int64
	User          string
	Command       string
	Description   string
	Enabled       bool
	CaptureStdout bool
	CaptureStderr bool
	Schedule      Schedule
}

// Schedule represents a cron schedule.
type Schedule struct {
	Minute string
	Hour   string
	Dom    string
	Month  string
	Dow    string
}

// CreateCronJobOpts contains options for creating a cron job.
type CreateCronJobOpts struct {
	User          string
	Command       string
	Description   string
	Enabled       bool
	CaptureStdout bool
	CaptureStderr bool
	Schedule      Schedule
}

// UpdateCronJobOpts contains options for updating a cron job.
// All fields are always sent on update.
type UpdateCronJobOpts = CreateCronJobOpts

// Caller is the interface used by services to make API calls.
// It is satisfied by client.Client and client.MockClient.
type Caller interface {
	Call(ctx context.Context, method string, params any) (json.RawMessage, error)
}

// CronService provides typed methods for the cronjob.* API namespace.
type CronService struct {
	client Caller
}

// NewCronService creates a new CronService.
func NewCronService(c Caller) *CronService {
	return &CronService{client: c}
}

// Create creates a cron job and returns the full object.
func (s *CronService) Create(ctx context.Context, opts CreateCronJobOpts) (*CronJob, error) {
	params := optsToParams(opts)
	result, err := s.client.Call(ctx, "cronjob.create", params)
	if err != nil {
		return nil, err
	}

	var createResp struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(result, &createResp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return s.Get(ctx, createResp.ID)
}

// Get returns a cron job by ID, or nil if not found.
func (s *CronService) Get(ctx context.Context, id int64) (*CronJob, error) {
	filter := [][]any{{"id", "=", id}}
	result, err := s.client.Call(ctx, "cronjob.query", filter)
	if err != nil {
		return nil, err
	}

	var jobs []CronJobResponse
	if err := json.Unmarshal(result, &jobs); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	if len(jobs) == 0 {
		return nil, nil
	}

	job := cronJobFromResponse(jobs[0])
	return &job, nil
}

// List returns all cron jobs.
func (s *CronService) List(ctx context.Context) ([]CronJob, error) {
	result, err := s.client.Call(ctx, "cronjob.query", nil)
	if err != nil {
		return nil, err
	}

	var responses []CronJobResponse
	if err := json.Unmarshal(result, &responses); err != nil {
		return nil, fmt.Errorf("parse query response: %w", err)
	}

	jobs := make([]CronJob, len(responses))
	for i, resp := range responses {
		jobs[i] = cronJobFromResponse(resp)
	}
	return jobs, nil
}

// Update updates a cron job and returns the full object.
func (s *CronService) Update(ctx context.Context, id int64, opts UpdateCronJobOpts) (*CronJob, error) {
	params := optsToParams(opts)
	_, err := s.client.Call(ctx, "cronjob.update", []any{id, params})
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, id)
}

// Delete deletes a cron job by ID.
func (s *CronService) Delete(ctx context.Context, id int64) error {
	_, err := s.client.Call(ctx, "cronjob.delete", id)
	return err
}

// optsToParams converts CreateCronJobOpts to API parameters.
// Inverts CaptureStdout/CaptureStderr to the API's stdout/stderr semantics.
func optsToParams(opts CreateCronJobOpts) map[string]any {
	return map[string]any{
		"user":        opts.User,
		"command":     opts.Command,
		"description": opts.Description,
		"enabled":     opts.Enabled,
		"stdout":      !opts.CaptureStdout,
		"stderr":      !opts.CaptureStderr,
		"schedule": map[string]any{
			"minute": opts.Schedule.Minute,
			"hour":   opts.Schedule.Hour,
			"dom":    opts.Schedule.Dom,
			"month":  opts.Schedule.Month,
			"dow":    opts.Schedule.Dow,
		},
	}
}

// cronJobFromResponse converts a wire-format CronJobResponse to a user-facing CronJob.
// Inverts the API's stdout/stderr fields to CaptureStdout/CaptureStderr.
func cronJobFromResponse(resp CronJobResponse) CronJob {
	return CronJob{
		ID:            resp.ID,
		User:          resp.User,
		Command:       resp.Command,
		Description:   resp.Description,
		Enabled:       resp.Enabled,
		CaptureStdout: !resp.Stdout,
		CaptureStderr: !resp.Stderr,
		Schedule: Schedule{
			Minute: resp.Schedule.Minute,
			Hour:   resp.Schedule.Hour,
			Dom:    resp.Schedule.Dom,
			Month:  resp.Schedule.Month,
			Dow:    resp.Schedule.Dow,
		},
	}
}
