package util

import "path/filepath"

// service names
var (
	APIServer = "apiserver"
)

// base paths
const (
	BaseConfigPath = "./config"
	BaseCodePath   = "internal/pkg/code"
)

// specific paths
var (
	CodePath = map[string]string{
		"apiserver": filepath.Join(BaseCodePath, "apiserver.go"),
	}
)
