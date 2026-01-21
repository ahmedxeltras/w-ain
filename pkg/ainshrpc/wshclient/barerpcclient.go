// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package wshclient

import (
	"sync"

	"github.com/wavetermdev/ainterm/pkg/ainps"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainshutil"
)

type WshServer struct{}

func (*WshServer) WshServerImpl() {}

var WshServerImpl = WshServer{}

const (
	DefaultOutputChSize = 32
	DefaultInputChSize  = 32
)

var ainSrvClient_Singleton *ainshutil.WshRpc
var ainSrvClient_Once = &sync.Once{}

const BareClientRoute = "bare"

func GetBareRpcClient() *ainshutil.WshRpc {
	ainSrvClient_Once.Do(func() {
		ainSrvClient_Singleton = ainshutil.MakeWshRpc(ainshrpc.RpcContext{}, &WshServerImpl, "bare-client")
		ainshutil.DefaultRouter.RegisterTrustedLeaf(ainSrvClient_Singleton, BareClientRoute)
		ainps.Broker.SetClient(ainshutil.DefaultRouter)
	})
	return ainSrvClient_Singleton
}
