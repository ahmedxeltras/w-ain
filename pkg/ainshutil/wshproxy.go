// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package ainshutil

import (
	"fmt"
	"sync"

	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/baseds"
	"github.com/wavetermdev/ainterm/pkg/panichandler"
)

type WshRpcProxy struct {
	Lock         *sync.Mutex
	RpcContext   *ainshrpc.RpcContext
	ToRemoteCh   chan []byte
	FromRemoteCh chan baseds.RpcInputChType
	PeerInfo     string
}

func MakeRpcProxy(peerInfo string) *WshRpcProxy {
	return &WshRpcProxy{
		Lock:         &sync.Mutex{},
		ToRemoteCh:   make(chan []byte, DefaultInputChSize),
		FromRemoteCh: make(chan baseds.RpcInputChType, DefaultOutputChSize),
		PeerInfo:     peerInfo,
	}
}

func (p *WshRpcProxy) GetPeerInfo() string {
	return p.PeerInfo
}

func (p *WshRpcProxy) SetPeerInfo(peerInfo string) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.PeerInfo = peerInfo
}

// TODO: Figure out who is sending to closed routes and why we're not catching it
func (p *WshRpcProxy) SendRpcMessage(msg []byte, ingressLinkId baseds.LinkId, debugStr string) {
	defer func() {
		panicCtx := "WshRpcProxy.SendRpcMessage"
		if debugStr != "" {
			panicCtx = fmt.Sprintf("%s:%s", panicCtx, debugStr)
		}
		panichandler.PanicHandler(panicCtx, recover())
	}()
	p.ToRemoteCh <- msg
}

func (p *WshRpcProxy) RecvRpcMessage() ([]byte, bool) {
	inputVal, more := <-p.FromRemoteCh
	return inputVal.MsgBytes, more
}
