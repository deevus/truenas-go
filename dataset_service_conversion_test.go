package truenas

import "testing"

func TestDatasetFromResponse(t *testing.T) {
	resp := DatasetResponse{
		ID:          "pool1/ds1",
		Name:        "pool1/ds1",
		Pool:        "pool1",
		Type:        "FILESYSTEM",
		Mountpoint:  "/mnt/pool1/ds1",
		Comments:    PropertyValue{Value: "my dataset"},
		Compression: PropertyValue{Value: "zstd"},
		Quota:       SizePropertyField{Parsed: 2147483648, Value: "2G"},
		RefQuota:    SizePropertyField{Parsed: 1073741824, Value: "1G"},
		Atime:       PropertyValue{Value: "off"},
	}

	ds := datasetFromResponse(resp)

	if ds.ID != "pool1/ds1" {
		t.Errorf("expected ID pool1/ds1, got %s", ds.ID)
	}
	if ds.Name != "pool1/ds1" {
		t.Errorf("expected Name pool1/ds1, got %s", ds.Name)
	}
	if ds.Pool != "pool1" {
		t.Errorf("expected Pool pool1, got %s", ds.Pool)
	}
	if ds.Mountpoint != "/mnt/pool1/ds1" {
		t.Errorf("expected Mountpoint /mnt/pool1/ds1, got %s", ds.Mountpoint)
	}
	if ds.Comments != "my dataset" {
		t.Errorf("expected Comments 'my dataset', got %s", ds.Comments)
	}
	if ds.Compression != "zstd" {
		t.Errorf("expected Compression zstd, got %s", ds.Compression)
	}
	if ds.Quota != 2147483648 {
		t.Errorf("expected Quota 2147483648, got %d", ds.Quota)
	}
	if ds.RefQuota != 1073741824 {
		t.Errorf("expected RefQuota 1073741824, got %d", ds.RefQuota)
	}
	if ds.Atime != "off" {
		t.Errorf("expected Atime off, got %s", ds.Atime)
	}
}

func TestZvolFromResponse(t *testing.T) {
	resp := DatasetResponse{
		ID:           "pool1/zvol1",
		Name:         "pool1/zvol1",
		Pool:         "pool1",
		Type:         "VOLUME",
		Comments:     PropertyValue{Value: "my zvol"},
		Compression:  PropertyValue{Value: "lz4"},
		Volsize:      SizePropertyField{Parsed: 10737418240, Value: "10G"},
		Volblocksize: PropertyValue{Value: "16K"},
		Sparse:       PropertyValue{Value: "true"},
	}

	zvol := zvolFromResponse(resp)

	if zvol.ID != "pool1/zvol1" {
		t.Errorf("expected ID pool1/zvol1, got %s", zvol.ID)
	}
	if zvol.Name != "pool1/zvol1" {
		t.Errorf("expected Name pool1/zvol1, got %s", zvol.Name)
	}
	if zvol.Pool != "pool1" {
		t.Errorf("expected Pool pool1, got %s", zvol.Pool)
	}
	if zvol.Comments != "my zvol" {
		t.Errorf("expected Comments 'my zvol', got %s", zvol.Comments)
	}
	if zvol.Compression != "lz4" {
		t.Errorf("expected Compression lz4, got %s", zvol.Compression)
	}
	if zvol.Volsize != 10737418240 {
		t.Errorf("expected Volsize 10737418240, got %d", zvol.Volsize)
	}
	if zvol.Volblocksize != "16K" {
		t.Errorf("expected Volblocksize 16K, got %s", zvol.Volblocksize)
	}
	if !zvol.Sparse {
		t.Error("expected Sparse=true")
	}
}

func TestZvolFromResponse_SparseTrue(t *testing.T) {
	resp := DatasetResponse{
		Sparse: PropertyValue{Value: "true"},
	}
	zvol := zvolFromResponse(resp)
	if !zvol.Sparse {
		t.Error("expected Sparse=true for value 'true'")
	}
}

func TestZvolFromResponse_SparseFalse(t *testing.T) {
	resp := DatasetResponse{
		Sparse: PropertyValue{Value: "false"},
	}
	zvol := zvolFromResponse(resp)
	if zvol.Sparse {
		t.Error("expected Sparse=false for value 'false'")
	}
}

func TestPoolFromResponse(t *testing.T) {
	resp := PoolResponse{
		ID:   1,
		Name: "tank",
		Path: "/mnt/tank",
	}

	pool := poolFromResponse(resp)

	if pool.ID != 1 {
		t.Errorf("expected ID 1, got %d", pool.ID)
	}
	if pool.Name != "tank" {
		t.Errorf("expected Name tank, got %s", pool.Name)
	}
	if pool.Path != "/mnt/tank" {
		t.Errorf("expected Path /mnt/tank, got %s", pool.Path)
	}
}

func TestDatasetCreateParams(t *testing.T) {
	opts := CreateDatasetOpts{
		Name:        "pool1/ds1",
		Comments:    "test",
		Compression: "lz4",
		Quota:       1073741824,
		RefQuota:    536870912,
		Atime:       "on",
	}

	params := datasetCreateParams(opts)

	if params["name"] != "pool1/ds1" {
		t.Errorf("expected name pool1/ds1, got %v", params["name"])
	}
	if params["type"] != "FILESYSTEM" {
		t.Errorf("expected type FILESYSTEM, got %v", params["type"])
	}
	if params["comments"] != "test" {
		t.Errorf("expected comments test, got %v", params["comments"])
	}
	if params["compression"] != "lz4" {
		t.Errorf("expected compression lz4, got %v", params["compression"])
	}
	if params["quota"] != int64(1073741824) {
		t.Errorf("expected quota 1073741824, got %v", params["quota"])
	}
	if params["refquota"] != int64(536870912) {
		t.Errorf("expected refquota 536870912, got %v", params["refquota"])
	}
	if params["atime"] != "on" {
		t.Errorf("expected atime on, got %v", params["atime"])
	}
}

func TestDatasetCreateParams_Minimal(t *testing.T) {
	opts := CreateDatasetOpts{
		Name: "pool1/ds1",
	}

	params := datasetCreateParams(opts)

	if params["name"] != "pool1/ds1" {
		t.Errorf("expected name pool1/ds1, got %v", params["name"])
	}
	if params["type"] != "FILESYSTEM" {
		t.Errorf("expected type FILESYSTEM, got %v", params["type"])
	}
	if _, ok := params["comments"]; ok {
		t.Error("expected no comments key")
	}
	if _, ok := params["compression"]; ok {
		t.Error("expected no compression key")
	}
	if _, ok := params["quota"]; ok {
		t.Error("expected no quota key")
	}
	if _, ok := params["refquota"]; ok {
		t.Error("expected no refquota key")
	}
	if _, ok := params["atime"]; ok {
		t.Error("expected no atime key")
	}
}

func TestDatasetUpdateParams(t *testing.T) {
	opts := UpdateDatasetOpts{
		Compression: "zstd",
		Quota:       Int64Ptr(2147483648),
		RefQuota:    Int64Ptr(1073741824),
		Atime:       "off",
		Comments:    StringPtr("updated"),
	}

	params := datasetUpdateParams(opts)

	if params["compression"] != "zstd" {
		t.Errorf("expected compression zstd, got %v", params["compression"])
	}
	if params["quota"] != int64(2147483648) {
		t.Errorf("expected quota 2147483648, got %v", params["quota"])
	}
	if params["refquota"] != int64(1073741824) {
		t.Errorf("expected refquota 1073741824, got %v", params["refquota"])
	}
	if params["atime"] != "off" {
		t.Errorf("expected atime off, got %v", params["atime"])
	}
	if params["comments"] != "updated" {
		t.Errorf("expected comments updated, got %v", params["comments"])
	}
}

func TestDatasetUpdateParams_Empty(t *testing.T) {
	opts := UpdateDatasetOpts{}

	params := datasetUpdateParams(opts)

	if len(params) != 0 {
		t.Errorf("expected empty params, got %v", params)
	}
}

func TestDatasetUpdateParams_CompressionAndAtime(t *testing.T) {
	opts := UpdateDatasetOpts{
		Compression: "gzip",
		Atime:       "on",
	}

	params := datasetUpdateParams(opts)

	if params["compression"] != "gzip" {
		t.Errorf("expected compression gzip, got %v", params["compression"])
	}
	if params["atime"] != "on" {
		t.Errorf("expected atime on, got %v", params["atime"])
	}
	if _, ok := params["comments"]; ok {
		t.Error("expected no comments key when nil")
	}
	if _, ok := params["quota"]; ok {
		t.Error("expected no quota key when nil")
	}
	if _, ok := params["refquota"]; ok {
		t.Error("expected no refquota key when nil")
	}
	if len(params) != 2 {
		t.Errorf("expected 2 params, got %d", len(params))
	}
}

func TestZvolCreateParams(t *testing.T) {
	opts := CreateZvolOpts{
		Name:         "pool1/zvol1",
		Volsize:      10737418240,
		Volblocksize: "16K",
		Sparse:       true,
		ForceSize:    true,
		Compression:  "lz4",
		Comments:     "my zvol",
	}

	params := zvolCreateParams(opts)

	if params["name"] != "pool1/zvol1" {
		t.Errorf("expected name pool1/zvol1, got %v", params["name"])
	}
	if params["type"] != "VOLUME" {
		t.Errorf("expected type VOLUME, got %v", params["type"])
	}
	if params["volsize"] != int64(10737418240) {
		t.Errorf("expected volsize 10737418240, got %v", params["volsize"])
	}
	if params["volblocksize"] != "16K" {
		t.Errorf("expected volblocksize 16K, got %v", params["volblocksize"])
	}
	if params["sparse"] != true {
		t.Errorf("expected sparse true, got %v", params["sparse"])
	}
	if params["force_size"] != true {
		t.Errorf("expected force_size true, got %v", params["force_size"])
	}
	if params["compression"] != "lz4" {
		t.Errorf("expected compression lz4, got %v", params["compression"])
	}
	if params["comments"] != "my zvol" {
		t.Errorf("expected comments 'my zvol', got %v", params["comments"])
	}
}

func TestZvolCreateParams_Minimal(t *testing.T) {
	opts := CreateZvolOpts{
		Name:    "pool1/zvol1",
		Volsize: 10737418240,
	}

	params := zvolCreateParams(opts)

	if params["name"] != "pool1/zvol1" {
		t.Errorf("expected name pool1/zvol1, got %v", params["name"])
	}
	if params["type"] != "VOLUME" {
		t.Errorf("expected type VOLUME, got %v", params["type"])
	}
	if params["volsize"] != int64(10737418240) {
		t.Errorf("expected volsize 10737418240, got %v", params["volsize"])
	}
	if _, ok := params["volblocksize"]; ok {
		t.Error("expected no volblocksize key")
	}
	if _, ok := params["sparse"]; ok {
		t.Error("expected no sparse key")
	}
	if _, ok := params["force_size"]; ok {
		t.Error("expected no force_size key")
	}
	if _, ok := params["compression"]; ok {
		t.Error("expected no compression key")
	}
	if _, ok := params["comments"]; ok {
		t.Error("expected no comments key")
	}
}

func TestZvolUpdateParams(t *testing.T) {
	opts := UpdateZvolOpts{
		Volsize:     Int64Ptr(21474836480),
		ForceSize:   true,
		Compression: "zstd",
		Comments:    StringPtr("resized"),
	}

	params := zvolUpdateParams(opts)

	if params["volsize"] != int64(21474836480) {
		t.Errorf("expected volsize 21474836480, got %v", params["volsize"])
	}
	if params["force_size"] != true {
		t.Errorf("expected force_size true, got %v", params["force_size"])
	}
	if params["compression"] != "zstd" {
		t.Errorf("expected compression zstd, got %v", params["compression"])
	}
	if params["comments"] != "resized" {
		t.Errorf("expected comments resized, got %v", params["comments"])
	}
}

func TestZvolUpdateParams_Empty(t *testing.T) {
	opts := UpdateZvolOpts{}

	params := zvolUpdateParams(opts)

	if len(params) != 0 {
		t.Errorf("expected empty params, got %v", params)
	}
}

func TestZvolUpdateParams_ForceSizeOnly(t *testing.T) {
	opts := UpdateZvolOpts{
		ForceSize: true,
	}

	params := zvolUpdateParams(opts)

	if params["force_size"] != true {
		t.Errorf("expected force_size true, got %v", params["force_size"])
	}
	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
}

func TestInt64Ptr(t *testing.T) {
	p := Int64Ptr(42)
	if *p != 42 {
		t.Errorf("expected 42, got %d", *p)
	}

	z := Int64Ptr(0)
	if *z != 0 {
		t.Errorf("expected 0, got %d", *z)
	}
}

func TestStringPtr(t *testing.T) {
	p := StringPtr("hello")
	if *p != "hello" {
		t.Errorf("expected hello, got %s", *p)
	}

	e := StringPtr("")
	if *e != "" {
		t.Errorf("expected empty string, got %s", *e)
	}
}
