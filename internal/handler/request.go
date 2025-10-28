package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	request     *Request
	onceRequest sync.Once
)

func GetRequest() *Request {
	if request == nil {
		onceRequest.Do(func() {
			requestId := fmt.Sprintf("WGO%d", time.Now().UnixMicro())
			envID := os.Getenv("WGO_TEST_REQUEST_ID")
			if envID != "" {
				requestId = envID
			}
			request = &Request{
				ID: requestId,
			}
			workspace, _ := os.Getwd()
			request.Workspace = workspace
			request.MainDir = filepath.Join(
				GetWorkspace(),
				".wgo",
				request.ID,
			)
			request.MainFile = filepath.Join(
				GetMainDir(),
				"main.go",
			)
			request.TempDir = filepath.Join(
				os.TempDir(),
				"wgo",
				request.ID,
			)
		})
	}
	return request
}

type Request struct {
	ID        string
	Workspace string
	MainDir   string
	MainFile  string
	TempDir   string
}

func (r Request) ToCode() string {
	tpl := `package main

func init() {
	RequestID = "%s"
	TempDir = "%s"
}
`
	return fmt.Sprintf(
		tpl,
		r.ID,
		r.TempDir,
	)
}

func GetWorkspace() string {
	return GetRequest().Workspace
}

func GetMainDir() string {
	return GetRequest().MainDir
}

func GetMainFile() string {
	return GetRequest().MainFile
}

func GetTempDir() string {
	return GetRequest().TempDir
}
