// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wavetermdev/ainterm/pkg/ainconfig"
	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc/wshclient"
)

var (
	identityFiles []string
	newBlock      bool
)

var sshCmd = &cobra.Command{
	Use:     "ssh",
	Short:   "connect this terminal to a remote host",
	Args:    cobra.ExactArgs(1),
	RunE:    sshRun,
	PreRunE: preRunSetupRpcClient,
}

func init() {
	sshCmd.Flags().StringArrayVarP(&identityFiles, "identityfile", "i", []string{}, "add an identity file for publickey authentication")
	sshCmd.Flags().BoolVarP(&newBlock, "new", "n", false, "create a new terminal block with this connection")
	rootCmd.AddCommand(sshCmd)
}

func sshRun(cmd *cobra.Command, args []string) (rtnErr error) {
	defer func() {
		sendActivity("ssh", rtnErr == nil)
	}()

	sshArg := args[0]
	blockId := RpcContext.BlockId
	if blockId == "" && !newBlock {
		return fmt.Errorf("cannot determine blockid (not in JWT)")
	}

	// Create connection request
	connOpts := ainshrpc.ConnRequest{
		Host:       sshArg,
		LogBlockId: blockId,
		Keywords: ainconfig.ConnKeywords{
			SshIdentityFile: identityFiles,
		},
	}
	wshclient.ConnConnectCommand(RpcClient, connOpts, &ainshrpc.RpcOpts{Timeout: 60000})

	if newBlock {
		tabId := getTabIdFromEnv()
		if tabId == "" {
			return fmt.Errorf("no AINTERM_TABID env var set")
		}

		// Create a new block with the SSH connection
		createMeta := map[string]any{
			ainobj.MetaKey_View:       "term",
			ainobj.MetaKey_Controller: "shell",
			ainobj.MetaKey_Connection: sshArg,
		}
		if RpcContext.Conn != "" {
			createMeta[ainobj.MetaKey_Connection] = RpcContext.Conn
		}
		createBlockData := ainshrpc.CommandCreateBlockData{
			TabId: tabId,
			BlockDef: &ainobj.BlockDef{
				Meta: createMeta,
			},
			Focused: true,
		}
		oref, err := wshclient.CreateBlockCommand(RpcClient, createBlockData, nil)
		if err != nil {
			return fmt.Errorf("creating new terminal block: %w", err)
		}
		WriteStdout("new terminal block created with connection to %q: %s\n", sshArg, oref)
		return nil
	}

	// Update existing block with the new connection
	data := ainshrpc.CommandSetMetaData{
		ORef: ainobj.MakeORef(ainobj.OType_Block, blockId),
		Meta: map[string]any{
			ainobj.MetaKey_Connection: sshArg,
			ainobj.MetaKey_CmdCwd:     nil,
		},
	}
	err := wshclient.SetMetaCommand(RpcClient, data, nil)
	if err != nil {
		return fmt.Errorf("setting connection in block: %w", err)
	}
	WriteStderr("switched connection to %q\n", sshArg)
	return nil
}
