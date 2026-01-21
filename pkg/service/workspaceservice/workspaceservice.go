// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wavetermdev/ainterm/pkg/aincore"
	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainps"
	"github.com/wavetermdev/ainterm/pkg/ainstore"
	"github.com/wavetermdev/ainterm/pkg/blockcontroller"
	"github.com/wavetermdev/ainterm/pkg/panichandler"
	"github.com/wavetermdev/ainterm/pkg/tsgen/tsgenmeta"
)

const DefaultTimeout = 2 * time.Second

type WorkspaceService struct{}

func (svc *WorkspaceService) CreateWorkspace_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames:   []string{"ctx", "name", "icon", "color", "applyDefaults"},
		ReturnDesc: "workspaceId",
	}
}

func (svc *WorkspaceService) CreateWorkspace(ctx context.Context, name string, icon string, color string, applyDefaults bool) (string, error) {
	newWS, err := aincore.CreateWorkspace(ctx, name, icon, color, applyDefaults, false)
	if err != nil {
		return "", fmt.Errorf("error creating workspace: %w", err)
	}
	return newWS.OID, nil
}

func (svc *WorkspaceService) UpdateWorkspace_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"ctx", "workspaceId", "name", "icon", "color", "applyDefaults"},
	}
}

func (svc *WorkspaceService) UpdateWorkspace(ctx context.Context, workspaceId string, name string, icon string, color string, applyDefaults bool) (ainobj.UpdatesRtnType, error) {
	ctx = ainobj.ContextWithUpdates(ctx)
	_, updated, err := aincore.UpdateWorkspace(ctx, workspaceId, name, icon, color, applyDefaults)
	if err != nil {
		return nil, fmt.Errorf("error updating workspace: %w", err)
	}
	if !updated {
		return nil, nil
	}

	ainps.Broker.Publish(ainps.WaveEvent{
		Event: ainps.Event_WorkspaceUpdate,
	})

	updates := ainobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WorkspaceService:UpdateWorkspace:SendUpdateEvents", recover())
		}()
		ainps.Broker.SendUpdateEvents(updates)
	}()
	return updates, nil
}

func (svc *WorkspaceService) GetWorkspace_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames:   []string{"workspaceId"},
		ReturnDesc: "workspace",
	}
}

func (svc *WorkspaceService) GetWorkspace(workspaceId string) (*ainobj.Workspace, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	ws, err := ainstore.DBGet[*ainobj.Workspace](ctx, workspaceId)
	if err != nil {
		return nil, fmt.Errorf("error getting workspace: %w", err)
	}
	return ws, nil
}

func (svc *WorkspaceService) DeleteWorkspace_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"workspaceId"},
	}
}

func (svc *WorkspaceService) DeleteWorkspace(workspaceId string) (ainobj.UpdatesRtnType, string, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	ctx = ainobj.ContextWithUpdates(ctx)
	deleted, claimableWorkspace, err := aincore.DeleteWorkspace(ctx, workspaceId, true)
	if claimableWorkspace != "" {
		return nil, claimableWorkspace, nil
	}
	if err != nil {
		return nil, claimableWorkspace, fmt.Errorf("error deleting workspace: %w", err)
	}
	if !deleted {
		return nil, claimableWorkspace, nil
	}
	updates := ainobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WorkspaceService:DeleteWorkspace:SendUpdateEvents", recover())
		}()
		ainps.Broker.SendUpdateEvents(updates)
	}()
	return updates, claimableWorkspace, nil
}

func (svc *WorkspaceService) ListWorkspaces() (ainobj.WorkspaceList, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	return aincore.ListWorkspaces(ctx)
}

func (svc *WorkspaceService) CreateTab_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames:   []string{"workspaceId", "tabName", "activateTab"},
		ReturnDesc: "tabId",
	}
}

func (svc *WorkspaceService) GetColors_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ReturnDesc: "colors",
	}
}

func (svc *WorkspaceService) GetColors() []string {
	return aincore.WorkspaceColors[:]
}

func (svc *WorkspaceService) GetIcons_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ReturnDesc: "icons",
	}
}

func (svc *WorkspaceService) GetIcons() []string {
	return aincore.WorkspaceIcons[:]
}

func (svc *WorkspaceService) CreateTab(workspaceId string, tabName string, activateTab bool) (string, ainobj.UpdatesRtnType, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	ctx = ainobj.ContextWithUpdates(ctx)
	tabId, err := aincore.CreateTab(ctx, workspaceId, tabName, activateTab, false)
	if err != nil {
		return "", nil, fmt.Errorf("error creating tab: %w", err)
	}
	updates := ainobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WorkspaceService:CreateTab:SendUpdateEvents", recover())
		}()
		ainps.Broker.SendUpdateEvents(updates)
	}()
	return tabId, updates, nil
}

func (svc *WorkspaceService) UpdateTabIds_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"uiContext", "workspaceId", "tabIds"},
	}
}

func (svc *WorkspaceService) UpdateTabIds(uiContext ainobj.UIContext, workspaceId string, tabIds []string) (ainobj.UpdatesRtnType, error) {
	log.Printf("UpdateTabIds %s %v\n", workspaceId, tabIds)
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	ctx = ainobj.ContextWithUpdates(ctx)
	err := aincore.UpdateWorkspaceTabIds(ctx, workspaceId, tabIds)
	if err != nil {
		return nil, fmt.Errorf("error updating workspace tab ids: %w", err)
	}
	return ainobj.ContextGetUpdatesRtn(ctx), nil
}

func (svc *WorkspaceService) SetActiveTab_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"workspaceId", "tabId"},
	}
}

func (svc *WorkspaceService) SetActiveTab(workspaceId string, tabId string) (ainobj.UpdatesRtnType, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	ctx = ainobj.ContextWithUpdates(ctx)
	err := aincore.SetActiveTab(ctx, workspaceId, tabId)
	if err != nil {
		return nil, fmt.Errorf("error setting active tab: %w", err)
	}
	// check all blocks in tab and start controllers (if necessary)
	tab, err := ainstore.DBMustGet[*ainobj.Tab](ctx, tabId)
	if err != nil {
		return nil, fmt.Errorf("error getting tab: %w", err)
	}
	blockORefs := tab.GetBlockORefs()
	blocks, err := ainstore.DBSelectORefs(ctx, blockORefs)
	if err != nil {
		return nil, fmt.Errorf("error getting tab blocks: %w", err)
	}
	updates := ainobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WorkspaceService:SetActiveTab:SendUpdateEvents", recover())
		}()
		ainps.Broker.SendUpdateEvents(updates)
	}()
	var extraUpdates ainobj.UpdatesRtnType
	extraUpdates = append(extraUpdates, updates...)
	extraUpdates = append(extraUpdates, ainobj.MakeUpdate(tab))
	extraUpdates = append(extraUpdates, ainobj.MakeUpdates(blocks)...)
	return extraUpdates, nil
}

type CloseTabRtnType struct {
	CloseWindow    bool   `json:"closewindow,omitempty"`
	NewActiveTabId string `json:"newactivetabid,omitempty"`
}

func (svc *WorkspaceService) CloseTab_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames:   []string{"ctx", "workspaceId", "tabId", "fromElectron"},
		ReturnDesc: "CloseTabRtn",
	}
}

// returns the new active tabid
func (svc *WorkspaceService) CloseTab(ctx context.Context, workspaceId string, tabId string, fromElectron bool) (*CloseTabRtnType, ainobj.UpdatesRtnType, error) {
	ctx = ainobj.ContextWithUpdates(ctx)
	tab, err := ainstore.DBGet[*ainobj.Tab](ctx, tabId)
	if err == nil && tab != nil {
		go func() {
			for _, blockId := range tab.BlockIds {
				blockcontroller.StopBlockController(blockId)
			}
		}()
	}
	newActiveTabId, err := aincore.DeleteTab(ctx, workspaceId, tabId, true)
	if err != nil {
		return nil, nil, fmt.Errorf("error closing tab: %w", err)
	}
	rtn := &CloseTabRtnType{}
	if newActiveTabId == "" {
		rtn.CloseWindow = true
	} else {
		rtn.NewActiveTabId = newActiveTabId
	}
	updates := ainobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WorkspaceService:CloseTab:SendUpdateEvents", recover())
		}()
		ainps.Broker.SendUpdateEvents(updates)
	}()
	return rtn, updates, nil
}
