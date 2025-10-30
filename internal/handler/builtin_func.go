package handler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/wxnacy/go-tools"
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

// 通过运行代码判断方法是否有输出
func HasFunctionReturnByRun(funcStr, mainFile string) bool {
	tools.DirExistsOrCreate(filepath.Dir(mainFile))
	code := fmt.Sprintf(PRINT_FUNC_HAS_OUT_TEMPLATE, funcStr)
	logger.Debugf("HasFunctionReturnByRun code %s\n", code)
	out, err := WriteAndRunCode(code, mainFile)
	logger.Debugf("HasFunctionReturnByRun run out: %s err: %v", out, err)
	if err != nil {
		return false
	}
	flag, err := strconv.ParseBool(out)
	if err != nil {
		return false
	}
	return flag
}

const PRINT_FUNC_HAS_OUT_TEMPLATE = `package main

import (
	"fmt"
	"reflect"
)

func HasOut(i any) bool {
	t := reflect.TypeOf(i)
	return t.NumOut() > 0
}

func main() {
	fmt.Println(HasOut(%s))
}
`

var BuiltinFuncCode = `package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"
)

var (
	RequestID string
	TempDir   string
)

type functionRecord struct {
	Name string
}

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

var (
	funcRegistry        = map[string]any{}
	funcPointerRegistry = map[uintptr]string{}
	funcRegistryMu      sync.RWMutex
)

func init() {
	gob.Register(functionRecord{})
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

	t := reflect.TypeOf(value)
	registerType(t)
	if t != nil && t.Kind() != reflect.Func {
		gob.Register(value)
	}

	if err := SerializeToFile(filePath, value); err != nil {
		return err
	}

	if t == nil {
		return errors.New("无法获取对象类型")
	}

	if err := os.WriteFile(typePath, []byte(t.String()), 0o644); err != nil {
		return fmt.Errorf("写入类型信息失败: %w", err)
	}

	return nil
}

func _Deserialize[T any](name string) (T, error) {
	dir := GetSerializeDir()
	filePath := filepath.Join(dir, name)
	return DeserializeFromFile[T](filePath)
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

	if t.Kind() == reflect.Func {
		return deserializeFuncAny(filePath, t)
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
	return TempDir
}

// SerializeToFile 使用 gob 将任意对象编码并写入指定文件路径。
// 功能需求:
// - 将基础类型和任何对象可以序列化保存到文件
// - 将方法也可以序列化保存到文件
//
// 增加测试用例
func SerializeToFile[T any](filePath string, value T) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("ensure dir: %w", err)
	}

	if isFuncValue(value) {
		return serializeFuncValue(filePath, any(value))
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
// 功能需求:
// - 将基础类型和任何对象可以反序列化到泛型对象
// - 将方法也可以反序列化到泛型对象
//
// 增加测试用例
func DeserializeFromFile[T any](filePath string) (T, error) {
	var result T

	typeOfT := reflect.TypeOf((*T)(nil)).Elem()
	if typeOfT.Kind() == reflect.Func {
		return deserializeFuncValue[T](filePath)
	}

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

func isFuncValue[T any](value T) bool {
	if !reflect.ValueOf(value).IsValid() {
		return false
	}
	return reflect.TypeOf(value).Kind() == reflect.Func
}

// serializeFuncValue 用于序列化函数对象，便于随时移除相关实现。
func serializeFuncValue(filePath string, fn any) error {
	name, err := ensureFuncName(fn)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	record := functionRecord{Name: name}
	if err := gob.NewEncoder(file).Encode(record); err != nil {
		return fmt.Errorf("encode function record: %w", err)
	}

	return nil
}

func deserializeFuncValue[T any](filePath string) (T, error) {
	var zero T
	record, err := readFunctionRecord(filePath)
	if err != nil {
		return zero, err
	}

	fn, err := lookupFunc(record.Name)
	if err != nil {
		return zero, err
	}

	converted, ok := fn.(T)
	if !ok {
		return zero, fmt.Errorf("函数类型不匹配: %s", reflect.TypeOf(fn))
	}

	return converted, nil
}

func deserializeFuncAny(filePath string, target reflect.Type) (any, error) {
	record, err := readFunctionRecord(filePath)
	if err != nil {
		return nil, err
	}

	fn, err := lookupFunc(record.Name)
	if err != nil {
		return nil, err
	}

	value := reflect.ValueOf(fn)
	if !value.Type().AssignableTo(target) {
		return nil, fmt.Errorf("函数类型不匹配: 期望 %s, 实际 %s", target, value.Type())
	}

	return fn, nil
}

func readFunctionRecord(filePath string) (functionRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return functionRecord{}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var record functionRecord
	if err := gob.NewDecoder(file).Decode(&record); err != nil {
		return functionRecord{}, fmt.Errorf("decode function record: %w", err)
	}
	return record, nil
}

func ensureFuncName(fn any) (string, error) {
	val := reflect.ValueOf(fn)
	if !val.IsValid() || val.Kind() != reflect.Func {
		return "", errors.New("目标不是函数")
	}

	ptr := val.Pointer()
	funcRegistryMu.RLock()
	if name, ok := funcPointerRegistry[ptr]; ok {
		funcRegistryMu.RUnlock()
		return name, nil
	}
	funcRegistryMu.RUnlock()

	name := fmt.Sprintf("func-%d", time.Now().UnixNano())
	funcRegistryMu.Lock()
	funcRegistry[name] = fn
	funcPointerRegistry[ptr] = name
	funcRegistryMu.Unlock()

	return name, nil
}

func lookupFunc(name string) (any, error) {
	funcRegistryMu.RLock()
	defer funcRegistryMu.RUnlock()

	fn, ok := funcRegistry[name]
	if !ok {
		return nil, fmt.Errorf("函数 %s 未注册", name)
	}
	return fn, nil
}
`
