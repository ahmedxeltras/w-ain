// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package aiusechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainstore"
	"github.com/wavetermdev/ainterm/pkg/aiusechat/uctypes"
	"github.com/wavetermdev/ainterm/pkg/blockcontroller"
	"github.com/wavetermdev/ainterm/pkg/util/utilfn"
)

func getTsunamiShortDesc(rtInfo *ainobj.ObjRTInfo) string {
	if rtInfo == nil || rtInfo.TsunamiAppMeta == nil {
		return ""
	}
	var appMeta ainshrpc.AppMeta
	if err := utilfn.ReUnmarshal(&appMeta, rtInfo.TsunamiAppMeta); err == nil && appMeta.ShortDesc != "" {
		return appMeta.ShortDesc
	}
	return ""
}

func handleTsunamiBlockDesc(block *ainobj.Block) string {
	status := blockcontroller.GetBlockControllerRuntimeStatus(block.OID)
	if status == nil || status.ShellProcStatus != blockcontroller.Status_Running {
		return "tsunami framework widget that is currently not running"
	}

	blockORef := ainobj.MakeORef(ainobj.OType_Block, block.OID)
	rtInfo := ainstore.GetRTInfo(blockORef)
	if shortDesc := getTsunamiShortDesc(rtInfo); shortDesc != "" {
		return fmt.Sprintf("tsunami widget - %s", shortDesc)
	}
	return "tsunami widget - unknown description"
}

func makeTsunamiGetCallback(status *blockcontroller.BlockControllerRuntimeStatus, apiPath string) func(any, *uctypes.UIMessageDataToolUse) (any, error) {
	return func(input any, toolUseData *uctypes.UIMessageDataToolUse) (any, error) {
		if status.TsunamiPort == 0 {
			return nil, fmt.Errorf("tsunami port not available")
		}

		url := fmt.Sprintf("http://localhost:%d%s", status.TsunamiPort, apiPath)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request to tsunami: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("tsunami returned status %d", resp.StatusCode)
		}

		var result any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode tsunami response: %w", err)
		}

		return result, nil
	}
}

func makeTsunamiPostCallback(status *blockcontroller.BlockControllerRuntimeStatus, apiPath string) func(any, *uctypes.UIMessageDataToolUse) (any, error) {
	return func(input any, toolUseData *uctypes.UIMessageDataToolUse) (any, error) {
		if status.TsunamiPort == 0 {
			return nil, fmt.Errorf("tsunami port not available")
		}

		url := fmt.Sprintf("http://localhost:%d%s", status.TsunamiPort, apiPath)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var reqBody []byte
		var err error
		if input != nil {
			reqBody, err = json.Marshal(input)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal input: %w", err)
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(reqBody)))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request to tsunami: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("tsunami returned status %d", resp.StatusCode)
		}

		return true, nil
	}
}

func GetTsunamiGetDataToolDefinition(block *ainobj.Block, rtInfo *ainobj.ObjRTInfo, status *blockcontroller.BlockControllerRuntimeStatus) *uctypes.ToolDefinition {
	blockIdPrefix := block.OID[:8]
	toolName := fmt.Sprintf("tsunami_getdata_%s", blockIdPrefix)

	desc := "tsunami widget"
	if shortDesc := getTsunamiShortDesc(rtInfo); shortDesc != "" {
		desc = shortDesc
	}

	return &uctypes.ToolDefinition{
		Name:        toolName,
		ToolLogName: "tsunami:getdata",
		Strict:      true,
		InputSchema: map[string]any{
			"type":                 "object",
			"properties":           map[string]any{},
			"additionalProperties": false,
		},
		ToolCallDesc: func(input any, output any, toolUseData *uctypes.UIMessageDataToolUse) string {
			return fmt.Sprintf("getting data from %s (%s)", desc, blockIdPrefix)
		},
		ToolAnyCallback: makeTsunamiGetCallback(status, "/api/data"),
	}
}

func GetTsunamiGetConfigToolDefinition(block *ainobj.Block, rtInfo *ainobj.ObjRTInfo, status *blockcontroller.BlockControllerRuntimeStatus) *uctypes.ToolDefinition {
	blockIdPrefix := block.OID[:8]
	toolName := fmt.Sprintf("tsunami_getconfig_%s", blockIdPrefix)

	desc := "tsunami widget"
	if shortDesc := getTsunamiShortDesc(rtInfo); shortDesc != "" {
		desc = shortDesc
	}

	return &uctypes.ToolDefinition{
		Name:        toolName,
		ToolLogName: "tsunami:getconfig",
		Strict:      true,
		InputSchema: map[string]any{
			"type":                 "object",
			"properties":           map[string]any{},
			"additionalProperties": false,
		},
		ToolCallDesc: func(input any, output any, toolUseData *uctypes.UIMessageDataToolUse) string {
			return fmt.Sprintf("getting config from %s (%s)", desc, blockIdPrefix)
		},
		ToolAnyCallback: makeTsunamiGetCallback(status, "/api/config"),
	}
}

func GetTsunamiSetConfigToolDefinition(block *ainobj.Block, rtInfo *ainobj.ObjRTInfo, status *blockcontroller.BlockControllerRuntimeStatus) *uctypes.ToolDefinition {
	blockIdPrefix := block.OID[:8]
	toolName := fmt.Sprintf("tsunami_setconfig_%s", blockIdPrefix)

	var inputSchema map[string]any
	if rtInfo != nil && rtInfo.TsunamiSchemas != nil {
		if schemasMap, ok := rtInfo.TsunamiSchemas.(map[string]any); ok {
			if configSchema, exists := schemasMap["config"]; exists {
				inputSchema = configSchema.(map[string]any)
			}
		}
	}

	if inputSchema == nil {
		return nil
	}

	desc := "tsunami widget"
	if shortDesc := getTsunamiShortDesc(rtInfo); shortDesc != "" {
		desc = shortDesc
	}

	return &uctypes.ToolDefinition{
		Name:        toolName,
		ToolLogName: "tsunami:setconfig",
		InputSchema: inputSchema,
		ToolCallDesc: func(input any, output any, toolUseData *uctypes.UIMessageDataToolUse) string {
			return fmt.Sprintf("updating config for %s (%s)", desc, blockIdPrefix)
		},
		ToolAnyCallback: makeTsunamiPostCallback(status, "/api/config"),
	}
}
