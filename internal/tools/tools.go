//go:build tools
// +build tools

package tools

import (
	_ "github.com/alecthomas/kingpin/v2"
	_ "github.com/golang-migrate/migrate/v4/cmd/migrate"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/petergtz/pegomock/v4"

	// _ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "golang.org/x/tools/cmd/stringer"
)
