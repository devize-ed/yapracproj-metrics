package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"go.uber.org/zap"
)

const srcFile = `
	package test

	// generate:reset
	type ResetableStruct struct {
		i     int
		str   string
		strP  *string
		s     []int
		m     map[string]string
		child *ResetableStruct
	}

	// non-resetable struct
	type NonResetableStruct struct {
		nonResetable      int
		nonResetableP     *int
		nonResetableS     []int
		nonResetableM     map[string]string
		nonResetableChild *NonResetableStruct
	}
`

const expectedFile = `
package test

func (rs *ResetableStruct) Reset() {
    if rs == nil {
        return
    }

    rs.i = 0
    rs.str = ""
    if rs.strP != nil {
        *rs.strP = ""
    }
    rs.s = rs.s[:0]
    clear(rs.m)
    if resetter, ok := rs.child.(interface{ Reset() }); ok && rs.child != nil {
        resetter.Reset()
    }
} 
`

func Test_searchStructsInFile(t *testing.T) {
	// create a temporary directory
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "pkg", "x.go")
	if err := os.MkdirAll(filepath.Dir(srcPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(srcPath, []byte(srcFile), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// prepare template
	tpl, err := template.New("reset").Parse(templateStr)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	logger := zap.NewNop().Sugar()

	// search the structs in the file
	if err := searchStructsInFile(srcPath, tpl, logger); err != nil {
		t.Fatalf("searchStructsInFile: %v", err)
	}

	// check if the generated file exists next to the source file
	genPath := filepath.Join(filepath.Dir(srcPath), "ResetableStruct_reset.go")
	if _, err := os.Stat(genPath); err != nil {
		t.Fatalf("generated file missing: %v", err)
	}
	b, err := os.ReadFile(genPath)
	if err != nil {
		t.Fatalf("read generated: %v", err)
	}
	gen := string(b)

	exp := strings.ReplaceAll(expectedFile, "(rs *ResetableStruct)", "(r *ResetableStruct)")
	exp = strings.ReplaceAll(exp, "rs.", "r.")
	exp = strings.ReplaceAll(exp, "if rs == nil", "if r == nil")

	// check if the essential lines from the expected file are present in the generated output
	for _, want := range []string{
		"package test",
		"func (r *ResetableStruct) Reset()",
		"if r == nil {",
		"r.i = 0",
		"r.str = \"\"",
		"if r.strP != nil { *r.strP = \"\" }",
		"r.s = r.s[:0]",
		"clear(r.m)",
		"if resetter, ok := interface{}(r.child).(interface{ Reset() }); ok && r.child != nil { resetter.Reset() }",
	} {
		if !strings.Contains(gen, want) {
			t.Fatalf("missing expected snippet: %q\n +++++++ GENERATED ++++++ \n%s\n++++++ EXPECTED ++++++ \n%s", want, gen, exp)
		}
	}
}
