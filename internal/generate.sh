#!/usr/bin/env bash
set -e

mockgen -source generate.go -destination mocks.go -package internal
sed -i '/^import /a \"testing\"' mocks.go
sed -i '/^type MockTB struct/a testing.TB' mocks.go
go fmt mocks.go 1>/dev/null