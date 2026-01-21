// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package wshserver

import (
	"sync"

	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainshutil"
)

const (
	DefaultOutputChSize = 32
	DefaultInputChSize  = 32
)

var waveSrvClient_Singleton *ainshutil.WshRpc
var waveSrvClient_Once = &sync.Once{}

// returns the wavesrv main rpc client singleton
func GetMainRpcClient() *ainshutil.WshRpc {
	waveSrvClient_Once.Do(func() {
		waveSrvClient_Singleton = ainshutil.MakeWshRpc(ainshrpc.RpcContext{}, &WshServerImpl, "main-client")
	})
	return waveSrvClient_Singleton
}
