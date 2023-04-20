package wgo

import (
	"fmt"
	"time"

	"github.com/wxnacy/wgo/commands"
)

const (
	Version = "1.1.0"
)

var (
	versionFormat = `%s
Wgo version %s
Copyright (C) %d wxnacy
`
)

// 获取版本的完整信息
func GetWholeVersion() (string, error) {
	goVer, err := commands.Command("go", "version")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(versionFormat, goVer, Version, time.Now().Year()), nil
}

// 获取版本的简要信息
func GetVersion() string {
	return fmt.Sprintf("Wgo %s", Version)
}
