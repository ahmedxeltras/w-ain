// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package ainshutil

import (
	"context"
	"fmt"
	"log"

	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/baseds"
	"github.com/wavetermdev/ainterm/pkg/util/shellutil"
	"github.com/wavetermdev/ainterm/pkg/util/utilfn"
)

type WshRouterControlImpl struct {
	Router *WshRouter
}

func (impl *WshRouterControlImpl) WshServerImpl() {}

func (impl *WshRouterControlImpl) RouteAnnounceCommand(ctx context.Context) error {
	source := GetRpcSourceFromContext(ctx)
	if source == "" {
		return fmt.Errorf("no source in routeannounce")
	}
	handler := GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return fmt.Errorf("no response handler in context")
	}
	linkId := handler.GetIngressLinkId()
	if linkId == baseds.NoLinkId {
		return fmt.Errorf("no ingress link found")
	}
	return impl.Router.bindRoute(linkId, source, false)
}

func (impl *WshRouterControlImpl) RouteUnannounceCommand(ctx context.Context) error {
	source := GetRpcSourceFromContext(ctx)
	if source == "" {
		return fmt.Errorf("no source in routeunannounce")
	}
	handler := GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return fmt.Errorf("no response handler in context")
	}
	linkId := handler.GetIngressLinkId()
	if linkId == baseds.NoLinkId {
		return fmt.Errorf("no ingress link found")
	}
	return impl.Router.unbindRoute(linkId, source)
}

func (impl *WshRouterControlImpl) SetPeerInfoCommand(ctx context.Context, peerInfo string) error {
	source := GetRpcSourceFromContext(ctx)
	linkId := impl.Router.GetLinkIdForRoute(source)
	if linkId == baseds.NoLinkId {
		return fmt.Errorf("no link found for source route %q", source)
	}
	lm := impl.Router.getLinkMeta(linkId)
	if lm == nil {
		return fmt.Errorf("no link meta found for linkId %d", linkId)
	}
	if proxy, ok := lm.client.(*WshRpcProxy); ok {
		proxy.SetPeerInfo(peerInfo)
		return nil
	}
	return fmt.Errorf("setpeerinfo only valid for proxy connections")
}

func (impl *WshRouterControlImpl) AuthenticateCommand(ctx context.Context, data string) (ainshrpc.CommandAuthenticateRtnData, error) {
	handler := GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no response handler in context")
	}
	linkId := handler.GetIngressLinkId()
	if linkId == baseds.NoLinkId {
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no ingress link found")
	}

	newCtx, err := ValidateAndExtractRpcContextFromToken(data)
	if err != nil {
		log.Printf("wshrouter authenticate error linkid=%d: %v", linkId, err)
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("error validating token: %w", err)
	}
	routeId, err := validateRpcContextFromAuth(newCtx)
	if err != nil {
		return ainshrpc.CommandAuthenticateRtnData{}, err
	}

	rtnData := ainshrpc.CommandAuthenticateRtnData{}
	if newCtx.IsRouter {
		log.Printf("wshrouter authenticate success linkid=%d (router)", linkId)
		impl.Router.trustLink(linkId, LinkKind_Router)
	} else {
		log.Printf("wshrouter authenticate success linkid=%d routeid=%q", linkId, routeId)
		impl.Router.trustLink(linkId, LinkKind_Leaf)
		impl.Router.bindRoute(linkId, routeId, true)
	}

	return rtnData, nil
}

func (impl *WshRouterControlImpl) AuthenticateTokenCommand(ctx context.Context, data ainshrpc.CommandAuthenticateTokenData) (ainshrpc.CommandAuthenticateRtnData, error) {
	handler := GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no response handler in context")
	}
	linkId := handler.GetIngressLinkId()
	if linkId == baseds.NoLinkId {
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no ingress link found")
	}

	if data.Token == "" {
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no token in authenticatetoken message")
	}

	var rtnData ainshrpc.CommandAuthenticateRtnData
	var rpcContext *ainshrpc.RpcContext
	if impl.Router.IsRootRouter() {
		entry := shellutil.GetAndRemoveTokenSwapEntry(data.Token)
		if entry == nil {
			log.Printf("wshrouter authenticate-token error linkid=%d: no token entry found", linkId)
			return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no token entry found")
		}
		_, err := validateRpcContextFromAuth(entry.RpcContext)
		if err != nil {
			return ainshrpc.CommandAuthenticateRtnData{}, err
		}
		if entry.RpcContext.IsRouter {
			return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("cannot auth router via token")
		}
		if entry.RpcContext.RouteId == "" {
			return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no routeid")
		}
		rpcContext = entry.RpcContext
		rtnData = ainshrpc.CommandAuthenticateRtnData{
			Env:            entry.Env,
			InitScriptText: entry.ScriptText,
			RpcContext:     rpcContext,
		}
	} else {
		wshRpc := GetWshRpcFromContext(ctx)
		if wshRpc == nil {
			return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no wshrpc in context")
		}
		respData, err := wshRpc.SendRpcRequest(ainshrpc.Command_AuthenticateTokenVerify, data, &ainshrpc.RpcOpts{Route: ControlRootRoute})
		if err != nil {
			log.Printf("wshrouter authenticate-token error linkid=%d: failed to verify token: %v", linkId, err)
			return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("failed to verify token: %w", err)
		}
		err = utilfn.ReUnmarshal(&rtnData, respData)
		if err != nil {
			return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		rpcContext = rtnData.RpcContext
	}

	if rpcContext == nil {
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no rpccontext in token response")
	}
	log.Printf("wshrouter authenticate-token success linkid=%d routeid=%q", linkId, rpcContext.RouteId)
	impl.Router.trustLink(linkId, LinkKind_Leaf)
	impl.Router.bindRoute(linkId, rpcContext.RouteId, true)

	return rtnData, nil
}

func (impl *WshRouterControlImpl) AuthenticateTokenVerifyCommand(ctx context.Context, data ainshrpc.CommandAuthenticateTokenData) (ainshrpc.CommandAuthenticateRtnData, error) {
	if !impl.Router.IsRootRouter() {
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("authenticatetokenverify can only be called on root router")
	}

	if data.Token == "" {
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no token in authenticatetoken message")
	}
	entry := shellutil.GetAndRemoveTokenSwapEntry(data.Token)
	if entry == nil {
		log.Printf("wshrouter authenticate-token-verify error: no token entry found")
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no token entry found")
	}
	_, err := validateRpcContextFromAuth(entry.RpcContext)
	if err != nil {
		return ainshrpc.CommandAuthenticateRtnData{}, err
	}
	if entry.RpcContext.IsRouter {
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("cannot auth router via token")
	}
	if entry.RpcContext.RouteId == "" {
		return ainshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no routeid")
	}

	rtnData := ainshrpc.CommandAuthenticateRtnData{
		Env:            entry.Env,
		InitScriptText: entry.ScriptText,
		RpcContext:     entry.RpcContext,
	}

	log.Printf("wshrouter authenticate-token-verify success routeid=%q", entry.RpcContext.RouteId)
	return rtnData, nil
}

func validateRpcContextFromAuth(newCtx *ainshrpc.RpcContext) (string, error) {
	if newCtx == nil {
		return "", fmt.Errorf("no context found in jwt token")
	}
	if newCtx.IsRouter && newCtx.RouteId != "" {
		return "", fmt.Errorf("invalid context, router cannot have a routeid")
	}
	if !newCtx.IsRouter && newCtx.RouteId == "" {
		return "", fmt.Errorf("invalid context, must have a routeid")
	}
	if newCtx.IsRouter {
		return "", nil
	}
	return newCtx.RouteId, nil
}
