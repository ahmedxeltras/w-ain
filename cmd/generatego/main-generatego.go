// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/wavetermdev/ainterm/pkg/ainconfig"
	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/gogen"
	"github.com/wavetermdev/ainterm/pkg/util/utilfn"
)

const WshClientFileName = "pkg/ainshrpc/wshclient/wshclient.go"
const WaveObjMetaConstsFileName = "pkg/ainobj/metaconsts.go"
const SettingsMetaConstsFileName = "pkg/ainconfig/metaconsts.go"

func GenerateWshClient() error {
	fmt.Fprintf(os.Stderr, "generating wshclient file to %s\n", WshClientFileName)
	var buf strings.Builder
	gogen.GenerateBoilerplate(&buf, "wshclient", []string{
		"github.com/wavetermdev/ainterm/pkg/telemetry/telemetrydata",
		"github.com/wavetermdev/ainterm/pkg/ainshutil",
		"github.com/wavetermdev/ainterm/pkg/ainshrpc",
		"github.com/wavetermdev/ainterm/pkg/ainconfig",
		"github.com/wavetermdev/ainterm/pkg/ainobj",
		"github.com/wavetermdev/ainterm/pkg/ainps",
		"github.com/wavetermdev/ainterm/pkg/vdom",
		"github.com/wavetermdev/ainterm/pkg/util/iochan/iochantypes",
		"github.com/wavetermdev/ainterm/pkg/aiusechat/uctypes",
	})
	wshDeclMap := ainshrpc.GenerateWshCommandDeclMap()
	for _, key := range utilfn.GetOrderedMapKeys(wshDeclMap) {
		methodDecl := wshDeclMap[key]
		if methodDecl.CommandType == ainshrpc.RpcType_ResponseStream {
			gogen.GenMethod_ResponseStream(&buf, methodDecl)
		} else if methodDecl.CommandType == ainshrpc.RpcType_Call {
			gogen.GenMethod_Call(&buf, methodDecl)
		} else {
			panic("unsupported command type " + methodDecl.CommandType)
		}
	}
	buf.WriteString("\n")
	written, err := utilfn.WriteFileIfDifferent(WshClientFileName, []byte(buf.String()))
	if !written {
		fmt.Fprintf(os.Stderr, "no changes to %s\n", WshClientFileName)
	}
	return err
}

func GenerateWaveObjMetaConsts() error {
	fmt.Fprintf(os.Stderr, "generating waveobj meta consts file to %s\n", WaveObjMetaConstsFileName)
	var buf strings.Builder
	gogen.GenerateBoilerplate(&buf, "ainobj", []string{})
	gogen.GenerateMetaMapConsts(&buf, "MetaKey_", reflect.TypeOf(ainobj.MetaTSType{}), false)
	buf.WriteString("\n")
	written, err := utilfn.WriteFileIfDifferent(WaveObjMetaConstsFileName, []byte(buf.String()))
	if !written {
		fmt.Fprintf(os.Stderr, "no changes to %s\n", WaveObjMetaConstsFileName)
	}
	return err
}

func GenerateSettingsMetaConsts() error {
	fmt.Fprintf(os.Stderr, "generating settings meta consts file to %s\n", SettingsMetaConstsFileName)
	var buf strings.Builder
	gogen.GenerateBoilerplate(&buf, "ainconfig", []string{})
	gogen.GenerateMetaMapConsts(&buf, "ConfigKey_", reflect.TypeOf(ainconfig.SettingsType{}), false)
	buf.WriteString("\n")
	written, err := utilfn.WriteFileIfDifferent(SettingsMetaConstsFileName, []byte(buf.String()))
	if !written {
		fmt.Fprintf(os.Stderr, "no changes to %s\n", SettingsMetaConstsFileName)
	}
	return err
}

func main() {
	err := GenerateWshClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating wshclient: %v\n", err)
		return
	}
	err = GenerateWaveObjMetaConsts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating waveobj meta consts: %v\n", err)
		return
	}
	err = GenerateSettingsMetaConsts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating settings meta consts: %v\n", err)
		return
	}
}
