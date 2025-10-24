package handler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitBuiltinFuncCodeWritesFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "sub", "builtin.go")

	if err := InitBuiltinFuncCode(filePath); err != nil {
		t.Fatalf("InitBuiltinFuncCode error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}

	if string(data) != BuiltinFuncCode {
		t.Fatalf("unexpected file content\nwant: %q\n got: %q", BuiltinFuncCode, string(data))
	}
}

func TestInitBuiltinFuncCodeEmptyFilename(t *testing.T) {
	if err := InitBuiltinFuncCode(""); err == nil {
		t.Fatal("expected error for empty filename, got nil")
	}
}
