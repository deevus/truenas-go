// Command featurematrix generates a markdown feature matrix comparing
// implemented Go service methods against the full TrueNAS API surface.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/deevus/truenas-go/api"
)

// goMethod records a mapping from a TrueNAS API method to the Go code that calls it.
type goMethod struct {
	ServiceStruct string // e.g. "SnapshotService"
	GoMethodName  string // e.g. "Create"
	APIMethod     string // e.g. "zfs.snapshot.create"
}

// serviceMethodInfo combines API and Go information for one service method.
type serviceMethodInfo struct {
	APIMethod   string
	GoMethod    string
	Implemented bool
	TestCount   int
}

// serviceGroup holds all methods for one Go service, keyed by namespace.
type serviceGroup struct {
	ServiceStruct string
	Namespaces    []string
	Methods       []serviceMethodInfo
}

func main() {
	dir := flag.String("dir", ".", "project root directory")
	output := flag.String("o", "", "output file path (default: stdout)")
	version := flag.String("version", "", "TrueNAS version (default: latest embedded)")
	flag.Parse()

	if err := run(*dir, *output, *version); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(dir, output, version string) error {
	if version == "" {
		version = api.LatestVersion()
	}
	if version == "" {
		return fmt.Errorf("no embedded API versions found")
	}

	apiMethods, err := api.Methods(version)
	if err != nil {
		return err
	}

	goMethods, err := scanGoMethods(dir)
	if err != nil {
		return fmt.Errorf("scanning go source: %w", err)
	}

	testFuncs, err := scanTestFunctions(dir)
	if err != nil {
		return fmt.Errorf("scanning tests: %w", err)
	}

	var w io.Writer = os.Stdout
	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	writeMatrix(w, version, apiMethods, goMethods, testFuncs)
	return nil
}

// scanGoMethods parses *_service.go files and extracts API method calls.
func scanGoMethods(dir string) ([]goMethod, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, "_service.go") &&
			!strings.HasSuffix(name, "_iface.go") &&
			!strings.HasSuffix(name, "_test.go") {
			files = append(files, filepath.Join(dir, name))
		}
	}

	var results []goMethod
	fset := token.NewFileSet()

	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}

		methods := extractAPICalls(f)
		results = append(results, methods...)
	}

	return results, nil
}

// extractAPICalls walks an AST file and finds Call/CallAndWait/Subscribe calls
// with string literal API method names. Also handles indirect method resolution
// (e.g. resolveSnapshotMethod) by scanning for resolver function calls and
// extracting the method suffix from constants.
func extractAPICalls(f *ast.File) []goMethod {
	// First pass: collect string constants (e.g. methodSnapshotCreate = "create")
	constants := extractStringConstants(f)

	// Second pass: find resolver patterns (e.g. resolveSnapshotMethod)
	resolvers := extractResolverPrefixes(f)

	var results []goMethod

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}

		recvType := receiverTypeName(fn.Recv.List[0].Type)
		if recvType == "" || !strings.HasSuffix(recvType, "Service") {
			continue
		}

		methodName := fn.Name.Name
		if !ast.IsExported(methodName) {
			continue
		}

		// Track resolved method variables: varName → []apiMethod (multiple prefixes)
		resolvedVars := make(map[string][]string)

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			// Check for resolver calls: method := resolveXxxMethod(v, constant)
			if assign, ok := n.(*ast.AssignStmt); ok {
				for i, rhs := range assign.Rhs {
					if call, ok := rhs.(*ast.CallExpr); ok {
						if ident, ok := call.Fun.(*ast.Ident); ok {
							if prefixes, found := resolvers[ident.Name]; found && len(call.Args) >= 2 {
								suffix := resolveArgToString(call.Args[1], constants)
								if suffix != "" && i < len(assign.Lhs) {
									if lhsIdent, ok := assign.Lhs[i].(*ast.Ident); ok {
										var apiMethods []string
										for _, prefix := range prefixes {
											apiMethods = append(apiMethods, prefix+"."+suffix)
										}
										resolvedVars[lhsIdent.Name] = apiMethods
									}
								}
							}
						}
					}
				}
			}

			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			funcName := sel.Sel.Name
			if funcName != "Call" && funcName != "CallAndWait" && funcName != "Subscribe" {
				return true
			}

			if len(call.Args) < 2 {
				return true
			}

			// Try string literal first
			if lit, ok := call.Args[1].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				apiMethod := strings.Trim(lit.Value, `"`)
				results = append(results, goMethod{
					ServiceStruct: recvType,
					GoMethodName:  methodName,
					APIMethod:     apiMethod,
				})
				return true
			}

			// Try resolved variable (may have multiple prefixes)
			if ident, ok := call.Args[1].(*ast.Ident); ok {
				if apiMethods, found := resolvedVars[ident.Name]; found {
					for _, apiMethod := range apiMethods {
						results = append(results, goMethod{
							ServiceStruct: recvType,
							GoMethodName:  methodName,
							APIMethod:     apiMethod,
						})
					}
				}
			}

			return true
		})
	}

	return results
}

// extractStringConstants collects top-level const string values from a file.
func extractStringConstants(f *ast.File) map[string]string {
	consts := make(map[string]string)
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
				continue
			}
			for i, name := range vs.Names {
				if i < len(vs.Values) {
					if lit, ok := vs.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						consts[name.Name] = strings.Trim(lit.Value, `"`)
					}
				}
			}
		}
	}
	return consts
}

// extractResolverPrefixes finds functions like resolveSnapshotMethod that build
// API method names from a prefix + suffix. Returns a map of function name to
// all possible prefixes.
func extractResolverPrefixes(f *ast.File) map[string][]string {
	resolvers := make(map[string][]string)

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}

		if !strings.HasPrefix(fn.Name.Name, "resolve") || !strings.Contains(fn.Name.Name, "Method") {
			continue
		}

		// Collect all string prefixes used in prefix + "." + method patterns
		var prefixes []string
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for _, rhs := range assign.Rhs {
				if lit, ok := rhs.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					val := strings.Trim(lit.Value, `"`)
					if val != "" {
						prefixes = append(prefixes, val)
					}
				}
			}
			return true
		})

		if len(prefixes) > 0 {
			resolvers[fn.Name.Name] = prefixes
		}
	}

	return resolvers
}

// resolveArgToString attempts to resolve a function argument to a string value,
// either from a string literal or a constant reference.
func resolveArgToString(expr ast.Expr, constants map[string]string) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			return strings.Trim(e.Value, `"`)
		}
	case *ast.Ident:
		if val, ok := constants[e.Name]; ok {
			return val
		}
	}
	return ""
}

// receiverTypeName extracts the type name from a receiver expression,
// handling both value and pointer receivers.
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// scanTestFunctions finds all Test* functions in service test files.
func scanTestFunctions(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, "_test.go") &&
			strings.Contains(name, "_service") &&
			!strings.Contains(name, "_iface_test.go") {
			files = append(files, filepath.Join(dir, name))
		}
	}

	var results []string
	fset := token.NewFileSet()

	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			if strings.HasPrefix(fn.Name.Name, "Test") {
				results = append(results, fn.Name.Name)
			}
		}
	}

	return results, nil
}

// pct returns the percentage of n out of total, or 0 if total is 0.
func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}

// matchTestCount counts tests matching a service struct and method name.
func matchTestCount(testFuncs []string, structName, methodName string) int {
	prefix := "Test" + structName + "_" + methodName
	count := 0
	for _, name := range testFuncs {
		if name == prefix || strings.HasPrefix(name, prefix+"_") {
			count++
		}
	}
	return count
}

// apiMethodMapping holds all Go methods that call a given API method.
type apiMethodMapping struct {
	goMethods []goMethod
	service   string
}

// buildAPIMapping creates a map from API method names to the Go methods that call them.
// Handles one-to-many: multiple Go methods can call the same API method (e.g. app.query → GetApp, ListApps).
func buildAPIMapping(goMethods []goMethod) map[string]*apiMethodMapping {
	m := make(map[string]*apiMethodMapping)

	addMapping := func(apiMethod string, gm goMethod) {
		if existing, ok := m[apiMethod]; ok {
			// Deduplicate by Go method name
			for _, eg := range existing.goMethods {
				if eg.GoMethodName == gm.GoMethodName {
					return
				}
			}
			existing.goMethods = append(existing.goMethods, gm)
		} else {
			m[apiMethod] = &apiMethodMapping{
				goMethods: []goMethod{gm},
				service:   gm.ServiceStruct,
			}
		}
	}

	for _, gm := range goMethods {
		addMapping(gm.APIMethod, gm)

		// Snapshot aliases: zfs.snapshot.* → pool.snapshot.*
		if strings.HasPrefix(gm.APIMethod, "zfs.snapshot.") {
			alias := "pool.snapshot." + strings.TrimPrefix(gm.APIMethod, "zfs.snapshot.")
			addMapping(alias, gm)
		}
	}

	return m
}

// goMethodNames returns a comma-separated list of Go method names.
func (m *apiMethodMapping) goMethodNames() string {
	names := make([]string, len(m.goMethods))
	for i, gm := range m.goMethods {
		names[i] = gm.GoMethodName
	}
	return strings.Join(names, ", ")
}

// totalTestCount returns the total number of tests across all Go methods.
func (m *apiMethodMapping) totalTestCount(testFuncs []string) int {
	total := 0
	for _, gm := range m.goMethods {
		total += matchTestCount(testFuncs, gm.ServiceStruct, gm.GoMethodName)
	}
	return total
}

// writeMatrix generates the markdown feature matrix.
func writeMatrix(w io.Writer, version string, apiMethods map[string]api.MethodDef, goMethods []goMethod, testFuncs []string) {
	apiToGo := buildAPIMapping(goMethods)

	// Group API methods by namespace
	type nsInfo struct {
		namespace string
		methods   []string
	}
	nsMap := make(map[string][]string)
	for method := range apiMethods {
		ns := api.Namespace(method)
		nsMap[ns] = append(nsMap[ns], method)
	}

	// Sort methods within each namespace
	for ns := range nsMap {
		sort.Strings(nsMap[ns])
	}

	// Determine which namespaces are covered (have at least one Go implementation)
	coveredNS := make(map[string]string) // namespace → service struct
	for method, mapping := range apiToGo {
		if _, inAPI := apiMethods[method]; inAPI {
			ns := api.Namespace(method)
			coveredNS[ns] = mapping.service
		}
	}

	// Also mark namespaces covered if a Go method targets a method that resolves there
	for _, gm := range goMethods {
		ns := api.Namespace(gm.APIMethod)
		if _, exists := coveredNS[ns]; !exists {
			if _, exists := nsMap[ns]; exists {
				coveredNS[ns] = gm.ServiceStruct
			}
		}
	}

	// Group covered namespaces by service
	serviceNS := make(map[string][]string)
	for ns, svc := range coveredNS {
		serviceNS[svc] = append(serviceNS[svc], ns)
	}
	for svc := range serviceNS {
		sort.Strings(serviceNS[svc])
	}

	// Sort services
	var services []string
	for svc := range serviceNS {
		services = append(services, svc)
	}
	sort.Strings(services)

	// Count totals
	totalAPI := len(apiMethods)
	totalImpl := 0
	totalTested := 0
	for method := range apiMethods {
		if _, ok := apiToGo[method]; ok {
			totalImpl++
		}
	}
	for method := range apiMethods {
		if mapping, ok := apiToGo[method]; ok {
			if mapping.totalTestCount(testFuncs) > 0 {
				totalTested++
			}
		}
	}

	// Write header
	implPct := pct(totalImpl, totalAPI)
	testedPct := pct(totalTested, totalImpl)
	fmt.Fprintf(w, "# TrueNAS API Feature Matrix\n\n")
	fmt.Fprintf(w, "TrueNAS version: %s\n\n", version)
	fmt.Fprintf(w, "Total API methods: %d | Implemented: %d (%.1f%%) | Tested: %d (%.1f%% of implemented)\n\n", totalAPI, totalImpl, implPct, totalTested, testedPct)

	// Write covered summary table
	fmt.Fprintf(w, "## Covered Namespaces\n\n")
	fmt.Fprintf(w, "| Go Service | Namespaces | API Methods | Implemented | Tested |\n")
	fmt.Fprintf(w, "|------------|------------|:-----------:|:-----------:|:------:|\n")

	for _, svc := range services {
		namespaces := serviceNS[svc]
		nsDisplay := strings.Join(namespaces, ", ")

		apiCount := 0
		implCount := 0
		testedCount := 0
		for _, ns := range namespaces {
			for _, method := range nsMap[ns] {
				apiCount++
				if mapping, ok := apiToGo[method]; ok {
					implCount++
					if mapping.totalTestCount(testFuncs) > 0 {
						testedCount++
					}
				}
			}
		}

		fmt.Fprintf(w, "| %s | %s | %d | %d (%.0f%%) | %d (%.0f%%) |\n",
			svc, nsDisplay, apiCount,
			implCount, pct(implCount, apiCount),
			testedCount, pct(testedCount, implCount))
	}

	fmt.Fprintln(w)

	// Write per-service detail tables
	for _, svc := range services {
		namespaces := serviceNS[svc]

		for _, ns := range namespaces {
			methods := nsMap[ns]

			fmt.Fprintf(w, "### %s — `%s` (%d methods)\n\n", svc, ns, len(methods))
			fmt.Fprintf(w, "| API Method | Implemented | Go Method | Tested | Tests |\n")
			fmt.Fprintf(w, "|------------|:-----------:|-----------|:------:|------:|\n")

			for _, method := range methods {
				mapping, impl := apiToGo[method]
				implStr := ""
				goMethodStr := ""
				testedStr := ""
				testCountStr := ""

				if impl {
					implStr = "✓"
					goMethodStr = mapping.goMethodNames()
					tc := mapping.totalTestCount(testFuncs)
					if tc > 0 {
						testedStr = "✓"
						testCountStr = fmt.Sprintf("%d", tc)
					}
				}

				fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n", method, implStr, goMethodStr, testedStr, testCountStr)
			}

			fmt.Fprintln(w)
		}
	}

	// Write uncovered namespaces
	var uncovered []nsInfo
	for ns, methods := range nsMap {
		if _, covered := coveredNS[ns]; !covered {
			uncovered = append(uncovered, nsInfo{namespace: ns, methods: methods})
		}
	}
	sort.Slice(uncovered, func(i, j int) bool {
		return uncovered[i].namespace < uncovered[j].namespace
	})

	if len(uncovered) > 0 {
		uncoveredMethodCount := 0
		for _, ns := range uncovered {
			uncoveredMethodCount += len(ns.methods)
		}

		fmt.Fprintf(w, "## Uncovered Namespaces (%d namespaces, %d methods)\n\n", len(uncovered), uncoveredMethodCount)
		fmt.Fprintf(w, "| Namespace | Methods |\n")
		fmt.Fprintf(w, "|-----------|--------:|\n")

		for _, ns := range uncovered {
			fmt.Fprintf(w, "| %s | %d |\n", ns.namespace, len(ns.methods))
		}

		fmt.Fprintln(w)
	}

	// Write subscribe-only / unmapped Go methods (API methods not in methods.json)
	var unmapped []goMethod
	seen := make(map[string]bool)
	for _, gm := range goMethods {
		if _, inAPI := apiMethods[gm.APIMethod]; !inAPI {
			key := gm.ServiceStruct + "." + gm.GoMethodName + "." + gm.APIMethod
			if !seen[key] {
				seen[key] = true
				unmapped = append(unmapped, gm)
			}
		}
	}

	if len(unmapped) > 0 {
		sort.Slice(unmapped, func(i, j int) bool {
			if unmapped[i].ServiceStruct != unmapped[j].ServiceStruct {
				return unmapped[i].ServiceStruct < unmapped[j].ServiceStruct
			}
			return unmapped[i].APIMethod < unmapped[j].APIMethod
		})

		fmt.Fprintf(w, "## Go Methods Not in API Schema (%d methods)\n\n", len(unmapped))
		fmt.Fprintf(w, "These Go methods call API endpoints not present in the %s method schema\n", version)
		fmt.Fprintf(w, "(e.g., subscription/event channels, version-specific aliases).\n\n")
		fmt.Fprintf(w, "| Go Service | Go Method | API Method |\n")
		fmt.Fprintf(w, "|------------|-----------|------------|\n")

		for _, gm := range unmapped {
			fmt.Fprintf(w, "| %s | %s | %s |\n", gm.ServiceStruct, gm.GoMethodName, gm.APIMethod)
		}

		fmt.Fprintln(w)
	}
}
