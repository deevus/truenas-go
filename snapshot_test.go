package truenas

import (
	"testing"
)

func TestSnapshotResponse_HasHold(t *testing.T) {
	tests := []struct {
		name     string
		snapshot SnapshotResponse
		want     bool
	}{
		{
			name:     "no holds - empty string",
			snapshot: SnapshotResponse{Properties: SnapshotProperties{UserRefs: UserRefsProperty{Parsed: ""}}},
			want:     false,
		},
		{
			name:     "no holds - zero",
			snapshot: SnapshotResponse{Properties: SnapshotProperties{UserRefs: UserRefsProperty{Parsed: "0"}}},
			want:     false,
		},
		{
			name:     "has one hold",
			snapshot: SnapshotResponse{Properties: SnapshotProperties{UserRefs: UserRefsProperty{Parsed: "1"}}},
			want:     true,
		},
		{
			name:     "has multiple holds",
			snapshot: SnapshotResponse{Properties: SnapshotProperties{UserRefs: UserRefsProperty{Parsed: "3"}}},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.snapshot.HasHold(); got != tt.want {
				t.Errorf("HasHold() = %v, want %v", got, tt.want)
			}
		})
	}
}
