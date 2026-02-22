package truenas

// DatasetResponse represents a dataset from the pool.dataset.query API.
type DatasetResponse struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Pool         string            `json:"pool"`
	Type         string            `json:"type"`
	Mountpoint   string            `json:"mountpoint"`
	Comments     PropertyValue     `json:"comments"`
	Compression  PropertyValue     `json:"compression"`
	Quota        SizePropertyField `json:"quota"`
	RefQuota     SizePropertyField `json:"refquota"`
	Atime        PropertyValue     `json:"atime"`
	Volsize      SizePropertyField `json:"volsize"`
	Volblocksize PropertyValue     `json:"volblocksize"`
	Sparse       PropertyValue     `json:"sparse"`
}

// SizePropertyField represents a ZFS size property with a parsed numeric value and string representation.
type SizePropertyField struct {
	Parsed int64  `json:"parsed"`
	Value  string `json:"value"`
}

// DatasetCreateResponse represents the response from pool.dataset.create.
type DatasetCreateResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Mountpoint string `json:"mountpoint"`
}

// PoolResponse represents a pool from the pool.query API.
type PoolResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}
