// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/wavetermdev/ainterm/cmd/ainsh/cmd"
	"github.com/wavetermdev/ainterm/pkg/ainbase"
)

// set by main-server.go
var WaveVersion = "0.0.0"
var BuildTime = "0"

func main() {
	ainbase.WaveVersion = WaveVersion
	ainbase.BuildTime = BuildTime
	cmd.Execute()
}
