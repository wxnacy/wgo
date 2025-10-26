package main

import (
	"os"
	"path/filepath"
	"testing"
)

type sampleStruct struct {
	ID   int
	Name string
}

func TestSerializeAndDeserialize(t *testing.T) {
	TempDir = t.TempDir()
	name := "sample.gob"
	original := sampleStruct{ID: 42, Name: "answer"}
	defer func() { TempDir = "" }()

	if err := _Serialize(name, original); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	value, err := Deserialize(name)
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}

	got, ok := value.(sampleStruct)
	if !ok {
		t.Fatalf("unexpected type: %T", value)
	}

	if got != original {
		t.Fatalf("mismatch\nwant: %#v\n got: %#v", original, got)
	}

	typePath := filepath.Join(GetSerializeDir(), name+".type")
	if err := os.WriteFile(typePath, []byte("int"), 0o644); err != nil {
		t.Fatalf("override type file error: %v", err)
	}

	if _, err := Deserialize(name); err == nil {
		t.Fatal("expected type mismatch error, got nil")
	}
}

func TestSerializeTypeFileWritten(t *testing.T) {
	TempDir = t.TempDir()
	name := "value.gob"
	defer func() { TempDir = "" }()

	if err := _Serialize(name, 123); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	val, err := Deserialize(name)
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}

	if v, ok := val.(int); !ok || v != 123 {
		t.Fatalf("unexpected value: %#v", val)
	}

	typePath := filepath.Join(GetSerializeDir(), name+".type")
	if _, err := os.Stat(typePath); err != nil {
		t.Fatalf("type file not created: %v", err)
	}
}

func TestSerializeAndDeserializeFunctionValue(t *testing.T) {
	resetFuncRegistryForTest()
	TempDir = t.TempDir()
	name := "func_value.gob"
	defer func() {
		TempDir = ""
		resetFuncRegistryForTest()
	}()

	original := func() string { return "wxnacy" }
	if err := _Serialize(name, original); err != nil {
		t.Fatalf("Serialize func error: %v", err)
	}

	value, err := Deserialize(name)
	if err != nil {
		t.Fatalf("Deserialize func error: %v", err)
	}

	restored, ok := value.(func() string)
	if !ok {
		t.Fatalf("unexpected type: %T", value)
	}

	if restored() != "wxnacy" {
		t.Fatalf("unexpected restored result: %s", restored())
	}
}
