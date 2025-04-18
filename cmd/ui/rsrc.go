//go:build windows
// +build windows

package main

import (
    _ "embed"
)

//go:embed assets/app.manifest
var manifest []byte