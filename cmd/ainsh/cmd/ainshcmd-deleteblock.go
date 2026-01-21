// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc/wshclient"
)

var deleteBlockCmd = &cobra.Command{
	Use:     "deleteblock",
	Short:   "delete a block",
	RunE:    deleteBlockRun,
	PreRunE: preRunSetupRpcClient,
}

func init() {
	rootCmd.AddCommand(deleteBlockCmd)
}

func deleteBlockRun(cmd *cobra.Command, args []string) (rtnErr error) {
	defer func() {
		sendActivity("deleteblock", rtnErr == nil)
	}()
	fullORef, err := resolveBlockArg()
	if err != nil {
		return err
	}
	if fullORef.OType != "block" {
		return fmt.Errorf("object reference is not a block")
	}
	deleteBlockData := &ainshrpc.CommandDeleteBlockData{
		BlockId: fullORef.OID,
	}
	err = wshclient.DeleteBlockCommand(RpcClient, *deleteBlockData, &ainshrpc.RpcOpts{Timeout: 2000})
	if err != nil {
		return fmt.Errorf("delete block failed: %v", err)
	}
	WriteStdout("block deleted\n")
	return nil
}
