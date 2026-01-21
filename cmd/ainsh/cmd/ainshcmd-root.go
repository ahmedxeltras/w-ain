// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc/wshclient"
	"github.com/wavetermdev/ainterm/pkg/ainshutil"
	"github.com/wavetermdev/ainterm/pkg/util/shellutil"
)

var (
	rootCmd = &cobra.Command{
		Use:          "wsh",
		Short:        "CLI tool to control Wave Terminal",
		Long:         `wsh is a small utility that lets you do cool things with Wave Terminal, right from the command line`,
		SilenceUsage: true,
	}
)

var WrappedStdin io.Reader = os.Stdin
var WrappedStdout io.Writer = &WrappedWriter{dest: os.Stdout}
var WrappedStderr io.Writer = &WrappedWriter{dest: os.Stderr}
var RpcClient *ainshutil.WshRpc
var RpcContext ainshrpc.RpcContext
var UsingTermWshMode bool
var blockArg string
var WshExitCode int

type WrappedWriter struct {
	dest io.Writer
}

func (w *WrappedWriter) Write(p []byte) (n int, err error) {
	if !UsingTermWshMode {
		return w.dest.Write(p)
	}
	count := 0
	for _, b := range p {
		if b == '\n' {
			count++
		}
	}
	if count == 0 {
		return w.dest.Write(p)
	}
	buf := make([]byte, len(p)+count) // Each '\n' adds one extra byte for '\r'
	writeIdx := 0
	for _, b := range p {
		if b == '\n' {
			buf[writeIdx] = '\r'
			buf[writeIdx+1] = '\n'
			writeIdx += 2
		} else {
			buf[writeIdx] = b
			writeIdx++
		}
	}
	return w.dest.Write(buf)
}

func WriteStderr(fmtStr string, args ...interface{}) {
	WrappedStderr.Write([]byte(fmt.Sprintf(fmtStr, args...)))
}

func WriteStdout(fmtStr string, args ...interface{}) {
	WrappedStdout.Write([]byte(fmt.Sprintf(fmtStr, args...)))
}

func OutputHelpMessage(cmd *cobra.Command) {
	cmd.SetOutput(WrappedStderr)
	cmd.Help()
	WriteStderr("\n")
}

func preRunSetupRpcClient(cmd *cobra.Command, args []string) error {
	jwtToken := os.Getenv(ainshutil.WaveJwtTokenVarName)
	if jwtToken == "" {
		return fmt.Errorf("wsh must be run inside a Wave-managed SSH session (AINTERM_JWT not found)")
	}
	err := setupRpcClient(nil, jwtToken)
	if err != nil {
		return err
	}
	return nil
}

func getIsTty() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}
	return false
}

type RunEFnType = func(*cobra.Command, []string) error

func activityWrap(activityStr string, origRunE RunEFnType) RunEFnType {
	return func(cmd *cobra.Command, args []string) (rtnErr error) {
		defer func() {
			sendActivity(activityStr, rtnErr == nil)
		}()
		return origRunE(cmd, args)
	}
}

func resolveBlockArg() (*ainobj.ORef, error) {
	oref := blockArg
	if oref == "" {
		oref = "this"
	}
	fullORef, err := resolveSimpleId(oref)
	if err != nil {
		return nil, fmt.Errorf("resolving blockid: %w", err)
	}
	return fullORef, nil
}

func setupRpcClientWithToken(swapTokenStr string) (ainshrpc.CommandAuthenticateRtnData, error) {
	var rtn ainshrpc.CommandAuthenticateRtnData
	token, err := shellutil.UnpackSwapToken(swapTokenStr)
	if err != nil {
		return rtn, fmt.Errorf("error unpacking token: %w", err)
	}
	if token.RpcContext == nil {
		return rtn, fmt.Errorf("no rpccontext in token")
	}
	if token.RpcContext.SockName == "" {
		return rtn, fmt.Errorf("no sockname in token")
	}
	RpcContext = *token.RpcContext
	RpcClient, err = ainshutil.SetupDomainSocketRpcClient(token.RpcContext.SockName, nil, "wshcmd")
	if err != nil {
		return rtn, fmt.Errorf("error setting up domain socket rpc client: %w", err)
	}
	return wshclient.AuthenticateTokenCommand(RpcClient, ainshrpc.CommandAuthenticateTokenData{Token: token.Token}, &ainshrpc.RpcOpts{Route: ainshutil.ControlRoute})
}

// returns the wrapped stdin and a new rpc client (that wraps the stdin input and stdout output)
func setupRpcClient(serverImpl ainshutil.ServerImpl, jwtToken string) error {
	rpcCtx, err := ainshutil.ExtractUnverifiedRpcContext(jwtToken)
	if err != nil {
		return fmt.Errorf("error extracting rpc context from %s: %v", ainshutil.WaveJwtTokenVarName, err)
	}
	RpcContext = *rpcCtx
	sockName, err := ainshutil.ExtractUnverifiedSocketName(jwtToken)
	if err != nil {
		return fmt.Errorf("error extracting socket name from %s: %v", ainshutil.WaveJwtTokenVarName, err)
	}
	RpcClient, err = ainshutil.SetupDomainSocketRpcClient(sockName, serverImpl, "wshcmd")
	if err != nil {
		return fmt.Errorf("error setting up domain socket rpc client: %v", err)
	}
	_, err = wshclient.AuthenticateCommand(RpcClient, jwtToken, &ainshrpc.RpcOpts{Route: ainshutil.ControlRoute})
	if err != nil {
		return fmt.Errorf("error authenticating: %v", err)
	}
	blockId := os.Getenv("AINTERM_BLOCKID")
	if blockId != "" {
		peerInfo := fmt.Sprintf("domain:block:%s", blockId)
		wshclient.SetPeerInfoCommand(RpcClient, peerInfo, &ainshrpc.RpcOpts{Route: ainshutil.ControlRoute})
	}
	// note we don't modify WrappedStdin here (just use os.Stdin)
	return nil
}

func isFullORef(orefStr string) bool {
	_, err := ainobj.ParseORef(orefStr)
	return err == nil
}

func resolveSimpleId(id string) (*ainobj.ORef, error) {
	if isFullORef(id) {
		orefObj, err := ainobj.ParseORef(id)
		if err != nil {
			return nil, fmt.Errorf("error parsing full ORef: %v", err)
		}
		return &orefObj, nil
	}
	blockId := os.Getenv("AINTERM_BLOCKID")
	if blockId == "" {
		return nil, fmt.Errorf("no AINTERM_BLOCKID env var set")
	}
	rtnData, err := wshclient.ResolveIdsCommand(RpcClient, ainshrpc.CommandResolveIdsData{
		BlockId: blockId,
		Ids:     []string{id},
	}, &ainshrpc.RpcOpts{Timeout: 2000})
	if err != nil {
		return nil, fmt.Errorf("error resolving ids: %v", err)
	}
	oref, ok := rtnData.ResolvedIds[id]
	if !ok {
		return nil, fmt.Errorf("id not found: %q", id)
	}
	return &oref, nil
}

func getTabIdFromEnv() string {
	return os.Getenv("AINTERM_TABID")
}

// this will send wsh activity to the client running on *your* local machine (it does not contact any wave cloud infrastructure)
// if you've turned off telemetry in your local client, this data never gets sent to us
// no parameters or timestamps are sent, as you can see below, it just sends the name of the command (and if there was an error)
// (e.g. "wsh ai ..." would send "ai")
// this helps us understand which commands are actually being used so we know where to concentrate our effort
func sendActivity(wshCmdName string, success bool) {
	if RpcClient == nil || wshCmdName == "" {
		return
	}
	dataMap := make(map[string]int)
	dataMap[wshCmdName] = 1
	if !success {
		dataMap[wshCmdName+"#"+"error"] = 1
	}
	wshclient.WshActivityCommand(RpcClient, dataMap, nil)
}

// Execute executes the root command.
func Execute() {
	defer func() {
		r := recover()
		if r != nil {
			WriteStderr("[panic] %v\n", r)
			debug.PrintStack()
			ainshutil.DoShutdown("", 1, true)
		} else {
			ainshutil.DoShutdown("", WshExitCode, false)
		}
	}()
	rootCmd.PersistentFlags().StringVarP(&blockArg, "block", "b", "", "for commands which require a block id")
	err := rootCmd.Execute()
	if err != nil {
		ainshutil.DoShutdown("", 1, true)
		return
	}
}
