package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/deevus/truenas-go/api"
)

func TestReceiverTypeName(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{`package p; func (s *FooService) M() {}`, "FooService"},
		{`package p; func (s FooService) M() {}`, "FooService"},
	}
	for _, tt := range tests {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "", tt.src, 0)
		if err != nil {
			t.Fatal(err)
		}
		fn := f.Decls[0].(*ast.FuncDecl)
		got := receiverTypeName(fn.Recv.List[0].Type)
		if got != tt.want {
			t.Errorf("receiverTypeName = %q, want %q", got, tt.want)
		}
	}
}

func TestReceiverTypeName_NonIdent(t *testing.T) {
	got := receiverTypeName(&ast.ArrayType{})
	if got != "" {
		t.Errorf("expected empty string for non-ident type, got %q", got)
	}
}

func TestExtractStringConstants(t *testing.T) {
	src := `package p
const (
	methodFoo = "foo"
	methodBar = "bar"
	intConst  = 42
)
const single = "single"
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	consts := extractStringConstants(f)
	if consts["methodFoo"] != "foo" {
		t.Errorf("methodFoo = %q, want %q", consts["methodFoo"], "foo")
	}
	if consts["methodBar"] != "bar" {
		t.Errorf("methodBar = %q, want %q", consts["methodBar"], "bar")
	}
	if _, ok := consts["intConst"]; ok {
		t.Error("intConst should not be in string constants")
	}
	if consts["single"] != "single" {
		t.Errorf("single = %q, want %q", consts["single"], "single")
	}
}

func TestExtractAPICalls_DirectLiteral(t *testing.T) {
	src := `package p
import "context"
type FooService struct { client interface{ Call(context.Context, string, any) (any, error) } }
func (s *FooService) DoStuff(ctx context.Context) error {
	_, err := s.client.Call(ctx, "foo.bar", nil)
	return err
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	methods := extractAPICalls(f)
	if len(methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(methods))
	}
	if methods[0].APIMethod != "foo.bar" {
		t.Errorf("APIMethod = %q, want %q", methods[0].APIMethod, "foo.bar")
	}
	if methods[0].ServiceStruct != "FooService" {
		t.Errorf("ServiceStruct = %q, want %q", methods[0].ServiceStruct, "FooService")
	}
	if methods[0].GoMethodName != "DoStuff" {
		t.Errorf("GoMethodName = %q, want %q", methods[0].GoMethodName, "DoStuff")
	}
}

func TestExtractAPICalls_CallAndWait(t *testing.T) {
	src := `package p
import "context"
type BarService struct { client interface{ CallAndWait(context.Context, string, any) (any, error) } }
func (s *BarService) Run(ctx context.Context) error {
	_, err := s.client.CallAndWait(ctx, "bar.run", nil)
	return err
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	methods := extractAPICalls(f)
	if len(methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(methods))
	}
	if methods[0].APIMethod != "bar.run" {
		t.Errorf("APIMethod = %q, want %q", methods[0].APIMethod, "bar.run")
	}
}

func TestExtractAPICalls_Subscribe(t *testing.T) {
	src := `package p
import "context"
type SubService struct { client interface{ Subscribe(context.Context, string, any) (any, error) } }
func (s *SubService) Watch(ctx context.Context) error {
	_, err := s.client.Subscribe(ctx, "sub.events", nil)
	return err
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	methods := extractAPICalls(f)
	if len(methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(methods))
	}
	if methods[0].APIMethod != "sub.events" {
		t.Errorf("APIMethod = %q, want %q", methods[0].APIMethod, "sub.events")
	}
}

func TestExtractAPICalls_SkipsUnexported(t *testing.T) {
	src := `package p
import "context"
type FooService struct { client interface{ Call(context.Context, string, any) (any, error) } }
func (s *FooService) doPrivate(ctx context.Context) error {
	_, err := s.client.Call(ctx, "foo.private", nil)
	return err
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	methods := extractAPICalls(f)
	if len(methods) != 0 {
		t.Errorf("expected 0 methods for unexported, got %d", len(methods))
	}
}

func TestExtractAPICalls_SkipsNonService(t *testing.T) {
	src := `package p
import "context"
type Helper struct { client interface{ Call(context.Context, string, any) (any, error) } }
func (h *Helper) DoStuff(ctx context.Context) error {
	_, err := h.client.Call(ctx, "helper.stuff", nil)
	return err
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	methods := extractAPICalls(f)
	if len(methods) != 0 {
		t.Errorf("expected 0 methods for non-service, got %d", len(methods))
	}
}

func TestExtractAPICalls_Resolver(t *testing.T) {
	src := `package p
import "context"

const (
	methodCreate = "create"
	methodQuery  = "query"
)

func resolveMethod(v int, method string) string {
	prefix := "old.ns"
	if v > 10 {
		prefix = "new.ns"
	}
	return prefix + "." + method
}

type MyService struct {
	client  interface{ Call(context.Context, string, any) (any, error) }
	version int
}

func (s *MyService) Create(ctx context.Context) error {
	method := resolveMethod(s.version, methodCreate)
	_, err := s.client.Call(ctx, method, nil)
	return err
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	methods := extractAPICalls(f)
	if len(methods) != 2 {
		t.Fatalf("expected 2 methods (both prefixes), got %d: %+v", len(methods), methods)
	}

	apiMethods := make(map[string]bool)
	for _, m := range methods {
		apiMethods[m.APIMethod] = true
	}
	if !apiMethods["old.ns.create"] {
		t.Error("expected old.ns.create")
	}
	if !apiMethods["new.ns.create"] {
		t.Error("expected new.ns.create")
	}
}

func TestExtractResolverPrefixes(t *testing.T) {
	src := `package p

func resolveSnapshotMethod(v int, method string) string {
	prefix := "zfs.snapshot"
	if v > 10 {
		prefix = "pool.snapshot"
	}
	return prefix + "." + method
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	resolvers := extractResolverPrefixes(f)
	prefixes, ok := resolvers["resolveSnapshotMethod"]
	if !ok {
		t.Fatal("expected resolveSnapshotMethod in resolvers")
	}
	if len(prefixes) != 2 {
		t.Fatalf("expected 2 prefixes, got %d: %v", len(prefixes), prefixes)
	}
}

func TestResolveArgToString(t *testing.T) {
	consts := map[string]string{"methodFoo": "foo"}

	// String literal
	lit := &ast.BasicLit{Kind: token.STRING, Value: `"bar"`}
	if got := resolveArgToString(lit, consts); got != "bar" {
		t.Errorf("literal: got %q, want %q", got, "bar")
	}

	// Constant reference
	ident := &ast.Ident{Name: "methodFoo"}
	if got := resolveArgToString(ident, consts); got != "foo" {
		t.Errorf("const ref: got %q, want %q", got, "foo")
	}

	// Unknown ident
	unknown := &ast.Ident{Name: "unknown"}
	if got := resolveArgToString(unknown, consts); got != "" {
		t.Errorf("unknown: got %q, want empty", got)
	}

	// Non-string literal
	intLit := &ast.BasicLit{Kind: token.INT, Value: "42"}
	if got := resolveArgToString(intLit, consts); got != "" {
		t.Errorf("int lit: got %q, want empty", got)
	}
}

func TestPct(t *testing.T) {
	if got := pct(1, 2); got != 50 {
		t.Errorf("pct(1,2) = %f, want 50", got)
	}
	if got := pct(0, 0); got != 0 {
		t.Errorf("pct(0,0) = %f, want 0", got)
	}
	if got := pct(3, 3); got != 100 {
		t.Errorf("pct(3,3) = %f, want 100", got)
	}
}

func TestMatchTestCount(t *testing.T) {
	testFuncs := []string{
		"TestFooService_Create",
		"TestFooService_Create_Error",
		"TestFooService_Create_ParseError",
		"TestFooService_CreateZvol",
		"TestFooService_Get",
		"TestBarService_Create",
	}

	tests := []struct {
		structName string
		methodName string
		want       int
	}{
		{"FooService", "Create", 3},
		{"FooService", "CreateZvol", 1},
		{"FooService", "Get", 1},
		{"FooService", "Delete", 0},
		{"BarService", "Create", 1},
		{"BazService", "Create", 0},
	}
	for _, tt := range tests {
		t.Run(tt.structName+"_"+tt.methodName, func(t *testing.T) {
			got := matchTestCount(testFuncs, tt.structName, tt.methodName)
			if got != tt.want {
				t.Errorf("matchTestCount(%s, %s) = %d, want %d", tt.structName, tt.methodName, got, tt.want)
			}
		})
	}
}

func TestBuildAPIMapping(t *testing.T) {
	goMethods := []goMethod{
		{ServiceStruct: "AppService", GoMethodName: "GetApp", APIMethod: "app.query"},
		{ServiceStruct: "AppService", GoMethodName: "ListApps", APIMethod: "app.query"},
		{ServiceStruct: "AppService", GoMethodName: "CreateApp", APIMethod: "app.create"},
	}

	m := buildAPIMapping(goMethods)

	// app.query should have 2 Go methods
	if mapping, ok := m["app.query"]; !ok {
		t.Fatal("expected app.query in mapping")
	} else {
		if len(mapping.goMethods) != 2 {
			t.Errorf("expected 2 Go methods for app.query, got %d", len(mapping.goMethods))
		}
		names := mapping.goMethodNames()
		if !strings.Contains(names, "GetApp") || !strings.Contains(names, "ListApps") {
			t.Errorf("expected GetApp and ListApps in names, got %q", names)
		}
	}

	// app.create should have 1 Go method
	if mapping, ok := m["app.create"]; !ok {
		t.Fatal("expected app.create in mapping")
	} else if len(mapping.goMethods) != 1 {
		t.Errorf("expected 1 Go method for app.create, got %d", len(mapping.goMethods))
	}
}

func TestBuildAPIMapping_SnapshotAliases(t *testing.T) {
	goMethods := []goMethod{
		{ServiceStruct: "SnapshotService", GoMethodName: "Create", APIMethod: "zfs.snapshot.create"},
	}

	m := buildAPIMapping(goMethods)

	// Both zfs.snapshot.create and pool.snapshot.create should be mapped
	if _, ok := m["zfs.snapshot.create"]; !ok {
		t.Error("expected zfs.snapshot.create in mapping")
	}
	if _, ok := m["pool.snapshot.create"]; !ok {
		t.Error("expected pool.snapshot.create alias in mapping")
	}
}

func TestBuildAPIMapping_Deduplication(t *testing.T) {
	goMethods := []goMethod{
		{ServiceStruct: "FooService", GoMethodName: "Get", APIMethod: "foo.query"},
		{ServiceStruct: "FooService", GoMethodName: "Get", APIMethod: "foo.query"},
	}

	m := buildAPIMapping(goMethods)
	if len(m["foo.query"].goMethods) != 1 {
		t.Errorf("expected deduplication, got %d methods", len(m["foo.query"].goMethods))
	}
}

func TestAPIMethodMapping_TotalTestCount(t *testing.T) {
	mapping := &apiMethodMapping{
		goMethods: []goMethod{
			{ServiceStruct: "FooService", GoMethodName: "Get"},
			{ServiceStruct: "FooService", GoMethodName: "List"},
		},
	}
	testFuncs := []string{
		"TestFooService_Get",
		"TestFooService_Get_Error",
		"TestFooService_List",
	}

	got := mapping.totalTestCount(testFuncs)
	if got != 3 {
		t.Errorf("totalTestCount = %d, want 3", got)
	}
}

func TestWriteMatrix_BasicOutput(t *testing.T) {
	apiMethods := map[string]api.MethodDef{
		"test.create": {},
		"test.query":  {},
		"test.delete": {},
		"other.foo":   {},
	}
	goMethods := []goMethod{
		{ServiceStruct: "TestService", GoMethodName: "Create", APIMethod: "test.create"},
		{ServiceStruct: "TestService", GoMethodName: "Get", APIMethod: "test.query"},
		{ServiceStruct: "TestService", GoMethodName: "List", APIMethod: "test.query"},
	}
	testFuncs := []string{
		"TestTestService_Create",
		"TestTestService_Create_Error",
		"TestTestService_Get",
	}

	var buf bytes.Buffer
	writeMatrix(&buf, "99.99", apiMethods, goMethods, testFuncs)
	output := buf.String()

	// Check header
	if !strings.Contains(output, "TrueNAS version: 99.99") {
		t.Error("expected version in output")
	}
	if !strings.Contains(output, "Total API methods: 4") {
		t.Error("expected total API methods count")
	}
	if !strings.Contains(output, "Implemented: 2 (50.0%)") {
		t.Errorf("expected implemented percentage, got header: %s", strings.SplitN(output, "\n", 8)[4])
	}
	if !strings.Contains(output, "100.0% of implemented") {
		t.Error("expected tested percentage of implemented")
	}

	// Check covered namespace
	if !strings.Contains(output, "TestService") {
		t.Error("expected TestService in output")
	}

	// Check detail table
	if !strings.Contains(output, "test.create") {
		t.Error("expected test.create in output")
	}
	if !strings.Contains(output, "Get, List") {
		t.Error("expected 'Get, List' for test.query")
	}

	// Check uncovered namespace
	if !strings.Contains(output, "Uncovered Namespaces") {
		t.Error("expected uncovered namespaces section")
	}
	if !strings.Contains(output, "other") {
		t.Error("expected 'other' namespace in uncovered section")
	}
}

func TestWriteMatrix_UnmappedMethods(t *testing.T) {
	apiMethods := map[string]api.MethodDef{
		"test.create": {},
	}
	goMethods := []goMethod{
		{ServiceStruct: "TestService", GoMethodName: "Create", APIMethod: "test.create"},
		{ServiceStruct: "TestService", GoMethodName: "SubscribeEvents", APIMethod: "test.events"},
	}

	var buf bytes.Buffer
	writeMatrix(&buf, "1.0", apiMethods, goMethods, nil)
	output := buf.String()

	if !strings.Contains(output, "Go Methods Not in API Schema") {
		t.Error("expected unmapped methods section")
	}
	if !strings.Contains(output, "test.events") {
		t.Error("expected test.events in unmapped section")
	}
	if !strings.Contains(output, "SubscribeEvents") {
		t.Error("expected SubscribeEvents in unmapped section")
	}
}

func TestWriteMatrix_NoUncovered(t *testing.T) {
	apiMethods := map[string]api.MethodDef{
		"ns.create": {},
	}
	goMethods := []goMethod{
		{ServiceStruct: "NsService", GoMethodName: "Create", APIMethod: "ns.create"},
	}

	var buf bytes.Buffer
	writeMatrix(&buf, "1.0", apiMethods, goMethods, nil)
	output := buf.String()

	if strings.Contains(output, "Uncovered Namespaces") {
		t.Error("should not have uncovered namespaces section when all are covered")
	}
}

func TestScanGoMethods_InvalidDir(t *testing.T) {
	_, err := scanGoMethods("/nonexistent/path")
	if err == nil {
		t.Error("expected error for invalid directory")
	}
}

func TestScanTestFunctions_InvalidDir(t *testing.T) {
	_, err := scanTestFunctions("/nonexistent/path")
	if err == nil {
		t.Error("expected error for invalid directory")
	}
}

func TestScanGoMethods_TempDir(t *testing.T) {
	dir := t.TempDir()

	// Write a service file
	svcSrc := `package truenas
import "context"
type FooService struct { client interface{ Call(context.Context, string, any) (any, error) } }
func (s *FooService) Get(ctx context.Context) (any, error) {
	return s.client.Call(ctx, "foo.get_instance", nil)
}
func (s *FooService) List(ctx context.Context) (any, error) {
	return s.client.Call(ctx, "foo.query", nil)
}
`
	if err := writeFile(dir, "foo_service.go", svcSrc); err != nil {
		t.Fatal(err)
	}

	// Write an iface file (should be ignored)
	ifaceSrc := `package truenas
type FooServiceAPI interface { Get() }
`
	if err := writeFile(dir, "foo_service_iface.go", ifaceSrc); err != nil {
		t.Fatal(err)
	}

	methods, err := scanGoMethods(dir)
	if err != nil {
		t.Fatalf("scanGoMethods error: %v", err)
	}
	if len(methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(methods))
	}

	apiMethods := make(map[string]bool)
	for _, m := range methods {
		apiMethods[m.APIMethod] = true
	}
	if !apiMethods["foo.get_instance"] {
		t.Error("expected foo.get_instance")
	}
	if !apiMethods["foo.query"] {
		t.Error("expected foo.query")
	}
}

func TestScanGoMethods_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	methods, err := scanGoMethods(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(methods) != 0 {
		t.Errorf("expected 0 methods, got %d", len(methods))
	}
}

func TestScanTestFunctions_TempDir(t *testing.T) {
	dir := t.TempDir()

	testSrc := `package truenas
import "testing"
func TestFooService_Get(t *testing.T) {}
func TestFooService_Get_Error(t *testing.T) {}
func TestFooFromResponse(t *testing.T) {}
`
	if err := writeFile(dir, "foo_service_test.go", testSrc); err != nil {
		t.Fatal(err)
	}

	// Subscribe test file should also be picked up
	subTestSrc := `package truenas
import "testing"
func TestFooService_Subscribe(t *testing.T) {}
`
	if err := writeFile(dir, "foo_service_subscribe_test.go", subTestSrc); err != nil {
		t.Fatal(err)
	}

	// Iface test file should be excluded
	ifaceTestSrc := `package truenas
import "testing"
func TestMockFooService_Implements(t *testing.T) {}
`
	if err := writeFile(dir, "foo_service_iface_test.go", ifaceTestSrc); err != nil {
		t.Fatal(err)
	}

	funcs, err := scanTestFunctions(dir)
	if err != nil {
		t.Fatalf("scanTestFunctions error: %v", err)
	}
	if len(funcs) != 4 {
		t.Fatalf("expected 4 test functions (excl iface), got %d: %v", len(funcs), funcs)
	}

	funcSet := make(map[string]bool)
	for _, f := range funcs {
		funcSet[f] = true
	}
	if funcSet["TestMockFooService_Implements"] {
		t.Error("iface_test.go should have been excluded")
	}
}

func TestRun_TempDir(t *testing.T) {
	dir := t.TempDir()

	// Use a real API method from methods.json so it shows up in covered namespaces
	svcSrc := `package truenas
import "context"
type SysService struct { client interface{ Call(context.Context, string, any) (any, error) } }
func (s *SysService) GetInfo(ctx context.Context) (any, error) {
	return s.client.Call(ctx, "system.info", nil)
}
`
	if err := writeFile(dir, "sys_service.go", svcSrc); err != nil {
		t.Fatal(err)
	}

	testSrc := `package truenas
import "testing"
func TestSysService_GetInfo(t *testing.T) {}
`
	if err := writeFile(dir, "sys_service_test.go", testSrc); err != nil {
		t.Fatal(err)
	}

	outFile := dir + "/output.md"
	err := run(dir, outFile, "25.04")
	if err != nil {
		t.Fatalf("run error: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "SysService") {
		t.Error("expected SysService in output")
	}
	if !strings.Contains(output, "system.info") {
		t.Error("expected system.info in output")
	}
}

func TestRun_InvalidVersion(t *testing.T) {
	dir := t.TempDir()
	err := run(dir, "", "99.99")
	if err == nil {
		t.Error("expected error for invalid version")
	}
}

func TestRun_InvalidOutputPath(t *testing.T) {
	dir := t.TempDir()
	err := run(dir, "/nonexistent/dir/output.md", "25.04")
	if err == nil {
		t.Error("expected error for invalid output path")
	}
}

func writeFile(dir, name, content string) error {
	return os.WriteFile(dir+"/"+name, []byte(content), 0644)
}
