package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

// SerializeToFile 使用 gob 将任意对象编码并写入指定文件路径。
func SerializeToFile[T any](filePath string, value T) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("ensure dir: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	if err := enc.Encode(value); err != nil {
		return fmt.Errorf("encode gob: %w", err)
	}

	return nil
}

// DeserializeFromFile 从文件中读取 gob 数据并反序列化为泛型对象。
func DeserializeFromFile[T any](filePath string) (T, error) {
	var result T

	file, err := os.Open(filePath)
	if err != nil {
		return result, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	if err := dec.Decode(&result); err != nil {
		return result, fmt.Errorf("decode gob: %w", err)
	}

	return result, nil
}
