// +build tools

//go:generate rm -rf .tools
//go:generate mkdir -p .tools
//go:generate go build -o .tools/ github.com/golang/mock/mockgen
//go:generate go build -o .tools/ golang.org/x/tools/cmd/goimports
package testcase

import (
	_ "github.com/golang/mock/gomock"
	_ "golang.org/x/tools/cmd/goimports"
)
