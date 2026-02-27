package truenas

import (
	"context"
	"testing"
)

func TestMockUserService_ImplementsInterface(t *testing.T) {
	var _ UserServiceAPI = (*UserService)(nil)
	var _ UserServiceAPI = (*MockUserService)(nil)
}

func TestMockUserService_DefaultsToNil(t *testing.T) {
	mock := &MockUserService{}
	ctx := context.Background()

	user, err := mock.Create(ctx, CreateUserOpts{})
	if err != nil {
		t.Fatalf("expected nil error from Create, got: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil result from Create, got: %v", user)
	}

	user, err = mock.Get(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil result, got: %v", user)
	}

	user, err = mock.GetByUsername(ctx, "test")
	if err != nil {
		t.Fatalf("expected nil error from GetByUsername, got: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil result from GetByUsername, got: %v", user)
	}

	user, err = mock.GetByUID(ctx, 1000)
	if err != nil {
		t.Fatalf("expected nil error from GetByUID, got: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil result from GetByUID, got: %v", user)
	}

	users, err := mock.List(ctx)
	if err != nil {
		t.Fatalf("expected nil error from List, got: %v", err)
	}
	if users != nil {
		t.Fatalf("expected nil result from List, got: %v", users)
	}

	user, err = mock.Update(ctx, 1, UpdateUserOpts{})
	if err != nil {
		t.Fatalf("expected nil error from Update, got: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil result from Update, got: %v", user)
	}

	err = mock.Delete(ctx, 1)
	if err != nil {
		t.Fatalf("expected nil error from Delete, got: %v", err)
	}
}

func TestMockUserService_CallsFunc(t *testing.T) {
	ctx := context.Background()

	mock := &MockUserService{
		CreateFunc: func(ctx context.Context, opts CreateUserOpts) (*User, error) {
			return &User{ID: 1, Username: opts.Username}, nil
		},
		GetFunc: func(ctx context.Context, id int64) (*User, error) {
			return &User{ID: id}, nil
		},
		GetByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
			return &User{Username: username}, nil
		},
		GetByUIDFunc: func(ctx context.Context, uid int64) (*User, error) {
			return &User{UID: uid}, nil
		},
		ListFunc: func(ctx context.Context) ([]User, error) {
			return []User{{ID: 1}}, nil
		},
		UpdateFunc: func(ctx context.Context, id int64, opts UpdateUserOpts) (*User, error) {
			return &User{ID: id, Username: opts.Username}, nil
		},
		DeleteFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	user, err := mock.Create(ctx, CreateUserOpts{Username: "test"})
	if err != nil || user.Username != "test" {
		t.Fatalf("Create: unexpected result: %v, %v", user, err)
	}

	user, err = mock.Get(ctx, 42)
	if err != nil || user.ID != 42 {
		t.Fatalf("Get: unexpected result: %v, %v", user, err)
	}

	user, err = mock.GetByUsername(ctx, "jdoe")
	if err != nil || user.Username != "jdoe" {
		t.Fatalf("GetByUsername: unexpected result: %v, %v", user, err)
	}

	user, err = mock.GetByUID(ctx, 1001)
	if err != nil || user.UID != 1001 {
		t.Fatalf("GetByUID: unexpected result: %v, %v", user, err)
	}

	users, err := mock.List(ctx)
	if err != nil || len(users) != 1 {
		t.Fatalf("List: unexpected result: %v, %v", users, err)
	}

	user, err = mock.Update(ctx, 1, UpdateUserOpts{Username: "new"})
	if err != nil || user.Username != "new" {
		t.Fatalf("Update: unexpected result: %v, %v", user, err)
	}

	if err := mock.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete: unexpected error: %v", err)
	}
}
