package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestVMService_CreateVM(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.create" {
					t.Errorf("expected method vm.create, got %s", method)
				}
				p := params.(map[string]any)
				if p["name"] != "test-vm" {
					t.Errorf("expected name test-vm, got %v", p["name"])
				}
				if p["memory"] != int64(2048) {
					t.Errorf("expected memory 2048, got %v", p["memory"])
				}
				return sampleVMJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.CreateVM(context.Background(), CreateVMOpts{
		Name:            "test-vm",
		Description:     "A test VM",
		VCPUs:           1,
		Cores:           2,
		Threads:         1,
		Memory:          2048,
		Autostart:       true,
		Time:            "LOCAL",
		Bootloader:      "UEFI",
		BootloaderOVMF:  "OVMF_CODE.fd",
		CPUMode:         "HOST-MODEL",
		ShutdownTimeout: 90,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm == nil {
		t.Fatal("expected non-nil vm")
	}
	if vm.ID != 1 {
		t.Errorf("expected ID 1, got %d", vm.ID)
	}
	if vm.Name != "test-vm" {
		t.Errorf("expected name test-vm, got %s", vm.Name)
	}
	if vm.State != "RUNNING" {
		t.Errorf("expected state RUNNING, got %s", vm.State)
	}
	if vm.CPUModel != "" {
		t.Errorf("expected empty cpu_model, got %s", vm.CPUModel)
	}
	// Verify only one call was made (no re-read)
	if len(mock.calls) != 1 {
		t.Errorf("expected 1 call (no re-read), got %d", len(mock.calls))
	}
}

func TestVMService_CreateVM_WithMinMemory(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				if _, ok := p["min_memory"]; !ok {
					t.Error("expected min_memory to be set in params")
				}
				if p["min_memory"] != int64(1024) {
					t.Errorf("expected min_memory 1024, got %v", p["min_memory"])
				}
				return sampleVMJSON(), nil
			},
		},
	}

	minMem := int64(1024)
	svc := NewVMService(mock, Version{})
	_, err := svc.CreateVM(context.Background(), CreateVMOpts{
		Name:      "test-vm",
		Memory:    2048,
		MinMemory: &minMem,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMService_CreateVM_WithCPUModel(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				p := params.(map[string]any)
				if p["cpu_model"] != "Haswell" {
					t.Errorf("expected cpu_model Haswell, got %v", p["cpu_model"])
				}
				return sampleVMWithCPUModelJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.CreateVM(context.Background(), CreateVMOpts{
		Name:     "model-vm",
		CPUMode:  "CUSTOM",
		CPUModel: "Haswell",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm.CPUModel != "Haswell" {
		t.Errorf("expected CPUModel Haswell, got %s", vm.CPUModel)
	}
}

func TestVMService_CreateVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("connection refused")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.CreateVM(context.Background(), CreateVMOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if vm != nil {
		t.Error("expected nil vm on error")
	}
}

func TestVMService_CreateVM_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.CreateVM(context.Background(), CreateVMOpts{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestVMService_GetVM(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.get_instance" {
					t.Errorf("expected method vm.get_instance, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 1 {
					t.Errorf("expected id 1, got %v", params)
				}
				return sampleVMJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.GetVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm == nil {
		t.Fatal("expected non-nil vm")
	}
	if vm.ID != 1 {
		t.Errorf("expected ID 1, got %d", vm.ID)
	}
	if vm.Name != "test-vm" {
		t.Errorf("expected name test-vm, got %s", vm.Name)
	}
	if vm.Description != "A test VM" {
		t.Errorf("expected description 'A test VM', got %q", vm.Description)
	}
	if vm.Memory != 2048 {
		t.Errorf("expected memory 2048, got %d", vm.Memory)
	}
	if vm.Autostart != true {
		t.Error("expected autostart true")
	}
	if vm.State != "RUNNING" {
		t.Errorf("expected state RUNNING, got %s", vm.State)
	}
}

func TestVMService_GetVM_NullCPUModel(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return sampleVMJSON(), nil // cpu_model is null
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.GetVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm.CPUModel != "" {
		t.Errorf("expected empty CPUModel for null, got %q", vm.CPUModel)
	}
}

func TestVMService_GetVM_NullMinMemory(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return sampleVMJSON(), nil // min_memory is null
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.GetVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm.MinMemory != nil {
		t.Errorf("expected nil MinMemory, got %v", vm.MinMemory)
	}
}

func TestVMService_GetVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("timeout")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.GetVM(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_GetVM_ParseError(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return json.RawMessage(`not json`), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.GetVM(context.Background(), 1)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestVMService_UpdateVM(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					if method != "vm.update" {
						t.Errorf("expected method vm.update, got %s", method)
					}
					slice, ok := params.([]any)
					if !ok {
						t.Fatal("expected []any params for update")
					}
					if len(slice) != 2 {
						t.Fatalf("expected 2 elements, got %d", len(slice))
					}
					id, ok := slice[0].(int64)
					if !ok || id != 1 {
						t.Errorf("expected id 1, got %v", slice[0])
					}
					return json.RawMessage(`{"id": 1}`), nil
				}
				// Re-read via GetVM
				if method != "vm.get_instance" {
					t.Errorf("expected method vm.get_instance for re-read, got %s", method)
				}
				return sampleVMJSON(), nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	vm, err := svc.UpdateVM(context.Background(), 1, UpdateVMOpts{
		Name:   "test-vm",
		Memory: 2048,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm == nil {
		t.Fatal("expected non-nil vm")
	}
	if vm.ID != 1 {
		t.Errorf("expected ID 1, got %d", vm.ID)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (update + re-read), got %d", callCount)
	}
}

func TestVMService_UpdateVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("not found")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.UpdateVM(context.Background(), 999, UpdateVMOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_UpdateVM_ReReadError(t *testing.T) {
	callCount := 0
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				callCount++
				if callCount == 1 {
					return json.RawMessage(`{"id": 1}`), nil
				}
				return nil, errors.New("re-read failed")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	_, err := svc.UpdateVM(context.Background(), 1, UpdateVMOpts{})
	if err == nil {
		t.Fatal("expected error on re-read")
	}
	if err.Error() != "re-read failed" {
		t.Errorf("expected 're-read failed', got %q", err.Error())
	}
}

func TestVMService_DeleteVM(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.delete" {
					t.Errorf("expected method vm.delete, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 5 {
					t.Errorf("expected id 5, got %v", params)
				}
				return nil, nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.DeleteVM(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMService_DeleteVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.DeleteVM(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_StartVM(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				if method != "vm.start" {
					t.Errorf("expected method vm.start, got %s", method)
				}
				id, ok := params.(int64)
				if !ok || id != 1 {
					t.Errorf("expected id 1, got %v", params)
				}
				return nil, nil
			},
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.StartVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify it used Call, not CallAndWait
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	if mock.calls[0].Method != "vm.start" {
		t.Errorf("expected method vm.start, got %s", mock.calls[0].Method)
	}
}

func TestVMService_StartVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		mockCaller: mockCaller{
			callFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
				return nil, errors.New("vm already running")
			},
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.StartVM(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_StopVM(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			if method != "vm.stop" {
				t.Errorf("expected method vm.stop, got %s", method)
			}
			slice, ok := params.([]any)
			if !ok {
				t.Fatal("expected []any params")
			}
			if len(slice) != 2 {
				t.Fatalf("expected 2 elements, got %d", len(slice))
			}
			id, ok := slice[0].(int64)
			if !ok || id != 1 {
				t.Errorf("expected id 1, got %v", slice[0])
			}
			stopParams, ok := slice[1].(map[string]any)
			if !ok {
				t.Fatal("expected map[string]any for stop params")
			}
			if stopParams["force"] != false {
				t.Errorf("expected force=false, got %v", stopParams["force"])
			}
			if stopParams["force_after_timeout"] != false {
				t.Errorf("expected force_after_timeout=false, got %v", stopParams["force_after_timeout"])
			}
			return nil, nil
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.StopVM(context.Background(), 1, StopVMOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify it used CallAndWait
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	if mock.calls[0].Method != "vm.stop" {
		t.Errorf("expected method vm.stop, got %s", mock.calls[0].Method)
	}
}

func TestVMService_StopVM_Error(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			return nil, errors.New("vm not running")
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.StopVM(context.Background(), 1, StopVMOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVMService_StopVM_ForceOptions(t *testing.T) {
	mock := &mockAsyncCaller{
		callAndWaitFunc: func(ctx context.Context, method string, params any) (json.RawMessage, error) {
			slice := params.([]any)
			stopParams := slice[1].(map[string]any)
			if stopParams["force"] != true {
				t.Errorf("expected force=true, got %v", stopParams["force"])
			}
			if stopParams["force_after_timeout"] != true {
				t.Errorf("expected force_after_timeout=true, got %v", stopParams["force_after_timeout"])
			}
			return nil, nil
		},
	}

	svc := NewVMService(mock, Version{})
	err := svc.StopVM(context.Background(), 1, StopVMOpts{
		Force:             true,
		ForceAfterTimeout: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
