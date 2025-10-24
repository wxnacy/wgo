package handler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// 功能需求
// 将 BuiltinFuncCode 内容写入到 filename 中
func InitBuiltinFuncCode(filename string) error {
	if filename == "" {
		return errors.New("filename 不能为空")
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("解析路径失败: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(BuiltinFuncCode), 0o644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

var BuiltinFuncCode = `package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

var (
	RequestID string
	TempDir   string
)

var typeRegistry = map[string]reflect.Type{
	"bool":    reflect.TypeOf(false),
	"int":     reflect.TypeOf(int(0)),
	"int8":    reflect.TypeOf(int8(0)),
	"int16":   reflect.TypeOf(int16(0)),
	"int32":   reflect.TypeOf(int32(0)),
	"int64":   reflect.TypeOf(int64(0)),
	"uint":    reflect.TypeOf(uint(0)),
	"uint8":   reflect.TypeOf(uint8(0)),
	"uint16":  reflect.TypeOf(uint16(0)),
	"uint32":  reflect.TypeOf(uint32(0)),
	"uint64":  reflect.TypeOf(uint64(0)),
	"float32": reflect.TypeOf(float32(0)),
	"float64": reflect.TypeOf(float64(0)),
	"string":  reflect.TypeOf(""),
}

// 功能需求
// 现在有问题，调用 DeserializeFromFile 方法的时候是动态调用的，无法确定 T
// 所以能不能在 Serialize 方法中调用 SerializeToFile 后，将数据的类型 T 也存到 typePath 中
// 然后 Deserialize 方法通过 typePath 获取到 T 后，再调用 DeserializeFromFile 方法返回对象内容
// 实现完成后增加测试用例
func _Serialize[T any](name string, value T) error {
	dir := GetSerializeDir()
	filePath := filepath.Join(dir, name)
	typePath := filepath.Join(dir, name+".type")

	registerType(reflect.TypeOf(value))
	gob.Register(value)

	if err := SerializeToFile(filePath, value); err != nil {
		return err
	}

	t := reflect.TypeOf(value)
	if t == nil {
		return errors.New("无法获取对象类型")
	}

	if err := os.WriteFile(typePath, []byte(t.String()), 0o644); err != nil {
		return fmt.Errorf("写入类型信息失败: %w", err)
	}

	return nil
}

func Deserialize(name string) (any, error) {
	dir := GetSerializeDir()
	filePath := filepath.Join(dir, name)
	typePath := filepath.Join(dir, name+".type")

	data, err := os.ReadFile(typePath)
	if err != nil {
		return nil, fmt.Errorf("读取类型信息失败: %w", err)
	}

	typeName := string(data)
	t, err := resolveType(typeName)
	if err != nil {
		return nil, err
	}

	ptr := reflect.New(t)
	if err := decodeToValue(filePath, ptr.Interface()); err != nil {
		return nil, err
	}

	value := ptr.Elem().Interface()
	if value == nil {
		return nil, nil
	}

	actualTypeName := reflect.TypeOf(value).String()
	if actualTypeName != typeName {
		return nil, fmt.Errorf("类型不匹配，期望 %s, 实际 %s", typeName, actualTypeName)
	}

	return value, nil
}

func GetSerializeDir() string {
	if TempDir == "" {
		TempDir = ".wgo"
	}
	// return filepath.Join(TempDir, fmt.Sprintf("var-%d", time.Now().UnixMicro()))
	return TempDir
}

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

func registerType(t reflect.Type) {
	if t == nil {
		return
	}
	typeRegistry[t.String()] = t
	if t.Kind() == reflect.Ptr {
		registerType(t.Elem())
	}
}

func resolveType(name string) (reflect.Type, error) {
	if t, ok := typeRegistry[name]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("类型 %s 未注册", name)
}

func decodeToValue(filePath string, target any) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("decode gob: %w", err)
	}

	return nil
}
`
