package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

type testSample struct {
	Name string
	Age  int
	Tags []string
}

func TestSerializeDeserialize(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nested", "sample.gob")
	input := testSample{Name: "Bob", Age: 32, Tags: []string{"go", "test"}}

	if err := SerializeToFile(filePath, input); err != nil {
		t.Fatalf("SerializeToFile error: %v", err)
	}

	output, err := DeserializeFromFile[testSample](filePath)
	if err != nil {
		t.Fatalf("DeserializeFromFile error: %v", err)
	}

	if !reflect.DeepEqual(input, output) {
		t.Fatalf("unexpected result\ninput: %#v\noutput: %#v", input, output)
	}
}

func TestDeserializeFromFileMissing(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "missing.gob")

	if _, err := DeserializeFromFile[testSample](tempFile); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
