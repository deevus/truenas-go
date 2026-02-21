package truenas

// Snapshot method names (without prefix).
const (
	MethodSnapshotCreate  = "create"
	MethodSnapshotQuery   = "query"
	MethodSnapshotDelete  = "delete"
	MethodSnapshotHold    = "hold"
	MethodSnapshotRelease = "release"
	MethodSnapshotClone   = "clone"
)

// ResolveSnapshotMethod returns the full API method name for the given version.
// Pre-25.10 uses "zfs.snapshot.*", 25.10+ uses "pool.snapshot.*".
func ResolveSnapshotMethod(v Version, method string) string {
	prefix := "zfs.snapshot"
	if v.AtLeast(25, 10) {
		prefix = "pool.snapshot"
	}
	return prefix + "." + method
}

// SnapshotResponse represents a snapshot from the query API.
type SnapshotResponse struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`          // Full snapshot ID (dataset@name)
	SnapshotName string             `json:"snapshot_name"` // Just the name part after @
	Dataset      string             `json:"dataset"`
	Properties   SnapshotProperties `json:"properties"`
}

// SnapshotProperties contains ZFS properties for a snapshot.
type SnapshotProperties struct {
	CreateTXG  PropertyValue    `json:"createtxg"`
	Used       ParsedValue      `json:"used"`
	Referenced ParsedValue      `json:"referenced"`
	UserRefs   UserRefsProperty `json:"userrefs"`
}

// UserRefsProperty represents the userrefs ZFS property (hold count).
type UserRefsProperty struct {
	Parsed string `json:"parsed"` // String like "0" or "1"
}

// PropertyValue represents a ZFS property with a string value.
type PropertyValue struct {
	Value string `json:"value"`
}

// ParsedValue represents a ZFS property with a parsed numeric value.
type ParsedValue struct {
	Parsed int64 `json:"parsed"`
}

// HasHold returns true if the snapshot has any holds.
// It checks the userrefs property which indicates the number of user holds.
func (s *SnapshotResponse) HasHold() bool {
	return s.Properties.UserRefs.Parsed != "" && s.Properties.UserRefs.Parsed != "0"
}
