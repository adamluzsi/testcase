// +build tools

//go:generate rm -rf .tools
//go:generate mkdir -p .tools
package testcase

//go:generate go build -o .tools/ github.com/golang/mock/mockgen
import _ "github.com/golang/mock/gomock"
