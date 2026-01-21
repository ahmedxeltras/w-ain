// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package clientservice

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wavetermdev/ainterm/pkg/ainconfig"
	"github.com/wavetermdev/ainterm/pkg/aincore"
	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainstore"
	"github.com/wavetermdev/ainterm/pkg/remote/conncontroller"
	"github.com/wavetermdev/ainterm/pkg/wslconn"
)

type ClientService struct{}

const DefaultTimeout = 2 * time.Second

func (cs *ClientService) GetClientData() (*ainobj.Client, error) {
	log.Println("GetClientData")
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	return aincore.GetClientData(ctx)
}

func (cs *ClientService) GetTab(tabId string) (*ainobj.Tab, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	tab, err := ainstore.DBGet[*ainobj.Tab](ctx, tabId)
	if err != nil {
		return nil, fmt.Errorf("error getting tab: %w", err)
	}
	return tab, nil
}

func (cs *ClientService) GetAllConnStatus(ctx context.Context) ([]ainshrpc.ConnStatus, error) {
	sshStatuses := conncontroller.GetAllConnStatus()
	wslStatuses := wslconn.GetAllConnStatus()
	return append(sshStatuses, wslStatuses...), nil
}

// moves the window to the front of the windowId stack
func (cs *ClientService) FocusWindow(ctx context.Context, windowId string) error {
	return aincore.FocusWindow(ctx, windowId)
}

func (cs *ClientService) AgreeTos(ctx context.Context) (ainobj.UpdatesRtnType, error) {
	ctx = ainobj.ContextWithUpdates(ctx)
	clientData, err := ainstore.DBGetSingleton[*ainobj.Client](ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting client data: %w", err)
	}
	timestamp := time.Now().UnixMilli()
	clientData.TosAgreed = timestamp
	err = ainstore.DBUpdate(ctx, clientData)
	if err != nil {
		return nil, fmt.Errorf("error updating client data: %w", err)
	}
	aincore.BootstrapStarterLayout(ctx)
	return ainobj.ContextGetUpdatesRtn(ctx), nil
}

func (cs *ClientService) TelemetryUpdate(ctx context.Context, telemetryEnabled bool) error {
	meta := ainobj.MetaMapType{
		ainconfig.ConfigKey_TelemetryEnabled: telemetryEnabled,
	}
	err := ainconfig.SetBaseConfigValue(meta)
	if err != nil {
		return fmt.Errorf("error setting telemetry value: %w", err)
	}
	return nil
}
