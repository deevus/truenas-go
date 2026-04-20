package truenas

import (
	"context"
	"testing"
)

func TestMockGroupService_ImplementsInterface(t *testing.T) {
	var _ GroupServiceAPI = (*GroupService)(nil)
	var _ GroupServiceAPI = (*MockGroupService)(nil)
}

func TestMockGroupService_DefaultsToNil(t *testing.T) {
	mock := &MockGroupService{}
	ctx := context.Background()

	group, err := mock.Create(ctx, CreateGroupOpts{})
	if err != nil {
		t.Fatalf("expected nil error from Create, got: %v", err)
	}
	if group != nil {
		t.Fatalf("expected nil result from Create, got: %v", group)
	}

	group, err = mock.Get(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if group != nil {
		t.Fatalf("expected nil result, got: %v", group)
	}

	group, err = mock.GetByName(ctx, "test")
	if err != nil {
		t.Fatalf("expected nil error from GetByName, got: %v", err)
	}
	if group != nil {
		t.Fatalf("expected nil result from GetByName, got: %v", group)
	}

	group, err = mock.GetByGID(ctx, 1000)
	if err != nil {
		t.Fatalf("expected nil error from GetByGID, got: %v", err)
	}
	if group != nil {
		t.Fatalf("expected nil result from GetByGID, got: %v", group)
	}

	groups, err := mock.List(ctx)
	if err != nil {
		t.Fatalf("expected nil error from List, got: %v", err)
	}
	if groups != nil {
		t.Fatalf("expected nil result from List, got: %v", groups)
	}

	group, err = mock.Update(ctx, 1, UpdateGroupOpts{})
	if err != nil {
		t.Fatalf("expected nil error from Update, got: %v", err)
	}
	if group != nil {
		t.Fatalf("expected nil result from Update, got: %v", group)
	}

	err = mock.Delete(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error from Delete, got: %v", err)
	}
}

func TestMockGroupService_CallsFunc(t *testing.T) {
	ctx := context.Background()

	mock := &MockGroupService{
		CreateFunc: func(ctx context.Context, opts CreateGroupOpts) (*Group, error) {
			return &Group{ID: 1, Name: opts.Name}, nil
		},
		GetFunc: func(ctx context.Context, id int64) (*Group, error) {
			return &Group{ID: id}, nil
		},
		GetByNameFunc: func(ctx context.Context, name string) (*Group, error) {
			return &Group{Name: name}, nil
		},
		GetByGIDFunc: func(ctx context.Context, gid int64) (*Group, error) {
			return &Group{GID: gid}, nil
		},
		ListFunc: func(ctx context.Context) ([]Group, error) {
			return []Group{{ID: 1}}, nil
		},
		UpdateFunc: func(ctx context.Context, id int64, opts UpdateGroupOpts) (*Group, error) {
			return &Group{ID: id, Name: opts.Name}, nil
		},
		DeleteFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	group, err := mock.Create(ctx, CreateGroupOpts{Name: "test"})
	if err != nil || group.Name != "test" {
		t.Fatalf("Create: unexpected result: %v, %v", group, err)
	}

	group, err = mock.Get(ctx, 42)
	if err != nil || group.ID != 42 {
		t.Fatalf("Get: unexpected result: %v, %v", group, err)
	}

	group, err = mock.GetByName(ctx, "devs")
	if err != nil || group.Name != "devs" {
		t.Fatalf("GetByName: unexpected result: %v, %v", group, err)
	}

	group, err = mock.GetByGID(ctx, 5000)
	if err != nil || group.GID != 5000 {
		t.Fatalf("GetByGID: unexpected result: %v, %v", group, err)
	}

	groups, err := mock.List(ctx)
	if err != nil || len(groups) != 1 {
		t.Fatalf("List: unexpected result: %v, %v", groups, err)
	}

	group, err = mock.Update(ctx, 1, UpdateGroupOpts{Name: "new"})
	if err != nil || group.Name != "new" {
		t.Fatalf("Update: unexpected result: %v, %v", group, err)
	}

	if err := mock.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete: unexpected error: %v", err)
	}
}
