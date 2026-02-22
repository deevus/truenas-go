package truenas

// StatResponse represents a filesystem stat result from the TrueNAS API.
type StatResponse struct {
	Mode int64 `json:"mode"`
	UID  int64 `json:"uid"`
	GID  int64 `json:"gid"`
}
