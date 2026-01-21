// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package wshserver

// this file contains the implementation of the wsh server methods

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/skratchdot/open-golang/open"
	"github.com/wavetermdev/ainterm/pkg/ainai"
	"github.com/wavetermdev/ainterm/pkg/ainappstore"
	"github.com/wavetermdev/ainterm/pkg/ainapputil"
	"github.com/wavetermdev/ainterm/pkg/ainbase"
	"github.com/wavetermdev/ainterm/pkg/aincloud"
	"github.com/wavetermdev/ainterm/pkg/ainconfig"
	"github.com/wavetermdev/ainterm/pkg/aincore"
	"github.com/wavetermdev/ainterm/pkg/ainjwt"
	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainps"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainshutil"
	"github.com/wavetermdev/ainterm/pkg/ainstore"
	"github.com/wavetermdev/ainterm/pkg/aiusechat"
	"github.com/wavetermdev/ainterm/pkg/aiusechat/chatstore"
	"github.com/wavetermdev/ainterm/pkg/aiusechat/uctypes"
	"github.com/wavetermdev/ainterm/pkg/blockcontroller"
	"github.com/wavetermdev/ainterm/pkg/blocklogger"
	"github.com/wavetermdev/ainterm/pkg/buildercontroller"
	"github.com/wavetermdev/ainterm/pkg/filebackup"
	"github.com/wavetermdev/ainterm/pkg/filestore"
	"github.com/wavetermdev/ainterm/pkg/genconn"
	"github.com/wavetermdev/ainterm/pkg/panichandler"
	"github.com/wavetermdev/ainterm/pkg/remote"
	"github.com/wavetermdev/ainterm/pkg/remote/awsconn"
	"github.com/wavetermdev/ainterm/pkg/remote/conncontroller"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare"
	"github.com/wavetermdev/ainterm/pkg/secretstore"
	"github.com/wavetermdev/ainterm/pkg/suggestion"
	"github.com/wavetermdev/ainterm/pkg/telemetry"
	"github.com/wavetermdev/ainterm/pkg/telemetry/telemetrydata"
	"github.com/wavetermdev/ainterm/pkg/util/envutil"
	"github.com/wavetermdev/ainterm/pkg/util/iochan/iochantypes"
	"github.com/wavetermdev/ainterm/pkg/util/iterfn"
	"github.com/wavetermdev/ainterm/pkg/util/shellutil"
	"github.com/wavetermdev/ainterm/pkg/util/utilfn"
	"github.com/wavetermdev/ainterm/pkg/util/wavefileutil"
	"github.com/wavetermdev/ainterm/pkg/wsl"
	"github.com/wavetermdev/ainterm/pkg/wslconn"
	"github.com/wavetermdev/ainterm/tsunami/build"
)

var InvalidWslDistroNames = []string{"docker-desktop", "docker-desktop-data"}

type WshServer struct{}

func (*WshServer) WshServerImpl() {}

var WshServerImpl = WshServer{}

func (ws *WshServer) GetJwtPublicKeyCommand(ctx context.Context) (string, error) {
	return ainjwt.GetPublicKeyBase64(), nil
}

func (ws *WshServer) TestCommand(ctx context.Context, data string) error {
	defer func() {
		panichandler.PanicHandler("TestCommand", recover())
	}()
	rpcSource := ainshutil.GetRpcSourceFromContext(ctx)
	log.Printf("TEST src:%s | %s\n", rpcSource, data)
	return nil
}

// for testing
func (ws *WshServer) MessageCommand(ctx context.Context, data ainshrpc.CommandMessageData) error {
	log.Printf("MESSAGE: %s\n", data.Message)
	return nil
}

// for testing
func (ws *WshServer) StreamTestCommand(ctx context.Context) chan ainshrpc.RespOrErrorUnion[int] {
	rtn := make(chan ainshrpc.RespOrErrorUnion[int])
	go func() {
		defer func() {
			panichandler.PanicHandler("StreamTestCommand", recover())
		}()
		for i := 1; i <= 5; i++ {
			rtn <- ainshrpc.RespOrErrorUnion[int]{Response: i}
			time.Sleep(1 * time.Second)
		}
		close(rtn)
	}()
	return rtn
}

func (ws *WshServer) StreamAinAiCommand(ctx context.Context, request ainshrpc.WaveAIStreamRequest) chan ainshrpc.RespOrErrorUnion[ainshrpc.WaveAIPacketType] {
return ainai.RunAICommand(ctx, request)
}


func MakePlotData(ctx context.Context, blockId string) error {
	block, err := ainstore.DBMustGet[*ainobj.Block](ctx, blockId)
	if err != nil {
		return err
	}
	viewName := block.Meta.GetString(ainobj.MetaKey_View, "")
	if viewName != "cpuplot" && viewName != "sysinfo" {
		return fmt.Errorf("invalid view type: %s", viewName)
	}
	return filestore.WFS.MakeFile(ctx, blockId, "cpuplotdata", nil, ainshrpc.FileOpts{})
}

func SavePlotData(ctx context.Context, blockId string, history string) error {
	block, err := ainstore.DBMustGet[*ainobj.Block](ctx, blockId)
	if err != nil {
		return err
	}
	viewName := block.Meta.GetString(ainobj.MetaKey_View, "")
	if viewName != "cpuplot" && viewName != "sysinfo" {
		return fmt.Errorf("invalid view type: %s", viewName)
	}
	// todo: interpret the data being passed
	// for now, this is just to throw an error if the block was closed
	historyBytes, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("unable to serialize plot data: %v", err)
	}
	// ignore MakeFile error (already exists is ok)
	return filestore.WFS.WriteFile(ctx, blockId, "cpuplotdata", historyBytes)
}

func (ws *WshServer) GetMetaCommand(ctx context.Context, data ainshrpc.CommandGetMetaData) (ainobj.MetaMapType, error) {
	obj, err := ainstore.DBGetORef(ctx, data.ORef)
	if err != nil {
		return nil, fmt.Errorf("error getting object: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("object not found: %s", data.ORef)
	}
	return ainobj.GetMeta(obj), nil
}

func (ws *WshServer) SetMetaCommand(ctx context.Context, data ainshrpc.CommandSetMetaData) error {
	log.Printf("SetMetaCommand: %s | %v\n", data.ORef, data.Meta)
	oref := data.ORef
	err := ainstore.UpdateObjectMeta(ctx, oref, data.Meta, false)
	if err != nil {
		return fmt.Errorf("error updating object meta: %w", err)
	}
	aincore.SendWaveObjUpdate(oref)
	return nil
}

func (ws *WshServer) GetRTInfoCommand(ctx context.Context, data ainshrpc.CommandGetRTInfoData) (*ainobj.ObjRTInfo, error) {
	return ainstore.GetRTInfo(data.ORef), nil
}

func (ws *WshServer) SetRTInfoCommand(ctx context.Context, data ainshrpc.CommandSetRTInfoData) error {
	if data.Delete {
		ainstore.DeleteRTInfo(data.ORef)
		return nil
	}
	ainstore.SetRTInfo(data.ORef, data.Data)
	return nil
}

func (ws *WshServer) ResolveIdsCommand(ctx context.Context, data ainshrpc.CommandResolveIdsData) (ainshrpc.CommandResolveIdsRtnData, error) {
	rtn := ainshrpc.CommandResolveIdsRtnData{}
	rtn.ResolvedIds = make(map[string]ainobj.ORef)
	var firstErr error
	for _, simpleId := range data.Ids {
		oref, err := resolveSimpleId(ctx, data, simpleId)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if oref == nil {
			continue
		}
		rtn.ResolvedIds[simpleId] = *oref
	}
	if firstErr != nil && len(data.Ids) == 1 {
		return rtn, firstErr
	}
	return rtn, nil
}

func (ws *WshServer) CreateBlockCommand(ctx context.Context, data ainshrpc.CommandCreateBlockData) (*ainobj.ORef, error) {
	ctx = ainobj.ContextWithUpdates(ctx)
	tabId := data.TabId
	blockData, err := aincore.CreateBlock(ctx, tabId, data.BlockDef, data.RtOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating block: %w", err)
	}
	var layoutAction *ainobj.LayoutActionData
	if data.TargetBlockId != "" {
		switch data.TargetAction {
		case "replace":
			layoutAction = &ainobj.LayoutActionData{
				ActionType:    aincore.LayoutActionDataType_Replace,
				TargetBlockId: data.TargetBlockId,
				BlockId:       blockData.OID,
				Focused:       data.Focused,
			}
			err = aincore.DeleteBlock(ctx, data.TargetBlockId, false)
			if err != nil {
				return nil, fmt.Errorf("error deleting block (trying to do block replace): %w", err)
			}
		case "splitright":
			layoutAction = &ainobj.LayoutActionData{
				ActionType:    aincore.LayoutActionDataType_SplitHorizontal,
				BlockId:       blockData.OID,
				TargetBlockId: data.TargetBlockId,
				Position:      "after",
				Focused:       data.Focused,
			}
		case "splitleft":
			layoutAction = &ainobj.LayoutActionData{
				ActionType:    aincore.LayoutActionDataType_SplitHorizontal,
				BlockId:       blockData.OID,
				TargetBlockId: data.TargetBlockId,
				Position:      "before",
				Focused:       data.Focused,
			}
		case "splitup":
			layoutAction = &ainobj.LayoutActionData{
				ActionType:    aincore.LayoutActionDataType_SplitVertical,
				BlockId:       blockData.OID,
				TargetBlockId: data.TargetBlockId,
				Position:      "before",
				Focused:       data.Focused,
			}
		case "splitdown":
			layoutAction = &ainobj.LayoutActionData{
				ActionType:    aincore.LayoutActionDataType_SplitVertical,
				BlockId:       blockData.OID,
				TargetBlockId: data.TargetBlockId,
				Position:      "after",
				Focused:       data.Focused,
			}
		default:
			return nil, fmt.Errorf("invalid target action: %s", data.TargetAction)
		}
	} else {
		layoutAction = &ainobj.LayoutActionData{
			ActionType: aincore.LayoutActionDataType_Insert,
			BlockId:    blockData.OID,
			Magnified:  data.Magnified,
			Ephemeral:  data.Ephemeral,
			Focused:    data.Focused,
		}
	}
	err = aincore.QueueLayoutActionForTab(ctx, tabId, *layoutAction)
	if err != nil {
		return nil, fmt.Errorf("error queuing layout action: %w", err)
	}
	updates := ainobj.ContextGetUpdatesRtn(ctx)
	ainps.Broker.SendUpdateEvents(updates)
	return &ainobj.ORef{OType: ainobj.OType_Block, OID: blockData.OID}, nil
}

func (ws *WshServer) CreateSubBlockCommand(ctx context.Context, data ainshrpc.CommandCreateSubBlockData) (*ainobj.ORef, error) {
	parentBlockId := data.ParentBlockId
	blockData, err := aincore.CreateSubBlock(ctx, parentBlockId, data.BlockDef)
	if err != nil {
		return nil, fmt.Errorf("error creating block: %w", err)
	}
	blockRef := &ainobj.ORef{OType: ainobj.OType_Block, OID: blockData.OID}
	return blockRef, nil
}

func (ws *WshServer) ControllerStopCommand(ctx context.Context, blockId string) error {
	blockcontroller.StopBlockController(blockId)
	return nil
}

func (ws *WshServer) ControllerResyncCommand(ctx context.Context, data ainshrpc.CommandControllerResyncData) error {
	ctx = genconn.ContextWithConnData(ctx, data.BlockId)
	ctx = termCtxWithLogBlockId(ctx, data.BlockId)
	return blockcontroller.ResyncController(ctx, data.TabId, data.BlockId, data.RtOpts, data.ForceRestart)
}

func (ws *WshServer) ControllerInputCommand(ctx context.Context, data ainshrpc.CommandBlockInputData) error {
	inputUnion := &blockcontroller.BlockInputUnion{
		SigName:  data.SigName,
		TermSize: data.TermSize,
	}
	if len(data.InputData64) > 0 {
		inputBuf := make([]byte, base64.StdEncoding.DecodedLen(len(data.InputData64)))
		nw, err := base64.StdEncoding.Decode(inputBuf, []byte(data.InputData64))
		if err != nil {
			return fmt.Errorf("error decoding input data: %w", err)
		}
		inputUnion.InputData = inputBuf[:nw]
	}
	return blockcontroller.SendInput(data.BlockId, inputUnion)
}

func (ws *WshServer) ControllerAppendOutputCommand(ctx context.Context, data ainshrpc.CommandControllerAppendOutputData) error {
	outputBuf := make([]byte, base64.StdEncoding.DecodedLen(len(data.Data64)))
	nw, err := base64.StdEncoding.Decode(outputBuf, []byte(data.Data64))
	if err != nil {
		return fmt.Errorf("error decoding output data: %w", err)
	}
	err = blockcontroller.HandleAppendBlockFile(data.BlockId, ainbase.BlockFile_Term, outputBuf[:nw])
	if err != nil {
		return fmt.Errorf("error appending to block file: %w", err)
	}
	return nil
}

func (ws *WshServer) FileCreateCommand(ctx context.Context, data ainshrpc.FileData) error {
	data.Data64 = ""
	err := fileshare.PutFile(ctx, data)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	return nil
}

func (ws *WshServer) FileMkdirCommand(ctx context.Context, data ainshrpc.FileData) error {
	return fileshare.Mkdir(ctx, data.Info.Path)
}

func (ws *WshServer) FileDeleteCommand(ctx context.Context, data ainshrpc.CommandDeleteFileData) error {
	return fileshare.Delete(ctx, data)
}

func (ws *WshServer) FileInfoCommand(ctx context.Context, data ainshrpc.FileData) (*ainshrpc.FileInfo, error) {
	return fileshare.Stat(ctx, data.Info.Path)
}

func (ws *WshServer) FileListCommand(ctx context.Context, data ainshrpc.FileListData) ([]*ainshrpc.FileInfo, error) {
	return fileshare.ListEntries(ctx, data.Path, data.Opts)
}

func (ws *WshServer) FileListStreamCommand(ctx context.Context, data ainshrpc.FileListData) <-chan ainshrpc.RespOrErrorUnion[ainshrpc.CommandRemoteListEntriesRtnData] {
	return fileshare.ListEntriesStream(ctx, data.Path, data.Opts)
}

func (ws *WshServer) FileWriteCommand(ctx context.Context, data ainshrpc.FileData) error {
	return fileshare.PutFile(ctx, data)
}

func (ws *WshServer) FileReadCommand(ctx context.Context, data ainshrpc.FileData) (*ainshrpc.FileData, error) {
	return fileshare.Read(ctx, data)
}

func (ws *WshServer) FileReadStreamCommand(ctx context.Context, data ainshrpc.FileData) <-chan ainshrpc.RespOrErrorUnion[ainshrpc.FileData] {
	return fileshare.ReadStream(ctx, data)
}

func (ws *WshServer) FileCopyCommand(ctx context.Context, data ainshrpc.CommandFileCopyData) error {
	return fileshare.Copy(ctx, data)
}

func (ws *WshServer) FileMoveCommand(ctx context.Context, data ainshrpc.CommandFileCopyData) error {
	return fileshare.Move(ctx, data)
}

func (ws *WshServer) FileStreamTarCommand(ctx context.Context, data ainshrpc.CommandRemoteStreamTarData) <-chan ainshrpc.RespOrErrorUnion[iochantypes.Packet] {
	return fileshare.ReadTarStream(ctx, data)
}

func (ws *WshServer) FileAppendCommand(ctx context.Context, data ainshrpc.FileData) error {
	return fileshare.Append(ctx, data)
}

func (ws *WshServer) FileAppendIJsonCommand(ctx context.Context, data ainshrpc.CommandAppendIJsonData) error {
	tryCreate := true
	if data.FileName == ainbase.BlockFile_VDom && tryCreate {
		err := filestore.WFS.MakeFile(ctx, data.ZoneId, data.FileName, nil, ainshrpc.FileOpts{MaxSize: blockcontroller.DefaultHtmlMaxFileSize, IJson: true})
		if err != nil && err != fs.ErrExist {
			return fmt.Errorf("error creating blockfile[vdom]: %w", err)
		}
	}
	err := filestore.WFS.AppendIJson(ctx, data.ZoneId, data.FileName, data.Data)
	if err != nil {
		return fmt.Errorf("error appending to blockfile(ijson): %w", err)
	}
	ainps.Broker.Publish(ainps.WaveEvent{
		Event:  ainps.Event_BlockFile,
		Scopes: []string{ainobj.MakeORef(ainobj.OType_Block, data.ZoneId).String()},
		Data: &ainps.WSFileEventData{
			ZoneId:   data.ZoneId,
			FileName: data.FileName,
			FileOp:   ainps.FileOp_Append,
			Data64:   base64.StdEncoding.EncodeToString([]byte("{}")),
		},
	})
	return nil
}

func (ws *WshServer) FileJoinCommand(ctx context.Context, paths []string) (*ainshrpc.FileInfo, error) {
	if len(paths) < 2 {
		if len(paths) == 0 {
			return nil, fmt.Errorf("no paths provided")
		}
		return fileshare.Stat(ctx, paths[0])
	}
	return fileshare.Join(ctx, paths[0], paths[1:]...)
}

func (ws *WshServer) FileShareCapabilityCommand(ctx context.Context, path string) (ainshrpc.FileShareCapability, error) {
	return fileshare.GetCapability(ctx, path)
}

func (ws *WshServer) FileRestoreBackupCommand(ctx context.Context, data ainshrpc.CommandFileRestoreBackupData) error {
	expandedBackupPath, err := ainbase.ExpandHomeDir(data.BackupFilePath)
	if err != nil {
		return fmt.Errorf("failed to expand backup file path: %w", err)
	}
	expandedRestorePath, err := ainbase.ExpandHomeDir(data.RestoreToFileName)
	if err != nil {
		return fmt.Errorf("failed to expand restore file path: %w", err)
	}
	return filebackup.RestoreBackup(expandedBackupPath, expandedRestorePath)
}

func (ws *WshServer) GetTempDirCommand(ctx context.Context, data ainshrpc.CommandGetTempDirData) (string, error) {
	tempDir := os.TempDir()
	if data.FileName != "" {
		// Reduce to a simple file name to avoid absolute paths or traversal
		name := filepath.Base(data.FileName)
		// Normalize/trim any stray separators and whitespace
		name = strings.Trim(name, `/\`+" ")
		if name == "" || name == "." {
			return tempDir, nil
		}
		return filepath.Join(tempDir, name), nil
	}
	return tempDir, nil
}

func (ws *WshServer) WriteTempFileCommand(ctx context.Context, data ainshrpc.CommandWriteTempFileData) (string, error) {
	if data.FileName == "" {
		return "", fmt.Errorf("filename is required")
	}
	name := filepath.Base(data.FileName)
	if name == "" || name == "." || name == ".." {
		return "", fmt.Errorf("invalid filename")
	}
	tempDir, err := os.MkdirTemp("", "waveterm-")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(data.Data64)
	if err != nil {
		return "", fmt.Errorf("error decoding base64 data: %w", err)
	}
	tempPath := filepath.Join(tempDir, name)
	err = os.WriteFile(tempPath, decoded, 0600)
	if err != nil {
		return "", fmt.Errorf("error writing temp file: %w", err)
	}
	return tempPath, nil
}

func (ws *WshServer) DeleteSubBlockCommand(ctx context.Context, data ainshrpc.CommandDeleteBlockData) error {
	if data.BlockId == "" {
		return fmt.Errorf("blockid is required")
	}
	err := aincore.DeleteBlock(ctx, data.BlockId, false)
	if err != nil {
		return fmt.Errorf("error deleting block: %w", err)
	}
	return nil
}

func (ws *WshServer) DeleteBlockCommand(ctx context.Context, data ainshrpc.CommandDeleteBlockData) error {
	if data.BlockId == "" {
		return fmt.Errorf("blockid is required")
	}
	ctx = ainobj.ContextWithUpdates(ctx)
	tabId, err := ainstore.DBFindTabForBlockId(ctx, data.BlockId)
	if err != nil {
		return fmt.Errorf("error finding tab for block: %w", err)
	}
	if tabId == "" {
		return fmt.Errorf("no tab found for block")
	}
	err = aincore.DeleteBlock(ctx, data.BlockId, true)
	if err != nil {
		return fmt.Errorf("error deleting block: %w", err)
	}
	aincore.QueueLayoutActionForTab(ctx, tabId, ainobj.LayoutActionData{
		ActionType: aincore.LayoutActionDataType_Remove,
		BlockId:    data.BlockId,
	})
	updates := ainobj.ContextGetUpdatesRtn(ctx)
	ainps.Broker.SendUpdateEvents(updates)
	return nil
}

func (ws *WshServer) WaitForRouteCommand(ctx context.Context, data ainshrpc.CommandWaitForRouteData) (bool, error) {
	waitCtx, cancelFn := context.WithTimeout(ctx, time.Duration(data.WaitMs)*time.Millisecond)
	defer cancelFn()
	err := ainshutil.DefaultRouter.WaitForRegister(waitCtx, data.RouteId)
	return err == nil, nil
}

func (ws *WshServer) EventRecvCommand(ctx context.Context, data ainps.WaveEvent) error {
	return nil
}

func (ws *WshServer) EventPublishCommand(ctx context.Context, data ainps.WaveEvent) error {
	rpcSource := ainshutil.GetRpcSourceFromContext(ctx)
	if rpcSource == "" {
		return fmt.Errorf("no rpc source set")
	}
	if data.Sender == "" {
		data.Sender = rpcSource
	}
	ainps.Broker.Publish(data)
	return nil
}

func (ws *WshServer) EventSubCommand(ctx context.Context, data ainps.SubscriptionRequest) error {
	rpcSource := ainshutil.GetRpcSourceFromContext(ctx)
	if rpcSource == "" {
		return fmt.Errorf("no rpc source set")
	}
	ainps.Broker.Subscribe(rpcSource, data)
	return nil
}

func (ws *WshServer) EventUnsubCommand(ctx context.Context, data string) error {
	rpcSource := ainshutil.GetRpcSourceFromContext(ctx)
	if rpcSource == "" {
		return fmt.Errorf("no rpc source set")
	}
	ainps.Broker.Unsubscribe(rpcSource, data)
	return nil
}

func (ws *WshServer) EventUnsubAllCommand(ctx context.Context) error {
	rpcSource := ainshutil.GetRpcSourceFromContext(ctx)
	if rpcSource == "" {
		return fmt.Errorf("no rpc source set")
	}
	ainps.Broker.UnsubscribeAll(rpcSource)
	return nil
}

func (ws *WshServer) EventReadHistoryCommand(ctx context.Context, data ainshrpc.CommandEventReadHistoryData) ([]*ainps.WaveEvent, error) {
	events := ainps.Broker.ReadEventHistory(data.Event, data.Scope, data.MaxItems)
	return events, nil
}

func (ws *WshServer) SetConfigCommand(ctx context.Context, data ainshrpc.MetaSettingsType) error {
	return ainconfig.SetBaseConfigValue(data.MetaMapType)
}

func (ws *WshServer) SetConnectionsConfigCommand(ctx context.Context, data ainshrpc.ConnConfigRequest) error {
	return ainconfig.SetConnectionsConfigValue(data.Host, data.MetaMapType)
}

func (ws *WshServer) GetFullConfigCommand(ctx context.Context) (ainconfig.FullConfigType, error) {
	watcher := ainconfig.GetWatcher()
	return watcher.GetFullConfig(), nil
}

func (ws *WshServer) GetAinAiModeConfigCommand(ctx context.Context) (ainconfig.AIModeConfigUpdate, error) {
	fullConfig := ainconfig.GetWatcher().GetFullConfig()
	resolvedConfigs := aiusechat.ComputeResolvedAIModeConfigs(fullConfig)
	return ainconfig.AIModeConfigUpdate{Configs: resolvedConfigs}, nil
}


func (ws *WshServer) ConnStatusCommand(ctx context.Context) ([]ainshrpc.ConnStatus, error) {
	rtn := conncontroller.GetAllConnStatus()
	return rtn, nil
}

func (ws *WshServer) WslStatusCommand(ctx context.Context) ([]ainshrpc.ConnStatus, error) {
	rtn := wslconn.GetAllConnStatus()
	return rtn, nil
}

func termCtxWithLogBlockId(ctx context.Context, logBlockId string) context.Context {
	if logBlockId == "" {
		return ctx
	}
	block, err := ainstore.DBMustGet[*ainobj.Block](ctx, logBlockId)
	if err != nil {
		return ctx
	}
	connDebug := block.Meta.GetString(ainobj.MetaKey_TermConnDebug, "")
	if connDebug == "" {
		return ctx
	}
	return blocklogger.ContextWithLogBlockId(ctx, logBlockId, connDebug == "debug")
}

func (ws *WshServer) ConnEnsureCommand(ctx context.Context, data ainshrpc.ConnExtData) error {
	// TODO: if we add proper wsh connections via aws, we'll need to handle that here
	if strings.HasPrefix(data.ConnName, "aws:") {
		profiles := awsconn.ParseProfiles()
		for profile := range profiles {
			if strings.HasPrefix(data.ConnName, profile) {
				return nil
			}
		}
	}
	ctx = genconn.ContextWithConnData(ctx, data.LogBlockId)
	ctx = termCtxWithLogBlockId(ctx, data.LogBlockId)
	if strings.HasPrefix(data.ConnName, "wsl://") {
		distroName := strings.TrimPrefix(data.ConnName, "wsl://")
		return wslconn.EnsureConnection(ctx, distroName)
	}
	return conncontroller.EnsureConnection(ctx, data.ConnName)
}

func (ws *WshServer) ConnDisconnectCommand(ctx context.Context, connName string) error {
	// TODO: if we add proper wsh connections via aws, we'll need to handle that here
	if strings.HasPrefix(connName, "aws:") {
		return nil
	}
	if conncontroller.IsLocalConnName(connName) {
		return nil
	}
	if strings.HasPrefix(connName, "wsl://") {
		distroName := strings.TrimPrefix(connName, "wsl://")
		conn := wslconn.GetWslConn(distroName)
		if conn == nil {
			return fmt.Errorf("distro not found: %s", connName)
		}
		return conn.Close()
	}
	connOpts, err := remote.ParseOpts(connName)
	if err != nil {
		return fmt.Errorf("error parsing connection name: %w", err)
	}
	conn := conncontroller.GetConn(connOpts)
	if conn == nil {
		return fmt.Errorf("connection not found: %s", connName)
	}
	return conn.Close()
}

func (ws *WshServer) ConnConnectCommand(ctx context.Context, connRequest ainshrpc.ConnRequest) error {
	// TODO: if we add proper wsh connections via aws, we'll need to handle that here
	if strings.HasPrefix(connRequest.Host, "aws:") {
		return nil
	}
	if conncontroller.IsLocalConnName(connRequest.Host) {
		return nil
	}
	ctx = genconn.ContextWithConnData(ctx, connRequest.LogBlockId)
	ctx = termCtxWithLogBlockId(ctx, connRequest.LogBlockId)
	connName := connRequest.Host
	if strings.HasPrefix(connName, "wsl://") {
		distroName := strings.TrimPrefix(connName, "wsl://")
		conn := wslconn.GetWslConn(distroName)
		if conn == nil {
			return fmt.Errorf("connection not found: %s", connName)
		}
		return conn.Connect(ctx)
	}
	connOpts, err := remote.ParseOpts(connName)
	if err != nil {
		return fmt.Errorf("error parsing connection name: %w", err)
	}
	conn := conncontroller.GetConn(connOpts)
	if conn == nil {
		return fmt.Errorf("connection not found: %s", connName)
	}
	return conn.Connect(ctx, &connRequest.Keywords)
}

func (ws *WshServer) ConnReinstallWshCommand(ctx context.Context, data ainshrpc.ConnExtData) error {
	// TODO: if we add proper wsh connections via aws, we'll need to handle that here
	if strings.HasPrefix(data.ConnName, "aws:") {
		return nil
	}
	if conncontroller.IsLocalConnName(data.ConnName) {
		return nil
	}
	ctx = genconn.ContextWithConnData(ctx, data.LogBlockId)
	ctx = termCtxWithLogBlockId(ctx, data.LogBlockId)
	connName := data.ConnName
	if strings.HasPrefix(connName, "wsl://") {
		distroName := strings.TrimPrefix(connName, "wsl://")
		conn := wslconn.GetWslConn(distroName)
		if conn == nil {
			return fmt.Errorf("connection not found: %s", connName)
		}
		return conn.InstallWsh(ctx, "")
	}
	connOpts, err := remote.ParseOpts(connName)
	if err != nil {
		return fmt.Errorf("error parsing connection name: %w", err)
	}
	conn := conncontroller.GetConn(connOpts)
	if conn == nil {
		return fmt.Errorf("connection not found: %s", connName)
	}
	return conn.InstallWsh(ctx, "")
}

func (ws *WshServer) ConnUpdateWshCommand(ctx context.Context, remoteInfo ainshrpc.RemoteInfo) (bool, error) {
	handler := ainshutil.GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return false, fmt.Errorf("could not determine handler from context")
	}
	connName := handler.GetRpcContext().Conn
	if connName == "" {
		return false, fmt.Errorf("invalid remote info: missing connection name")
	}

	log.Printf("checking wsh version for connection %s (current: %s)", connName, remoteInfo.ClientVersion)
	upToDate, _, _, err := conncontroller.IsWshVersionUpToDate(ctx, remoteInfo.ClientVersion)
	if err != nil {
		return false, fmt.Errorf("unable to compare wsh version: %w", err)
	}
	if upToDate {
		// no need to update
		log.Printf("wsh is already up to date for connection %s", connName)
		return false, nil
	}

	// todo: need to add user input code here for validation

	if strings.HasPrefix(connName, "wsl://") {
		return false, fmt.Errorf("connupdatewshcommand is not supported for wsl connections")
	}
	connOpts, err := remote.ParseOpts(connName)
	if err != nil {
		return false, fmt.Errorf("error parsing connection name: %w", err)
	}
	conn := conncontroller.GetConn(connOpts)
	if conn == nil {
		return false, fmt.Errorf("connection not found: %s", connName)
	}
	err = conn.UpdateWsh(ctx, connName, &remoteInfo)
	if err != nil {
		return false, fmt.Errorf("wsh update failed for connection %s: %w", connName, err)
	}

	// todo: need to add code for modifying configs?
	return true, nil
}

func (ws *WshServer) ConnListCommand(ctx context.Context) ([]string, error) {
	return conncontroller.GetConnectionsList()
}

func (ws *WshServer) ConnListAWSCommand(ctx context.Context) ([]string, error) {
	profilesMap := awsconn.ParseProfiles()
	return iterfn.MapKeysToSorted(profilesMap), nil
}

func (ws *WshServer) WslListCommand(ctx context.Context) ([]string, error) {
	distros, err := wsl.RegisteredDistros(ctx)
	if err != nil {
		return nil, err
	}
	var distroNames []string
	for _, distro := range distros {
		distroName := distro.Name()
		if utilfn.ContainsStr(InvalidWslDistroNames, distroName) {
			continue
		}
		distroNames = append(distroNames, distroName)
	}
	return distroNames, nil
}

func (ws *WshServer) WslDefaultDistroCommand(ctx context.Context) (string, error) {
	distro, ok, err := wsl.DefaultDistro(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to determine default distro: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("unable to determine default distro")
	}
	return distro.Name(), nil
}

/**
 * Dismisses the WshFail Command in runtime memory on the backend
 */
func (ws *WshServer) DismissWshFailCommand(ctx context.Context, connName string) error {
	if strings.HasPrefix(connName, "wsl://") {
		distroName := strings.TrimPrefix(connName, "wsl://")
		conn := wslconn.GetWslConn(distroName)
		if conn == nil {
			return fmt.Errorf("connection not found: %s", connName)
		}
		conn.ClearWshError()
		conn.FireConnChangeEvent()
		return nil
	}
	opts, err := remote.ParseOpts(connName)
	if err != nil {
		return err
	}
	conn := conncontroller.GetConn(opts)
	if conn == nil {
		return fmt.Errorf("connection %s not found", connName)
	}
	conn.ClearWshError()
	conn.FireConnChangeEvent()
	return nil
}

func (ws *WshServer) FindGitBashCommand(ctx context.Context, rescan bool) (string, error) {
	fullConfig := ainconfig.GetWatcher().GetFullConfig()
	return shellutil.FindGitBash(&fullConfig, rescan), nil
}

func (ws *WshServer) BlockInfoCommand(ctx context.Context, blockId string) (*ainshrpc.BlockInfoData, error) {
	blockData, err := ainstore.DBMustGet[*ainobj.Block](ctx, blockId)
	if err != nil {
		return nil, fmt.Errorf("error getting block: %w", err)
	}
	tabId, err := ainstore.DBFindTabForBlockId(ctx, blockId)
	if err != nil {
		return nil, fmt.Errorf("error finding tab for block: %w", err)
	}
	workspaceId, err := ainstore.DBFindWorkspaceForTabId(ctx, tabId)
	if err != nil {
		return nil, fmt.Errorf("error finding window for tab: %w", err)
	}
	fileList, err := filestore.WFS.ListFiles(ctx, blockId)
	if err != nil {
		return nil, fmt.Errorf("error listing blockfiles: %w", err)
	}
	fileInfoList := wavefileutil.WaveFileListToFileInfoList(fileList)
	return &ainshrpc.BlockInfoData{
		BlockId:     blockId,
		TabId:       tabId,
		WorkspaceId: workspaceId,
		Block:       blockData,
		Files:       fileInfoList,
	}, nil
}

func (ws *WshServer) WaveInfoCommand(ctx context.Context) (*ainshrpc.WaveInfoData, error) {
	client, err := ainstore.DBGetSingleton[*ainobj.Client](ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}
	return &ainshrpc.WaveInfoData{
		Version:   ainbase.WaveVersion,
		ClientId:  client.OID,
		BuildTime: ainbase.BuildTime,
		ConfigDir: ainbase.GetWaveConfigDir(),
		DataDir:   ainbase.GetWaveDataDir(),
	}, nil
}

// BlocksListCommand returns every block visible in the requested
// scope (current workspace by default).
func (ws *WshServer) BlocksListCommand(
	ctx context.Context,
	req ainshrpc.BlocksListRequest) ([]ainshrpc.BlocksListEntry, error) {
	var results []ainshrpc.BlocksListEntry

	// Resolve the set of workspaces to inspect
	var workspaceIDs []string
	if req.WorkspaceId != "" {
		workspaceIDs = []string{req.WorkspaceId}
	} else if req.WindowId != "" {
		win, err := aincore.GetWindow(ctx, req.WindowId)
		if err != nil {
			return nil, err
		}
		workspaceIDs = []string{win.WorkspaceId}
	} else {
		// "current" == first workspace in client focus list
		client, err := ainstore.DBGetSingleton[*ainobj.Client](ctx)
		if err != nil {
			return nil, err
		}
		if len(client.WindowIds) == 0 {
			return nil, fmt.Errorf("no active window")
		}
		win, err := aincore.GetWindow(ctx, client.WindowIds[0])
		if err != nil {
			return nil, err
		}
		workspaceIDs = []string{win.WorkspaceId}
	}

	for _, wsID := range workspaceIDs {
		wsData, err := aincore.GetWorkspace(ctx, wsID)
		if err != nil {
			return nil, err
		}

		windowId, err := ainstore.DBFindWindowForWorkspaceId(ctx, wsID)
		if err != nil {
			log.Printf("error finding window for workspace %s: %v", wsID, err)
		}

		for _, tabID := range wsData.TabIds {
			tab, err := ainstore.DBMustGet[*ainobj.Tab](ctx, tabID)
			if err != nil {
				return nil, err
			}
			for _, blkID := range tab.BlockIds {
				blk, err := ainstore.DBMustGet[*ainobj.Block](ctx, blkID)
				if err != nil {
					return nil, err
				}
				results = append(results, ainshrpc.BlocksListEntry{
					WindowId:    windowId,
					WorkspaceId: wsID,
					TabId:       tabID,
					BlockId:     blkID,
					Meta:        blk.Meta,
				})
			}
		}
	}
	return results, nil
}

func (ws *WshServer) WorkspaceListCommand(ctx context.Context) ([]ainshrpc.WorkspaceInfoData, error) {
	workspaceList, err := aincore.ListWorkspaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing workspaces: %w", err)
	}
	var rtn []ainshrpc.WorkspaceInfoData
	for _, workspaceEntry := range workspaceList {
		workspaceData, err := aincore.GetWorkspace(ctx, workspaceEntry.WorkspaceId)
		if err != nil {
			return nil, fmt.Errorf("error getting workspace: %w", err)
		}
		rtn = append(rtn, ainshrpc.WorkspaceInfoData{
			WindowId:      workspaceEntry.WindowId,
			WorkspaceData: workspaceData,
		})
	}
	return rtn, nil
}

func (ws *WshServer) ListAllAppsCommand(ctx context.Context) ([]ainshrpc.AppInfo, error) {
	return ainappstore.ListAllApps()
}

func (ws *WshServer) ListAllEditableAppsCommand(ctx context.Context) ([]ainshrpc.AppInfo, error) {
	return ainappstore.ListAllEditableApps()
}

func (ws *WshServer) ListAllAppFilesCommand(ctx context.Context, data ainshrpc.CommandListAllAppFilesData) (*ainshrpc.CommandListAllAppFilesRtnData, error) {
	if data.AppId == "" {
		return nil, fmt.Errorf("must provide an appId to ListAllAppFilesCommand")
	}
	result, err := ainappstore.ListAllAppFiles(data.AppId)
	if err != nil {
		return nil, err
	}
	entries := make([]ainshrpc.DirEntryOut, len(result.Entries))
	for i, entry := range result.Entries {
		entries[i] = ainshrpc.DirEntryOut{
			Name:         entry.Name,
			Dir:          entry.Dir,
			Symlink:      entry.Symlink,
			Size:         entry.Size,
			Mode:         entry.Mode,
			Modified:     entry.Modified,
			ModifiedTime: entry.ModifiedTime,
		}
	}
	return &ainshrpc.CommandListAllAppFilesRtnData{
		Path:         result.Path,
		AbsolutePath: result.AbsolutePath,
		ParentDir:    result.ParentDir,
		Entries:      entries,
		EntryCount:   result.EntryCount,
		TotalEntries: result.TotalEntries,
		Truncated:    result.Truncated,
	}, nil
}

func (ws *WshServer) ReadAppFileCommand(ctx context.Context, data ainshrpc.CommandReadAppFileData) (*ainshrpc.CommandReadAppFileRtnData, error) {
	if data.AppId == "" {
		return nil, fmt.Errorf("must provide an appId to ReadAppFileCommand")
	}
	fileData, err := ainappstore.ReadAppFile(data.AppId, data.FileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ainshrpc.CommandReadAppFileRtnData{
				NotFound: true,
			}, nil
		}
		return nil, fmt.Errorf("failed to read app file: %w", err)
	}
	return &ainshrpc.CommandReadAppFileRtnData{
		Data64: base64.StdEncoding.EncodeToString(fileData.Contents),
		ModTs:  fileData.ModTs,
	}, nil
}

func (ws *WshServer) WriteAppFileCommand(ctx context.Context, data ainshrpc.CommandWriteAppFileData) error {
	if data.AppId == "" {
		return fmt.Errorf("must provide an appId to WriteAppFileCommand")
	}
	contents, err := base64.StdEncoding.DecodeString(data.Data64)
	if err != nil {
		return fmt.Errorf("failed to decode data64: %w", err)
	}
	return ainappstore.WriteAppFile(data.AppId, data.FileName, contents)
}

func (ws *WshServer) WriteAppGoFileCommand(ctx context.Context, data ainshrpc.CommandWriteAppGoFileData) (*ainshrpc.CommandWriteAppGoFileRtnData, error) {
	if data.AppId == "" {
		return nil, fmt.Errorf("must provide an appId to WriteAppGoFileCommand")
	}
	contents, err := base64.StdEncoding.DecodeString(data.Data64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data64: %w", err)
	}

	formattedOutput := ainapputil.FormatGoCode(contents)

	err = ainappstore.WriteAppFile(data.AppId, "app.go", formattedOutput)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(formattedOutput)
	return &ainshrpc.CommandWriteAppGoFileRtnData{Data64: encoded}, nil
}

func (ws *WshServer) DeleteAppFileCommand(ctx context.Context, data ainshrpc.CommandDeleteAppFileData) error {
	if data.AppId == "" {
		return fmt.Errorf("must provide an appId to DeleteAppFileCommand")
	}
	return ainappstore.DeleteAppFile(data.AppId, data.FileName)
}

func (ws *WshServer) RenameAppFileCommand(ctx context.Context, data ainshrpc.CommandRenameAppFileData) error {
	if data.AppId == "" {
		return fmt.Errorf("must provide an appId to RenameAppFileCommand")
	}
	return ainappstore.RenameAppFile(data.AppId, data.FromFileName, data.ToFileName)
}

func (ws *WshServer) WriteAppSecretBindingsCommand(ctx context.Context, data ainshrpc.CommandWriteAppSecretBindingsData) error {
	if data.AppId == "" {
		return fmt.Errorf("must provide an appId to WriteAppSecretBindingsCommand")
	}
	return ainappstore.WriteAppSecretBindings(data.AppId, data.Bindings)
}

func (ws *WshServer) DeleteBuilderCommand(ctx context.Context, builderId string) error {
	if builderId == "" {
		return fmt.Errorf("must provide a builderId to DeleteBuilderCommand")
	}
	buildercontroller.DeleteController(builderId)
	return nil
}

func (ws *WshServer) StartBuilderCommand(ctx context.Context, data ainshrpc.CommandStartBuilderData) error {
	if data.BuilderId == "" {
		return fmt.Errorf("must provide a builderId to StartBuilderCommand")
	}
	bc := buildercontroller.GetOrCreateController(data.BuilderId)
	rtInfo := ainstore.GetRTInfo(ainobj.MakeORef("builder", data.BuilderId))
	if rtInfo == nil {
		return fmt.Errorf("builder rtinfo not found for builderid: %s", data.BuilderId)
	}
	appId := rtInfo.BuilderAppId
	if appId == "" {
		return fmt.Errorf("builder appid not set for builderid: %s", data.BuilderId)
	}
	return bc.Start(ctx, appId, rtInfo.BuilderEnv)
}

func (ws *WshServer) StopBuilderCommand(ctx context.Context, builderId string) error {
	if builderId == "" {
		return fmt.Errorf("must provide a builderId to StopBuilderCommand")
	}
	bc := buildercontroller.GetController(builderId)
	if bc == nil {
		return nil
	}
	return bc.Stop()
}

func (ws *WshServer) RestartBuilderAndWaitCommand(ctx context.Context, data ainshrpc.CommandRestartBuilderAndWaitData) (*ainshrpc.RestartBuilderAndWaitResult, error) {
	if data.BuilderId == "" {
		return nil, fmt.Errorf("must provide a builderId to RestartBuilderAndWaitCommand")
	}

	bc := buildercontroller.GetOrCreateController(data.BuilderId)
	rtInfo := ainstore.GetRTInfo(ainobj.MakeORef("builder", data.BuilderId))
	if rtInfo == nil {
		return nil, fmt.Errorf("builder rtinfo not found for builderid: %s", data.BuilderId)
	}

	appId := rtInfo.BuilderAppId
	if appId == "" {
		return nil, fmt.Errorf("builder appid not set for builderid: %s", data.BuilderId)
	}

	result, err := bc.RestartAndWaitForBuild(ctx, appId, rtInfo.BuilderEnv)
	if err != nil {
		return nil, err
	}

	return &ainshrpc.RestartBuilderAndWaitResult{
		Success:      result.Success,
		ErrorMessage: result.ErrorMessage,
		BuildOutput:  result.BuildOutput,
	}, nil
}

func (ws *WshServer) GetBuilderStatusCommand(ctx context.Context, builderId string) (*ainshrpc.BuilderStatusData, error) {
	if builderId == "" {
		return nil, fmt.Errorf("must provide a builderId to GetBuilderStatusCommand")
	}
	bc := buildercontroller.GetOrCreateController(builderId)
	status := bc.GetStatus()
	return &status, nil
}

func (ws *WshServer) GetBuilderOutputCommand(ctx context.Context, builderId string) ([]string, error) {
	if builderId == "" {
		return nil, fmt.Errorf("must provide a builderId to GetBuilderOutputCommand")
	}
	bc := buildercontroller.GetOrCreateController(builderId)
	return bc.GetOutput(), nil
}

func (ws *WshServer) CheckGoVersionCommand(ctx context.Context) (*ainshrpc.CommandCheckGoVersionRtnData, error) {
	watcher := ainconfig.GetWatcher()
	fullConfig := watcher.GetFullConfig()
	goPath := fullConfig.Settings.TsunamiGoPath

	result := build.CheckGoVersion(goPath)

	return &ainshrpc.CommandCheckGoVersionRtnData{
		GoStatus:    result.GoStatus,
		GoPath:      result.GoPath,
		GoVersion:   result.GoVersion,
		ErrorString: result.ErrorString,
	}, nil
}

func (ws *WshServer) PublishAppCommand(ctx context.Context, data ainshrpc.CommandPublishAppData) (*ainshrpc.CommandPublishAppRtnData, error) {
	publishedAppId, err := ainappstore.PublishDraft(data.AppId)
	if err != nil {
		return nil, fmt.Errorf("error publishing app: %w", err)
	}
	return &ainshrpc.CommandPublishAppRtnData{
		PublishedAppId: publishedAppId,
	}, nil
}

func (ws *WshServer) MakeDraftFromLocalCommand(ctx context.Context, data ainshrpc.CommandMakeDraftFromLocalData) (*ainshrpc.CommandMakeDraftFromLocalRtnData, error) {
	draftAppId, err := ainappstore.MakeDraftFromLocal(data.LocalAppId)
	if err != nil {
		return nil, fmt.Errorf("error making draft from local: %w", err)
	}
	return &ainshrpc.CommandMakeDraftFromLocalRtnData{
		DraftAppId: draftAppId,
	}, nil
}

func (ws *WshServer) RecordTEventCommand(ctx context.Context, data telemetrydata.TEvent) error {
	err := telemetry.RecordTEvent(ctx, &data)
	if err != nil {
		log.Printf("error recording telemetry event: %v", err)
	}
	return err
}

func (ws WshServer) SendTelemetryCommand(ctx context.Context) error {
	client, err := ainstore.DBGetSingleton[*ainobj.Client](ctx)
	if err != nil {
		return fmt.Errorf("getting client data for telemetry: %v", err)
	}
	return aincloud.SendAllTelemetry(client.OID)
}

func (ws *WshServer) AinAiEnableTelemetryCommand(ctx context.Context) error {
	// Enable telemetry in config
	meta := ainobj.MetaMapType{
		ainconfig.ConfigKey_TelemetryEnabled: true,
	}
	err := ainconfig.SetBaseConfigValue(meta)
	if err != nil {
		return fmt.Errorf("error setting telemetry enabled: %w", err)
	}

	// Get client for telemetry operations
	client, err := ainstore.DBGetSingleton[*ainobj.Client](ctx)
	if err != nil {
		return fmt.Errorf("getting client data for telemetry: %v", err)
	}

	// Record the telemetry event
	event := telemetrydata.MakeTEvent("waveai:enabletelemetry", telemetrydata.TEventProps{})
	err = telemetry.RecordTEvent(ctx, event)
	if err != nil {
		log.Printf("error recording waveai:enabletelemetry event: %v", err)
	}

	// Immediately send telemetry to cloud
	err = aincloud.SendAllTelemetry(client.OID)
	if err != nil {
		log.Printf("error sending telemetry after enabling: %v", err)
	}

	return nil
}


func (ws *WshServer) GetAinAiChatCommand(ctx context.Context, data ainshrpc.CommandGetWaveAIChatData) (*uctypes.UIChat, error) {
	aiChat := chatstore.DefaultChatStore.Get(data.ChatId)
	if aiChat == nil {
		return nil, nil
	}
	uiChat, err := aiusechat.ConvertAIChatToUIChat(aiChat)
	if err != nil {
		return nil, fmt.Errorf("error converting AI chat to UI chat: %w", err)
	}
	return uiChat, nil
}

func (ws *WshServer) GetAinAiRateLimitCommand(ctx context.Context) (*uctypes.RateLimitInfo, error) {
	return aiusechat.GetGlobalRateLimit(), nil
}

func (ws *WshServer) AinAiToolApproveCommand(ctx context.Context, data ainshrpc.CommandWaveAIToolApproveData) error {
	return aiusechat.UpdateToolApproval(data.ToolCallId, data.Approval)
}

func (ws *WshServer) AinAiAddContextCommand(ctx context.Context, data ainshrpc.CommandWaveAIAddContextData) error {
	// TODO: implement
	return nil
}

func (ws *WshServer) AinAiGetToolDiffCommand(ctx context.Context, data ainshrpc.CommandWaveAIGetToolDiffData) (*ainshrpc.CommandWaveAIGetToolDiffRtnData, error) {
	originalContent, modifiedContent, err := aiusechat.CreateWriteTextFileDiff(ctx, data.ChatId, data.ToolCallId)
	if err != nil {
		return nil, err
	}

	return &ainshrpc.CommandWaveAIGetToolDiffRtnData{
		OriginalContents64: base64.StdEncoding.EncodeToString(originalContent),
		ModifiedContents64: base64.StdEncoding.EncodeToString(modifiedContent),
	}, nil
}


var wshActivityRe = regexp.MustCompile(`^[a-z:#]+$`)

func (ws *WshServer) WshActivityCommand(ctx context.Context, data map[string]int) error {
	if len(data) == 0 {
		return nil
	}
	props := telemetrydata.TEventProps{}
	for key, value := range data {
		if len(key) > 20 {
			delete(data, key)
		}
		if !wshActivityRe.MatchString(key) {
			delete(data, key)
		}
		if value != 1 {
			delete(data, key)
		}
		if strings.HasSuffix(key, "#error") {
			props.WshHadError = true
		} else {
			props.WshCmd = key
		}
	}
	activityUpdate := ainshrpc.ActivityUpdate{
		WshCmds: data,
	}
	telemetry.GoUpdateActivityWrap(activityUpdate, "wsh-activity")
	telemetry.GoRecordTEventWrap(&telemetrydata.TEvent{
		Event: "wsh:run",
		Props: props,
	})
	return nil
}

func (ws *WshServer) ActivityCommand(ctx context.Context, activity ainshrpc.ActivityUpdate) error {
	telemetry.GoUpdateActivityWrap(activity, "wshrpc-activity")
	return nil
}

func (ws *WshServer) GetVarCommand(ctx context.Context, data ainshrpc.CommandVarData) (*ainshrpc.CommandVarResponseData, error) {
	_, fileData, err := filestore.WFS.ReadFile(ctx, data.ZoneId, data.FileName)
	if err == fs.ErrNotExist {
		return &ainshrpc.CommandVarResponseData{Key: data.Key, Exists: false}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading blockfile: %w", err)
	}
	envMap := envutil.EnvToMap(string(fileData))
	value, ok := envMap[data.Key]
	return &ainshrpc.CommandVarResponseData{Key: data.Key, Exists: ok, Val: value}, nil
}

func (ws *WshServer) SetVarCommand(ctx context.Context, data ainshrpc.CommandVarData) error {
	_, fileData, err := filestore.WFS.ReadFile(ctx, data.ZoneId, data.FileName)
	if err == fs.ErrNotExist {
		fileData = []byte{}
		err = filestore.WFS.MakeFile(ctx, data.ZoneId, data.FileName, nil, ainshrpc.FileOpts{})
		if err != nil {
			return fmt.Errorf("error creating blockfile: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error reading blockfile: %w", err)
	}
	envMap := envutil.EnvToMap(string(fileData))
	if data.Remove {
		delete(envMap, data.Key)
	} else {
		envMap[data.Key] = data.Val
	}
	envStr := envutil.MapToEnv(envMap)
	return filestore.WFS.WriteFile(ctx, data.ZoneId, data.FileName, []byte(envStr))
}

func (ws *WshServer) PathCommand(ctx context.Context, data ainshrpc.PathCommandData) (string, error) {
	pathType := data.PathType
	openInternal := data.Open
	openExternal := data.OpenExternal
	var path string
	switch pathType {
	case "config":
		path = ainbase.GetWaveConfigDir()
	case "data":
		path = ainbase.GetWaveDataDir()
	case "log":
		path = filepath.Join(ainbase.GetWaveDataDir(), "ainapp.log")
	}

	if openInternal && openExternal {
		return "", fmt.Errorf("open and openExternal cannot both be true")
	}

	if openInternal {
		_, err := ws.CreateBlockCommand(ctx, ainshrpc.CommandCreateBlockData{
			TabId: data.TabId,
			BlockDef: &ainobj.BlockDef{Meta: map[string]any{
				ainobj.MetaKey_View: "preview",
				ainobj.MetaKey_File: path,
			}},
			Ephemeral: true,
			Focused:   true,
		})

		if err != nil {
			return path, fmt.Errorf("error opening path: %w", err)
		}
	} else if openExternal {
		err := open.Run(path)
		if err != nil {
			return path, fmt.Errorf("error opening path: %w", err)
		}
	}
	return path, nil
}

func (ws *WshServer) FetchSuggestionsCommand(ctx context.Context, data ainshrpc.FetchSuggestionsData) (*ainshrpc.FetchSuggestionsResponse, error) {
	return suggestion.FetchSuggestions(ctx, data)
}

func (ws *WshServer) DisposeSuggestionsCommand(ctx context.Context, widgetId string) error {
	suggestion.DisposeSuggestions(ctx, widgetId)
	return nil
}

func (ws *WshServer) GetTabCommand(ctx context.Context, tabId string) (*ainobj.Tab, error) {
	tab, err := ainstore.DBGet[*ainobj.Tab](ctx, tabId)
	if err != nil {
		return nil, fmt.Errorf("error getting tab: %w", err)
	}
	return tab, nil
}

func (ws *WshServer) GetSecretsCommand(ctx context.Context, names []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, name := range names {
		value, exists, err := secretstore.GetSecret(name)
		if err != nil {
			return nil, fmt.Errorf("error getting secret %q: %w", name, err)
		}
		if exists {
			result[name] = value
		}
	}
	return result, nil
}

func (ws *WshServer) GetSecretsNamesCommand(ctx context.Context) ([]string, error) {
	names, err := secretstore.GetSecretNames()
	if err != nil {
		return nil, fmt.Errorf("error getting secret names: %w", err)
	}
	return names, nil
}

func (ws *WshServer) SetSecretsCommand(ctx context.Context, secrets map[string]*string) error {
	for name, value := range secrets {
		if value == nil {
			err := secretstore.DeleteSecret(name)
			if err != nil {
				return fmt.Errorf("error deleting secret %q: %w", name, err)
			}
		} else {
			err := secretstore.SetSecret(name, *value)
			if err != nil {
				return fmt.Errorf("error setting secret %q: %w", name, err)
			}
		}
	}
	return nil
}

func (ws *WshServer) GetSecretsLinuxStorageBackendCommand(ctx context.Context) (string, error) {
	backend, err := secretstore.GetLinuxStorageBackend()
	if err != nil {
		return "", fmt.Errorf("error getting linux storage backend: %w", err)
	}
	return backend, nil
}
