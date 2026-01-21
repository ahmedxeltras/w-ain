// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package aincore

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/wavetermdev/ainterm/pkg/ainconfig"
	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainps"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainstore"
	"github.com/wavetermdev/ainterm/pkg/eventbus"
	"github.com/wavetermdev/ainterm/pkg/telemetry"
	"github.com/wavetermdev/ainterm/pkg/telemetry/telemetrydata"
	"github.com/wavetermdev/ainterm/pkg/util/utilfn"
)

var WorkspaceColors = [...]string{
	"#58C142", // Green (accent)
	"#00FFDB", // Teal
	"#429DFF", // Blue
	"#BF55EC", // Purple
	"#FF453A", // Red
	"#FF9500", // Orange
	"#FFE900", // Yellow
}

var WorkspaceIcons = [...]string{
	"custom@wave-logo-solid",
	"triangle",
	"star",
	"heart",
	"bolt",
	"solid@cloud",
	"moon",
	"layer-group",
	"rocket",
	"flask",
	"paperclip",
	"chart-line",
	"graduation-cap",
	"mug-hot",
}

func CreateWorkspace(ctx context.Context, name string, icon string, color string, applyDefaults bool, isInitialLaunch bool) (*ainobj.Workspace, error) {
	ws := &ainobj.Workspace{
		OID:    uuid.NewString(),
		TabIds: []string{},
		Name:   "",
		Icon:   "",
		Color:  "",
	}
	err := ainstore.DBInsert(ctx, ws)
	if err != nil {
		return nil, fmt.Errorf("error inserting workspace: %w", err)
	}
	_, err = CreateTab(ctx, ws.OID, "", true, isInitialLaunch)
	if err != nil {
		return nil, fmt.Errorf("error creating tab: %w", err)
	}

	ainps.Broker.Publish(ainps.WaveEvent{
		Event: ainps.Event_WorkspaceUpdate,
	})

	ws, _, err = UpdateWorkspace(ctx, ws.OID, name, icon, color, applyDefaults)
	return ws, err
}

// Returns updated workspace, whether it was updated, error.
func UpdateWorkspace(ctx context.Context, workspaceId string, name string, icon string, color string, applyDefaults bool) (*ainobj.Workspace, bool, error) {
	ws, err := GetWorkspace(ctx, workspaceId)
	updated := false
	if err != nil {
		return nil, updated, fmt.Errorf("workspace %s not found: %w", workspaceId, err)
	}
	if name != "" {
		ws.Name = name
		updated = true
	} else if applyDefaults && ws.Name == "" {
		ws.Name = fmt.Sprintf("New Workspace (%s)", ws.OID[0:5])
		updated = true
	}
	if icon != "" {
		ws.Icon = icon
		updated = true
	} else if applyDefaults && ws.Icon == "" {
		ws.Icon = WorkspaceIcons[0]
		updated = true
	}
	if color != "" {
		ws.Color = color
		updated = true
	} else if applyDefaults && ws.Color == "" {
		wsList, err := ListWorkspaces(ctx)
		if err != nil {
			log.Printf("error listing workspaces: %v", err)
			wsList = ainobj.WorkspaceList{}
		}
		ws.Color = WorkspaceColors[len(wsList)%len(WorkspaceColors)]
		updated = true
	}
	if updated {
		ainstore.DBUpdate(ctx, ws)
	}
	return ws, updated, nil
}

// If force is true, it will delete even if workspace is named.
// If workspace is empty, it will be deleted, even if it is named.
// Returns true if workspace was deleted, false if it was not deleted.
func DeleteWorkspace(ctx context.Context, workspaceId string, force bool) (bool, string, error) {
	log.Printf("DeleteWorkspace %s\n", workspaceId)
	workspace, err := ainstore.DBMustGet[*ainobj.Workspace](ctx, workspaceId)
	if err != nil && ainstore.ErrNotFound == err {
		return true, "", fmt.Errorf("workspace already deleted %w", err)
	}
	// @jalileh list needs to be saved early on i assume
	workspaces, err := ListWorkspaces(ctx)
	if err != nil {
		return false, "", fmt.Errorf("error retrieving workspaceList: %w", err)
	}

	if workspace.Name != "" && workspace.Icon != "" && !force && len(workspace.TabIds) > 0 {
		log.Printf("Ignoring DeleteWorkspace for workspace %s as it is named\n", workspaceId)
		return false, "", nil
	}

	for _, tabId := range workspace.TabIds {
		log.Printf("deleting tab %s\n", tabId)
		_, err := DeleteTab(ctx, workspaceId, tabId, false)
		if err != nil {
			return false, "", fmt.Errorf("error closing tab: %w", err)
		}
	}
	windowId, _ := ainstore.DBFindWindowForWorkspaceId(ctx, workspaceId)
	err = ainstore.DBDelete(ctx, ainobj.OType_Workspace, workspaceId)
	if err != nil {
		return false, "", fmt.Errorf("error deleting workspace: %w", err)
	}
	log.Printf("deleted workspace %s\n", workspaceId)
	ainps.Broker.Publish(ainps.WaveEvent{
		Event: ainps.Event_WorkspaceUpdate,
	})

	if windowId != "" {

		UnclaimedWorkspace, findAfter := "", false
		for _, ws := range workspaces {
			if ws.WorkspaceId == workspaceId {
				if UnclaimedWorkspace != "" {
					break
				}
				findAfter = true
				continue
			}
			if findAfter && ws.WindowId == "" {
				UnclaimedWorkspace = ws.WorkspaceId
				break
			} else if ws.WindowId == "" {
				UnclaimedWorkspace = ws.WorkspaceId
			}
		}

		if UnclaimedWorkspace != "" {
			return true, UnclaimedWorkspace, nil
		} else {
			err = CloseWindow(ctx, windowId, false)
		}

		if err != nil {
			return false, "", fmt.Errorf("error closing window: %w", err)
		}
	}
	return true, "", nil
}

func GetWorkspace(ctx context.Context, wsID string) (*ainobj.Workspace, error) {
	return ainstore.DBMustGet[*ainobj.Workspace](ctx, wsID)
}

func getTabPresetMeta() (ainobj.MetaMapType, error) {
	settings := ainconfig.GetWatcher().GetFullConfig()
	tabPreset := settings.Settings.TabPreset
	if tabPreset == "" {
		return nil, nil
	}
	presetMeta := settings.Presets[tabPreset]
	return presetMeta, nil
}

// returns tabid
func CreateTab(ctx context.Context, workspaceId string, tabName string, activateTab bool, isInitialLaunch bool) (string, error) {
	if tabName == "" {
		ws, err := GetWorkspace(ctx, workspaceId)
		if err != nil {
			return "", fmt.Errorf("workspace %s not found: %w", workspaceId, err)
		}
		tabName = "T" + fmt.Sprint(len(ws.TabIds)+1)
	}

	tab, err := createTabObj(ctx, workspaceId, tabName, nil)
	if err != nil {
		return "", fmt.Errorf("error creating tab: %w", err)
	}
	if activateTab {
		err = SetActiveTab(ctx, workspaceId, tab.OID)
		if err != nil {
			return "", fmt.Errorf("error setting active tab: %w", err)
		}
	}

	// No need to apply an initial layout for the initial launch, since the starter layout will get applied after onboarding modal dismissal
	if !isInitialLaunch {
		err = ApplyPortableLayout(ctx, tab.OID, GetNewTabLayout(), true)
		if err != nil {
			return tab.OID, fmt.Errorf("error applying new tab layout: %w", err)
		}
		presetMeta, presetErr := getTabPresetMeta()
		if presetErr != nil {
			log.Printf("error getting tab preset meta: %v\n", presetErr)
		} else if len(presetMeta) > 0 {
			tabORef := ainobj.ORefFromWaveObj(tab)
			ainstore.UpdateObjectMeta(ctx, *tabORef, presetMeta, true)
		}
	}
	telemetry.GoUpdateActivityWrap(ainshrpc.ActivityUpdate{NewTab: 1}, "createtab")
	telemetry.GoRecordTEventWrap(&telemetrydata.TEvent{
		Event: "action:createtab",
	})
	return tab.OID, nil
}

func createTabObj(ctx context.Context, workspaceId string, name string, meta ainobj.MetaMapType) (*ainobj.Tab, error) {
	ws, err := GetWorkspace(ctx, workspaceId)
	if err != nil {
		return nil, fmt.Errorf("workspace %s not found: %w", workspaceId, err)
	}
	layoutStateId := uuid.NewString()
	tab := &ainobj.Tab{
		OID:         uuid.NewString(),
		Name:        name,
		BlockIds:    []string{},
		LayoutState: layoutStateId,
		Meta:        meta,
	}
	layoutState := &ainobj.LayoutState{
		OID: layoutStateId,
	}
	ws.TabIds = append(ws.TabIds, tab.OID)
	ainstore.DBInsert(ctx, tab)
	ainstore.DBInsert(ctx, layoutState)
	ainstore.DBUpdate(ctx, ws)
	return tab, nil
}

// Must delete all blocks individually first.
// Also deletes LayoutState.
// recursive: if true, will recursively close parent window, workspace, if they are empty.
// Returns new active tab id, error.
func DeleteTab(ctx context.Context, workspaceId string, tabId string, recursive bool) (string, error) {
	ws, _ := ainstore.DBGet[*ainobj.Workspace](ctx, workspaceId)
	if ws == nil {
		return "", fmt.Errorf("workspace not found: %q", workspaceId)
	}

	// ensure tab is in workspace
	tabIdx := utilfn.FindStringInSlice(ws.TabIds, tabId)
	if tabIdx == -1 {
		return "", fmt.Errorf("tab %s not found in workspace %s", tabId, workspaceId)
	}
	ws.TabIds = append(ws.TabIds[:tabIdx], ws.TabIds[tabIdx+1:]...)

	// close blocks (sends events + stops block controllers)
	tab, _ := ainstore.DBGet[*ainobj.Tab](ctx, tabId)
	if tab != nil {
		for _, blockId := range tab.BlockIds {
			err := DeleteBlock(ctx, blockId, false)
			if err != nil {
				return "", fmt.Errorf("error deleting block %s: %w", blockId, err)
			}
		}
	}

	// if the tab is active, determine new active tab
	newActiveTabId := ws.ActiveTabId
	if ws.ActiveTabId == tabId {
		if len(ws.TabIds) > 0 {
			newActiveTabId = ws.TabIds[max(0, min(tabIdx-1, len(ws.TabIds)-1))]
		} else {
			newActiveTabId = ""
		}
	}
	ws.ActiveTabId = newActiveTabId

	ainstore.DBUpdate(ctx, ws)
	ainstore.DBDelete(ctx, ainobj.OType_Tab, tabId)
	if tab != nil {
		ainstore.DBDelete(ctx, ainobj.OType_LayoutState, tab.LayoutState)
	}

	// if no tabs remaining, close window
	// DISABLED: Keep window open and show logo instead
	// if recursive && newActiveTabId == "" {
	// 	log.Printf("no tabs remaining in workspace %s, closing window\n", workspaceId)
	// 	windowId, err := ainstore.DBFindWindowForWorkspaceId(ctx, workspaceId)
	// 	if err != nil {
	// 		return newActiveTabId, fmt.Errorf("unable to find window for workspace id %v: %w", workspaceId, err)
	// 	}
	// 	err = CloseWindow(ctx, windowId, false)
	// 	if err != nil {
	// 		return newActiveTabId, err
	// 	}
	// }
	return newActiveTabId, nil
}

func SetActiveTab(ctx context.Context, workspaceId string, tabId string) error {
	if tabId != "" && workspaceId != "" {
		workspace, err := GetWorkspace(ctx, workspaceId)
		if err != nil {
			return fmt.Errorf("workspace %s not found: %w", workspaceId, err)
		}
		tab, _ := ainstore.DBGet[*ainobj.Tab](ctx, tabId)
		if tab == nil {
			return fmt.Errorf("tab not found: %q", tabId)
		}
		workspace.ActiveTabId = tabId
		ainstore.DBUpdate(ctx, workspace)
	}
	return nil
}

func SendActiveTabUpdate(ctx context.Context, workspaceId string, newActiveTabId string) {
	eventbus.SendEventToElectron(eventbus.WSEventType{
		EventType: eventbus.WSEvent_ElectronUpdateActiveTab,
		Data:      &ainobj.ActiveTabUpdate{WorkspaceId: workspaceId, NewActiveTabId: newActiveTabId},
	})
}

func UpdateWorkspaceTabIds(ctx context.Context, workspaceId string, tabIds []string) error {
	ws, _ := ainstore.DBGet[*ainobj.Workspace](ctx, workspaceId)
	if ws == nil {
		return fmt.Errorf("workspace not found: %q", workspaceId)
	}
	ws.TabIds = tabIds
	ainstore.DBUpdate(ctx, ws)
	return nil
}

func ListWorkspaces(ctx context.Context) (ainobj.WorkspaceList, error) {
	workspaces, err := ainstore.DBGetAllObjsByType[*ainobj.Workspace](ctx, ainobj.OType_Workspace)
	if err != nil {
		return nil, err
	}
	windows, err := ainstore.DBGetAllObjsByType[*ainobj.Window](ctx, ainobj.OType_Window)
	if err != nil {
		return nil, err
	}
	workspaceToWindow := make(map[string]string)
	for _, window := range windows {
		workspaceToWindow[window.WorkspaceId] = window.OID
	}

	var wl ainobj.WorkspaceList
	for _, workspace := range workspaces {
		if workspace.Name == "" || workspace.Icon == "" || workspace.Color == "" {
			continue
		}
		windowId, ok := workspaceToWindow[workspace.OID]
		if !ok {
			windowId = ""
		}
		wl = append(wl, &ainobj.WorkspaceListEntry{
			WorkspaceId: workspace.OID,
			WindowId:    windowId,
		})
	}
	return wl, nil
}

func SetIcon(workspaceId string, icon string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ws, e := ainstore.DBGet[*ainobj.Workspace](ctx, workspaceId)
	if e != nil {
		return e
	}
	if ws == nil {
		return fmt.Errorf("workspace not found: %q", workspaceId)
	}
	ws.Icon = icon
	ainstore.DBUpdate(ctx, ws)
	return nil
}

func SetColor(workspaceId string, color string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ws, e := ainstore.DBGet[*ainobj.Workspace](ctx, workspaceId)
	if e != nil {
		return e
	}
	if ws == nil {
		return fmt.Errorf("workspace not found: %q", workspaceId)
	}
	ws.Color = color
	ainstore.DBUpdate(ctx, ws)
	return nil
}

func SetName(workspaceId string, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ws, e := ainstore.DBGet[*ainobj.Workspace](ctx, workspaceId)
	if e != nil {
		return e
	}
	if ws == nil {
		return fmt.Errorf("workspace not found: %q", workspaceId)
	}
	ws.Name = name
	ainstore.DBUpdate(ctx, ws)
	return nil
}
