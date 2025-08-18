package util

import "path/filepath"

// service names
var (
	APIServer = "apiserver"
)

// base paths
var (
	baseConfigPath = "/config"
	baseCodePath   = "internal/pkg/code"
)

// specific paths
var (
	CodePath = map[string]string{
		"apiserver": filepath.Join(baseCodePath, "apiserver.go"),
	}
)
