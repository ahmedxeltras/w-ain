// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

// Generated Code. DO NOT EDIT.

package wshclient

import (
	"github.com/wavetermdev/ainterm/pkg/telemetry/telemetrydata"
	"github.com/wavetermdev/ainterm/pkg/ainshutil"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainconfig"
	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainps"
	"github.com/wavetermdev/ainterm/pkg/vdom"
	"github.com/wavetermdev/ainterm/pkg/util/iochan/iochantypes"
	"github.com/wavetermdev/ainterm/pkg/aiusechat/uctypes"
)

// command "activity", wshserver.ActivityCommand
func ActivityCommand(w *ainshutil.WshRpc, data ainshrpc.ActivityUpdate, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "activity", data, opts)
	return err
}

// command "ainaiaddcontext", wshserver.AinAiAddContextCommand
func AinAiAddContextCommand(w *ainshutil.WshRpc, data ainshrpc.CommandWaveAIAddContextData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "ainaiaddcontext", data, opts)
	return err
}

// command "ainaienabletelemetry", wshserver.AinAiEnableTelemetryCommand
func AinAiEnableTelemetryCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "ainaienabletelemetry", nil, opts)
	return err
}

// command "ainaigettooldiff", wshserver.AinAiGetToolDiffCommand
func AinAiGetToolDiffCommand(w *ainshutil.WshRpc, data ainshrpc.CommandWaveAIGetToolDiffData, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandWaveAIGetToolDiffRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandWaveAIGetToolDiffRtnData](w, "ainaigettooldiff", data, opts)
	return resp, err
}

// command "ainaitoolapprove", wshserver.AinAiToolApproveCommand
func AinAiToolApproveCommand(w *ainshutil.WshRpc, data ainshrpc.CommandWaveAIToolApproveData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "ainaitoolapprove", data, opts)
	return err
}

// command "aisendmessage", wshserver.AiSendMessageCommand
func AiSendMessageCommand(w *ainshutil.WshRpc, data ainshrpc.AiMessageData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "aisendmessage", data, opts)
	return err
}

// command "authenticate", wshserver.AuthenticateCommand
func AuthenticateCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) (ainshrpc.CommandAuthenticateRtnData, error) {
	resp, err := sendRpcRequestCallHelper[ainshrpc.CommandAuthenticateRtnData](w, "authenticate", data, opts)
	return resp, err
}

// command "authenticatetoken", wshserver.AuthenticateTokenCommand
func AuthenticateTokenCommand(w *ainshutil.WshRpc, data ainshrpc.CommandAuthenticateTokenData, opts *ainshrpc.RpcOpts) (ainshrpc.CommandAuthenticateRtnData, error) {
	resp, err := sendRpcRequestCallHelper[ainshrpc.CommandAuthenticateRtnData](w, "authenticatetoken", data, opts)
	return resp, err
}

// command "authenticatetokenverify", wshserver.AuthenticateTokenVerifyCommand
func AuthenticateTokenVerifyCommand(w *ainshutil.WshRpc, data ainshrpc.CommandAuthenticateTokenData, opts *ainshrpc.RpcOpts) (ainshrpc.CommandAuthenticateRtnData, error) {
	resp, err := sendRpcRequestCallHelper[ainshrpc.CommandAuthenticateRtnData](w, "authenticatetokenverify", data, opts)
	return resp, err
}

// command "blockinfo", wshserver.BlockInfoCommand
func BlockInfoCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) (*ainshrpc.BlockInfoData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.BlockInfoData](w, "blockinfo", data, opts)
	return resp, err
}

// command "blockslist", wshserver.BlocksListCommand
func BlocksListCommand(w *ainshutil.WshRpc, data ainshrpc.BlocksListRequest, opts *ainshrpc.RpcOpts) ([]ainshrpc.BlocksListEntry, error) {
	resp, err := sendRpcRequestCallHelper[[]ainshrpc.BlocksListEntry](w, "blockslist", data, opts)
	return resp, err
}

// command "captureblockscreenshot", wshserver.CaptureBlockScreenshotCommand
func CaptureBlockScreenshotCommand(w *ainshutil.WshRpc, data ainshrpc.CommandCaptureBlockScreenshotData, opts *ainshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "captureblockscreenshot", data, opts)
	return resp, err
}

// command "checkgoversion", wshserver.CheckGoVersionCommand
func CheckGoVersionCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandCheckGoVersionRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandCheckGoVersionRtnData](w, "checkgoversion", nil, opts)
	return resp, err
}

// command "connconnect", wshserver.ConnConnectCommand
func ConnConnectCommand(w *ainshutil.WshRpc, data ainshrpc.ConnRequest, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "connconnect", data, opts)
	return err
}

// command "conndisconnect", wshserver.ConnDisconnectCommand
func ConnDisconnectCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "conndisconnect", data, opts)
	return err
}

// command "connensure", wshserver.ConnEnsureCommand
func ConnEnsureCommand(w *ainshutil.WshRpc, data ainshrpc.ConnExtData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "connensure", data, opts)
	return err
}

// command "connlist", wshserver.ConnListCommand
func ConnListCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "connlist", nil, opts)
	return resp, err
}

// command "connlistaws", wshserver.ConnListAWSCommand
func ConnListAWSCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "connlistaws", nil, opts)
	return resp, err
}

// command "connreinstallwsh", wshserver.ConnReinstallWshCommand
func ConnReinstallWshCommand(w *ainshutil.WshRpc, data ainshrpc.ConnExtData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "connreinstallwsh", data, opts)
	return err
}

// command "connstatus", wshserver.ConnStatusCommand
func ConnStatusCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) ([]ainshrpc.ConnStatus, error) {
	resp, err := sendRpcRequestCallHelper[[]ainshrpc.ConnStatus](w, "connstatus", nil, opts)
	return resp, err
}

// command "connupdatewsh", wshserver.ConnUpdateWshCommand
func ConnUpdateWshCommand(w *ainshutil.WshRpc, data ainshrpc.RemoteInfo, opts *ainshrpc.RpcOpts) (bool, error) {
	resp, err := sendRpcRequestCallHelper[bool](w, "connupdatewsh", data, opts)
	return resp, err
}

// command "controllerappendoutput", wshserver.ControllerAppendOutputCommand
func ControllerAppendOutputCommand(w *ainshutil.WshRpc, data ainshrpc.CommandControllerAppendOutputData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "controllerappendoutput", data, opts)
	return err
}

// command "controllerinput", wshserver.ControllerInputCommand
func ControllerInputCommand(w *ainshutil.WshRpc, data ainshrpc.CommandBlockInputData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "controllerinput", data, opts)
	return err
}

// command "controllerresync", wshserver.ControllerResyncCommand
func ControllerResyncCommand(w *ainshutil.WshRpc, data ainshrpc.CommandControllerResyncData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "controllerresync", data, opts)
	return err
}

// command "controllerstop", wshserver.ControllerStopCommand
func ControllerStopCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "controllerstop", data, opts)
	return err
}

// command "createblock", wshserver.CreateBlockCommand
func CreateBlockCommand(w *ainshutil.WshRpc, data ainshrpc.CommandCreateBlockData, opts *ainshrpc.RpcOpts) (ainobj.ORef, error) {
	resp, err := sendRpcRequestCallHelper[ainobj.ORef](w, "createblock", data, opts)
	return resp, err
}

// command "createsubblock", wshserver.CreateSubBlockCommand
func CreateSubBlockCommand(w *ainshutil.WshRpc, data ainshrpc.CommandCreateSubBlockData, opts *ainshrpc.RpcOpts) (ainobj.ORef, error) {
	resp, err := sendRpcRequestCallHelper[ainobj.ORef](w, "createsubblock", data, opts)
	return resp, err
}

// command "deleteappfile", wshserver.DeleteAppFileCommand
func DeleteAppFileCommand(w *ainshutil.WshRpc, data ainshrpc.CommandDeleteAppFileData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "deleteappfile", data, opts)
	return err
}

// command "deleteblock", wshserver.DeleteBlockCommand
func DeleteBlockCommand(w *ainshutil.WshRpc, data ainshrpc.CommandDeleteBlockData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "deleteblock", data, opts)
	return err
}

// command "deletebuilder", wshserver.DeleteBuilderCommand
func DeleteBuilderCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "deletebuilder", data, opts)
	return err
}

// command "deletesubblock", wshserver.DeleteSubBlockCommand
func DeleteSubBlockCommand(w *ainshutil.WshRpc, data ainshrpc.CommandDeleteBlockData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "deletesubblock", data, opts)
	return err
}

// command "dismisswshfail", wshserver.DismissWshFailCommand
func DismissWshFailCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "dismisswshfail", data, opts)
	return err
}

// command "dispose", wshserver.DisposeCommand
func DisposeCommand(w *ainshutil.WshRpc, data ainshrpc.CommandDisposeData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "dispose", data, opts)
	return err
}

// command "disposesuggestions", wshserver.DisposeSuggestionsCommand
func DisposeSuggestionsCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "disposesuggestions", data, opts)
	return err
}

// command "electrondecrypt", wshserver.ElectronDecryptCommand
func ElectronDecryptCommand(w *ainshutil.WshRpc, data ainshrpc.CommandElectronDecryptData, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandElectronDecryptRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandElectronDecryptRtnData](w, "electrondecrypt", data, opts)
	return resp, err
}

// command "electronencrypt", wshserver.ElectronEncryptCommand
func ElectronEncryptCommand(w *ainshutil.WshRpc, data ainshrpc.CommandElectronEncryptData, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandElectronEncryptRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandElectronEncryptRtnData](w, "electronencrypt", data, opts)
	return resp, err
}

// command "eventpublish", wshserver.EventPublishCommand
func EventPublishCommand(w *ainshutil.WshRpc, data ainps.WaveEvent, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "eventpublish", data, opts)
	return err
}

// command "eventreadhistory", wshserver.EventReadHistoryCommand
func EventReadHistoryCommand(w *ainshutil.WshRpc, data ainshrpc.CommandEventReadHistoryData, opts *ainshrpc.RpcOpts) ([]*ainps.WaveEvent, error) {
	resp, err := sendRpcRequestCallHelper[[]*ainps.WaveEvent](w, "eventreadhistory", data, opts)
	return resp, err
}

// command "eventrecv", wshserver.EventRecvCommand
func EventRecvCommand(w *ainshutil.WshRpc, data ainps.WaveEvent, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "eventrecv", data, opts)
	return err
}

// command "eventsub", wshserver.EventSubCommand
func EventSubCommand(w *ainshutil.WshRpc, data ainps.SubscriptionRequest, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "eventsub", data, opts)
	return err
}

// command "eventunsub", wshserver.EventUnsubCommand
func EventUnsubCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "eventunsub", data, opts)
	return err
}

// command "eventunsuball", wshserver.EventUnsubAllCommand
func EventUnsubAllCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "eventunsuball", nil, opts)
	return err
}

// command "fetchsuggestions", wshserver.FetchSuggestionsCommand
func FetchSuggestionsCommand(w *ainshutil.WshRpc, data ainshrpc.FetchSuggestionsData, opts *ainshrpc.RpcOpts) (*ainshrpc.FetchSuggestionsResponse, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.FetchSuggestionsResponse](w, "fetchsuggestions", data, opts)
	return resp, err
}

// command "fileappend", wshserver.FileAppendCommand
func FileAppendCommand(w *ainshutil.WshRpc, data ainshrpc.FileData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "fileappend", data, opts)
	return err
}

// command "fileappendijson", wshserver.FileAppendIJsonCommand
func FileAppendIJsonCommand(w *ainshutil.WshRpc, data ainshrpc.CommandAppendIJsonData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "fileappendijson", data, opts)
	return err
}

// command "filecopy", wshserver.FileCopyCommand
func FileCopyCommand(w *ainshutil.WshRpc, data ainshrpc.CommandFileCopyData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filecopy", data, opts)
	return err
}

// command "filecreate", wshserver.FileCreateCommand
func FileCreateCommand(w *ainshutil.WshRpc, data ainshrpc.FileData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filecreate", data, opts)
	return err
}

// command "filedelete", wshserver.FileDeleteCommand
func FileDeleteCommand(w *ainshutil.WshRpc, data ainshrpc.CommandDeleteFileData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filedelete", data, opts)
	return err
}

// command "fileinfo", wshserver.FileInfoCommand
func FileInfoCommand(w *ainshutil.WshRpc, data ainshrpc.FileData, opts *ainshrpc.RpcOpts) (*ainshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.FileInfo](w, "fileinfo", data, opts)
	return resp, err
}

// command "filejoin", wshserver.FileJoinCommand
func FileJoinCommand(w *ainshutil.WshRpc, data []string, opts *ainshrpc.RpcOpts) (*ainshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.FileInfo](w, "filejoin", data, opts)
	return resp, err
}

// command "filelist", wshserver.FileListCommand
func FileListCommand(w *ainshutil.WshRpc, data ainshrpc.FileListData, opts *ainshrpc.RpcOpts) ([]*ainshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[[]*ainshrpc.FileInfo](w, "filelist", data, opts)
	return resp, err
}

// command "fileliststream", wshserver.FileListStreamCommand
func FileListStreamCommand(w *ainshutil.WshRpc, data ainshrpc.FileListData, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[ainshrpc.CommandRemoteListEntriesRtnData] {
	return sendRpcRequestResponseStreamHelper[ainshrpc.CommandRemoteListEntriesRtnData](w, "fileliststream", data, opts)
}

// command "filemkdir", wshserver.FileMkdirCommand
func FileMkdirCommand(w *ainshutil.WshRpc, data ainshrpc.FileData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filemkdir", data, opts)
	return err
}

// command "filemove", wshserver.FileMoveCommand
func FileMoveCommand(w *ainshutil.WshRpc, data ainshrpc.CommandFileCopyData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filemove", data, opts)
	return err
}

// command "fileread", wshserver.FileReadCommand
func FileReadCommand(w *ainshutil.WshRpc, data ainshrpc.FileData, opts *ainshrpc.RpcOpts) (*ainshrpc.FileData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.FileData](w, "fileread", data, opts)
	return resp, err
}

// command "filereadstream", wshserver.FileReadStreamCommand
func FileReadStreamCommand(w *ainshutil.WshRpc, data ainshrpc.FileData, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[ainshrpc.FileData] {
	return sendRpcRequestResponseStreamHelper[ainshrpc.FileData](w, "filereadstream", data, opts)
}

// command "filerestorebackup", wshserver.FileRestoreBackupCommand
func FileRestoreBackupCommand(w *ainshutil.WshRpc, data ainshrpc.CommandFileRestoreBackupData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filerestorebackup", data, opts)
	return err
}

// command "filesharecapability", wshserver.FileShareCapabilityCommand
func FileShareCapabilityCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) (ainshrpc.FileShareCapability, error) {
	resp, err := sendRpcRequestCallHelper[ainshrpc.FileShareCapability](w, "filesharecapability", data, opts)
	return resp, err
}

// command "filestreamtar", wshserver.FileStreamTarCommand
func FileStreamTarCommand(w *ainshutil.WshRpc, data ainshrpc.CommandRemoteStreamTarData, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[iochantypes.Packet] {
	return sendRpcRequestResponseStreamHelper[iochantypes.Packet](w, "filestreamtar", data, opts)
}

// command "filewrite", wshserver.FileWriteCommand
func FileWriteCommand(w *ainshutil.WshRpc, data ainshrpc.FileData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filewrite", data, opts)
	return err
}

// command "findgitbash", wshserver.FindGitBashCommand
func FindGitBashCommand(w *ainshutil.WshRpc, data bool, opts *ainshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "findgitbash", data, opts)
	return resp, err
}

// command "focuswindow", wshserver.FocusWindowCommand
func FocusWindowCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "focuswindow", data, opts)
	return err
}

// command "getainaichat", wshserver.GetAinAiChatCommand
func GetAinAiChatCommand(w *ainshutil.WshRpc, data ainshrpc.CommandGetWaveAIChatData, opts *ainshrpc.RpcOpts) (*uctypes.UIChat, error) {
	resp, err := sendRpcRequestCallHelper[*uctypes.UIChat](w, "getainaichat", data, opts)
	return resp, err
}

// command "getainaimodeconfig", wshserver.GetAinAiModeConfigCommand
func GetAinAiModeConfigCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (ainconfig.AIModeConfigUpdate, error) {
	resp, err := sendRpcRequestCallHelper[ainconfig.AIModeConfigUpdate](w, "getainaimodeconfig", nil, opts)
	return resp, err
}

// command "getainairatelimit", wshserver.GetAinAiRateLimitCommand
func GetAinAiRateLimitCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (*uctypes.RateLimitInfo, error) {
	resp, err := sendRpcRequestCallHelper[*uctypes.RateLimitInfo](w, "getainairatelimit", nil, opts)
	return resp, err
}

// command "getbuilderoutput", wshserver.GetBuilderOutputCommand
func GetBuilderOutputCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "getbuilderoutput", data, opts)
	return resp, err
}

// command "getbuilderstatus", wshserver.GetBuilderStatusCommand
func GetBuilderStatusCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) (*ainshrpc.BuilderStatusData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.BuilderStatusData](w, "getbuilderstatus", data, opts)
	return resp, err
}

// command "getfullconfig", wshserver.GetFullConfigCommand
func GetFullConfigCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (ainconfig.FullConfigType, error) {
	resp, err := sendRpcRequestCallHelper[ainconfig.FullConfigType](w, "getfullconfig", nil, opts)
	return resp, err
}

// command "getjwtpublickey", wshserver.GetJwtPublicKeyCommand
func GetJwtPublicKeyCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "getjwtpublickey", nil, opts)
	return resp, err
}

// command "getmeta", wshserver.GetMetaCommand
func GetMetaCommand(w *ainshutil.WshRpc, data ainshrpc.CommandGetMetaData, opts *ainshrpc.RpcOpts) (ainobj.MetaMapType, error) {
	resp, err := sendRpcRequestCallHelper[ainobj.MetaMapType](w, "getmeta", data, opts)
	return resp, err
}

// command "getrtinfo", wshserver.GetRTInfoCommand
func GetRTInfoCommand(w *ainshutil.WshRpc, data ainshrpc.CommandGetRTInfoData, opts *ainshrpc.RpcOpts) (*ainobj.ObjRTInfo, error) {
	resp, err := sendRpcRequestCallHelper[*ainobj.ObjRTInfo](w, "getrtinfo", data, opts)
	return resp, err
}

// command "getsecrets", wshserver.GetSecretsCommand
func GetSecretsCommand(w *ainshutil.WshRpc, data []string, opts *ainshrpc.RpcOpts) (map[string]string, error) {
	resp, err := sendRpcRequestCallHelper[map[string]string](w, "getsecrets", data, opts)
	return resp, err
}

// command "getsecretslinuxstoragebackend", wshserver.GetSecretsLinuxStorageBackendCommand
func GetSecretsLinuxStorageBackendCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "getsecretslinuxstoragebackend", nil, opts)
	return resp, err
}

// command "getsecretsnames", wshserver.GetSecretsNamesCommand
func GetSecretsNamesCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "getsecretsnames", nil, opts)
	return resp, err
}

// command "gettab", wshserver.GetTabCommand
func GetTabCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) (*ainobj.Tab, error) {
	resp, err := sendRpcRequestCallHelper[*ainobj.Tab](w, "gettab", data, opts)
	return resp, err
}

// command "gettempdir", wshserver.GetTempDirCommand
func GetTempDirCommand(w *ainshutil.WshRpc, data ainshrpc.CommandGetTempDirData, opts *ainshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "gettempdir", data, opts)
	return resp, err
}

// command "getupdatechannel", wshserver.GetUpdateChannelCommand
func GetUpdateChannelCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "getupdatechannel", nil, opts)
	return resp, err
}

// command "getvar", wshserver.GetVarCommand
func GetVarCommand(w *ainshutil.WshRpc, data ainshrpc.CommandVarData, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandVarResponseData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandVarResponseData](w, "getvar", data, opts)
	return resp, err
}

// command "listallappfiles", wshserver.ListAllAppFilesCommand
func ListAllAppFilesCommand(w *ainshutil.WshRpc, data ainshrpc.CommandListAllAppFilesData, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandListAllAppFilesRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandListAllAppFilesRtnData](w, "listallappfiles", data, opts)
	return resp, err
}

// command "listallapps", wshserver.ListAllAppsCommand
func ListAllAppsCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) ([]ainshrpc.AppInfo, error) {
	resp, err := sendRpcRequestCallHelper[[]ainshrpc.AppInfo](w, "listallapps", nil, opts)
	return resp, err
}

// command "listalleditableapps", wshserver.ListAllEditableAppsCommand
func ListAllEditableAppsCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) ([]ainshrpc.AppInfo, error) {
	resp, err := sendRpcRequestCallHelper[[]ainshrpc.AppInfo](w, "listalleditableapps", nil, opts)
	return resp, err
}

// command "makedraftfromlocal", wshserver.MakeDraftFromLocalCommand
func MakeDraftFromLocalCommand(w *ainshutil.WshRpc, data ainshrpc.CommandMakeDraftFromLocalData, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandMakeDraftFromLocalRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandMakeDraftFromLocalRtnData](w, "makedraftfromlocal", data, opts)
	return resp, err
}

// command "message", wshserver.MessageCommand
func MessageCommand(w *ainshutil.WshRpc, data ainshrpc.CommandMessageData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "message", data, opts)
	return err
}

// command "networkonline", wshserver.NetworkOnlineCommand
func NetworkOnlineCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (bool, error) {
	resp, err := sendRpcRequestCallHelper[bool](w, "networkonline", nil, opts)
	return resp, err
}

// command "notify", wshserver.NotifyCommand
func NotifyCommand(w *ainshutil.WshRpc, data ainshrpc.WaveNotificationOptions, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "notify", data, opts)
	return err
}

// command "path", wshserver.PathCommand
func PathCommand(w *ainshutil.WshRpc, data ainshrpc.PathCommandData, opts *ainshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "path", data, opts)
	return resp, err
}

// command "publishapp", wshserver.PublishAppCommand
func PublishAppCommand(w *ainshutil.WshRpc, data ainshrpc.CommandPublishAppData, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandPublishAppRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandPublishAppRtnData](w, "publishapp", data, opts)
	return resp, err
}

// command "readappfile", wshserver.ReadAppFileCommand
func ReadAppFileCommand(w *ainshutil.WshRpc, data ainshrpc.CommandReadAppFileData, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandReadAppFileRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandReadAppFileRtnData](w, "readappfile", data, opts)
	return resp, err
}

// command "recordtevent", wshserver.RecordTEventCommand
func RecordTEventCommand(w *ainshutil.WshRpc, data telemetrydata.TEvent, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "recordtevent", data, opts)
	return err
}

// command "remotefilecopy", wshserver.RemoteFileCopyCommand
func RemoteFileCopyCommand(w *ainshutil.WshRpc, data ainshrpc.CommandFileCopyData, opts *ainshrpc.RpcOpts) (bool, error) {
	resp, err := sendRpcRequestCallHelper[bool](w, "remotefilecopy", data, opts)
	return resp, err
}

// command "remotefiledelete", wshserver.RemoteFileDeleteCommand
func RemoteFileDeleteCommand(w *ainshutil.WshRpc, data ainshrpc.CommandDeleteFileData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotefiledelete", data, opts)
	return err
}

// command "remotefileinfo", wshserver.RemoteFileInfoCommand
func RemoteFileInfoCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) (*ainshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.FileInfo](w, "remotefileinfo", data, opts)
	return resp, err
}

// command "remotefilejoin", wshserver.RemoteFileJoinCommand
func RemoteFileJoinCommand(w *ainshutil.WshRpc, data []string, opts *ainshrpc.RpcOpts) (*ainshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.FileInfo](w, "remotefilejoin", data, opts)
	return resp, err
}

// command "remotefilemove", wshserver.RemoteFileMoveCommand
func RemoteFileMoveCommand(w *ainshutil.WshRpc, data ainshrpc.CommandFileCopyData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotefilemove", data, opts)
	return err
}

// command "remotefiletouch", wshserver.RemoteFileTouchCommand
func RemoteFileTouchCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotefiletouch", data, opts)
	return err
}

// command "remotegetinfo", wshserver.RemoteGetInfoCommand
func RemoteGetInfoCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (ainshrpc.RemoteInfo, error) {
	resp, err := sendRpcRequestCallHelper[ainshrpc.RemoteInfo](w, "remotegetinfo", nil, opts)
	return resp, err
}

// command "remoteinstallrcfiles", wshserver.RemoteInstallRcFilesCommand
func RemoteInstallRcFilesCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remoteinstallrcfiles", nil, opts)
	return err
}

// command "remotelistentries", wshserver.RemoteListEntriesCommand
func RemoteListEntriesCommand(w *ainshutil.WshRpc, data ainshrpc.CommandRemoteListEntriesData, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[ainshrpc.CommandRemoteListEntriesRtnData] {
	return sendRpcRequestResponseStreamHelper[ainshrpc.CommandRemoteListEntriesRtnData](w, "remotelistentries", data, opts)
}

// command "remotemkdir", wshserver.RemoteMkdirCommand
func RemoteMkdirCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotemkdir", data, opts)
	return err
}

// command "remotestreamcpudata", wshserver.RemoteStreamCpuDataCommand
func RemoteStreamCpuDataCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[ainshrpc.TimeSeriesData] {
	return sendRpcRequestResponseStreamHelper[ainshrpc.TimeSeriesData](w, "remotestreamcpudata", nil, opts)
}

// command "remotestreamfile", wshserver.RemoteStreamFileCommand
func RemoteStreamFileCommand(w *ainshutil.WshRpc, data ainshrpc.CommandRemoteStreamFileData, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[ainshrpc.FileData] {
	return sendRpcRequestResponseStreamHelper[ainshrpc.FileData](w, "remotestreamfile", data, opts)
}

// command "remotetarstream", wshserver.RemoteTarStreamCommand
func RemoteTarStreamCommand(w *ainshutil.WshRpc, data ainshrpc.CommandRemoteStreamTarData, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[iochantypes.Packet] {
	return sendRpcRequestResponseStreamHelper[iochantypes.Packet](w, "remotetarstream", data, opts)
}

// command "remotewritefile", wshserver.RemoteWriteFileCommand
func RemoteWriteFileCommand(w *ainshutil.WshRpc, data ainshrpc.FileData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotewritefile", data, opts)
	return err
}

// command "renameappfile", wshserver.RenameAppFileCommand
func RenameAppFileCommand(w *ainshutil.WshRpc, data ainshrpc.CommandRenameAppFileData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "renameappfile", data, opts)
	return err
}

// command "resolveids", wshserver.ResolveIdsCommand
func ResolveIdsCommand(w *ainshutil.WshRpc, data ainshrpc.CommandResolveIdsData, opts *ainshrpc.RpcOpts) (ainshrpc.CommandResolveIdsRtnData, error) {
	resp, err := sendRpcRequestCallHelper[ainshrpc.CommandResolveIdsRtnData](w, "resolveids", data, opts)
	return resp, err
}

// command "restartbuilderandwait", wshserver.RestartBuilderAndWaitCommand
func RestartBuilderAndWaitCommand(w *ainshutil.WshRpc, data ainshrpc.CommandRestartBuilderAndWaitData, opts *ainshrpc.RpcOpts) (*ainshrpc.RestartBuilderAndWaitResult, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.RestartBuilderAndWaitResult](w, "restartbuilderandwait", data, opts)
	return resp, err
}

// command "routeannounce", wshserver.RouteAnnounceCommand
func RouteAnnounceCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "routeannounce", nil, opts)
	return err
}

// command "routeunannounce", wshserver.RouteUnannounceCommand
func RouteUnannounceCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "routeunannounce", nil, opts)
	return err
}

// command "sendtelemetry", wshserver.SendTelemetryCommand
func SendTelemetryCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "sendtelemetry", nil, opts)
	return err
}

// command "setconfig", wshserver.SetConfigCommand
func SetConfigCommand(w *ainshutil.WshRpc, data ainshrpc.MetaSettingsType, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setconfig", data, opts)
	return err
}

// command "setconnectionsconfig", wshserver.SetConnectionsConfigCommand
func SetConnectionsConfigCommand(w *ainshutil.WshRpc, data ainshrpc.ConnConfigRequest, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setconnectionsconfig", data, opts)
	return err
}

// command "setmeta", wshserver.SetMetaCommand
func SetMetaCommand(w *ainshutil.WshRpc, data ainshrpc.CommandSetMetaData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setmeta", data, opts)
	return err
}

// command "setpeerinfo", wshserver.SetPeerInfoCommand
func SetPeerInfoCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setpeerinfo", data, opts)
	return err
}

// command "setrtinfo", wshserver.SetRTInfoCommand
func SetRTInfoCommand(w *ainshutil.WshRpc, data ainshrpc.CommandSetRTInfoData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setrtinfo", data, opts)
	return err
}

// command "setsecrets", wshserver.SetSecretsCommand
func SetSecretsCommand(w *ainshutil.WshRpc, data map[string]*string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setsecrets", data, opts)
	return err
}

// command "setvar", wshserver.SetVarCommand
func SetVarCommand(w *ainshutil.WshRpc, data ainshrpc.CommandVarData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setvar", data, opts)
	return err
}

// command "startbuilder", wshserver.StartBuilderCommand
func StartBuilderCommand(w *ainshutil.WshRpc, data ainshrpc.CommandStartBuilderData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "startbuilder", data, opts)
	return err
}

// command "stopbuilder", wshserver.StopBuilderCommand
func StopBuilderCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "stopbuilder", data, opts)
	return err
}

// command "streamainai", wshserver.StreamAinAiCommand
func StreamAinAiCommand(w *ainshutil.WshRpc, data ainshrpc.WaveAIStreamRequest, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[ainshrpc.WaveAIPacketType] {
	return sendRpcRequestResponseStreamHelper[ainshrpc.WaveAIPacketType](w, "streamainai", data, opts)
}

// command "streamcpudata", wshserver.StreamCpuDataCommand
func StreamCpuDataCommand(w *ainshutil.WshRpc, data ainshrpc.CpuDataRequest, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[ainshrpc.TimeSeriesData] {
	return sendRpcRequestResponseStreamHelper[ainshrpc.TimeSeriesData](w, "streamcpudata", data, opts)
}

// command "streamdata", wshserver.StreamDataCommand
func StreamDataCommand(w *ainshutil.WshRpc, data ainshrpc.CommandStreamData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "streamdata", data, opts)
	return err
}

// command "streamdataack", wshserver.StreamDataAckCommand
func StreamDataAckCommand(w *ainshutil.WshRpc, data ainshrpc.CommandStreamAckData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "streamdataack", data, opts)
	return err
}

// command "streamtest", wshserver.StreamTestCommand
func StreamTestCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[int] {
	return sendRpcRequestResponseStreamHelper[int](w, "streamtest", nil, opts)
}

// command "termgetscrollbacklines", wshserver.TermGetScrollbackLinesCommand
func TermGetScrollbackLinesCommand(w *ainshutil.WshRpc, data ainshrpc.CommandTermGetScrollbackLinesData, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandTermGetScrollbackLinesRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandTermGetScrollbackLinesRtnData](w, "termgetscrollbacklines", data, opts)
	return resp, err
}

// command "test", wshserver.TestCommand
func TestCommand(w *ainshutil.WshRpc, data string, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "test", data, opts)
	return err
}

// command "vdomasyncinitiation", wshserver.VDomAsyncInitiationCommand
func VDomAsyncInitiationCommand(w *ainshutil.WshRpc, data vdom.VDomAsyncInitiationRequest, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "vdomasyncinitiation", data, opts)
	return err
}

// command "vdomcreatecontext", wshserver.VDomCreateContextCommand
func VDomCreateContextCommand(w *ainshutil.WshRpc, data vdom.VDomCreateContext, opts *ainshrpc.RpcOpts) (*ainobj.ORef, error) {
	resp, err := sendRpcRequestCallHelper[*ainobj.ORef](w, "vdomcreatecontext", data, opts)
	return resp, err
}

// command "vdomrender", wshserver.VDomRenderCommand
func VDomRenderCommand(w *ainshutil.WshRpc, data vdom.VDomFrontendUpdate, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[*vdom.VDomBackendUpdate] {
	return sendRpcRequestResponseStreamHelper[*vdom.VDomBackendUpdate](w, "vdomrender", data, opts)
}

// command "vdomurlrequest", wshserver.VDomUrlRequestCommand
func VDomUrlRequestCommand(w *ainshutil.WshRpc, data ainshrpc.VDomUrlRequestData, opts *ainshrpc.RpcOpts) chan ainshrpc.RespOrErrorUnion[ainshrpc.VDomUrlRequestResponse] {
	return sendRpcRequestResponseStreamHelper[ainshrpc.VDomUrlRequestResponse](w, "vdomurlrequest", data, opts)
}

// command "waitforroute", wshserver.WaitForRouteCommand
func WaitForRouteCommand(w *ainshutil.WshRpc, data ainshrpc.CommandWaitForRouteData, opts *ainshrpc.RpcOpts) (bool, error) {
	resp, err := sendRpcRequestCallHelper[bool](w, "waitforroute", data, opts)
	return resp, err
}

// command "waveinfo", wshserver.WaveInfoCommand
func WaveInfoCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (*ainshrpc.WaveInfoData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.WaveInfoData](w, "waveinfo", nil, opts)
	return resp, err
}

// command "webselector", wshserver.WebSelectorCommand
func WebSelectorCommand(w *ainshutil.WshRpc, data ainshrpc.CommandWebSelectorData, opts *ainshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "webselector", data, opts)
	return resp, err
}

// command "workspacelist", wshserver.WorkspaceListCommand
func WorkspaceListCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) ([]ainshrpc.WorkspaceInfoData, error) {
	resp, err := sendRpcRequestCallHelper[[]ainshrpc.WorkspaceInfoData](w, "workspacelist", nil, opts)
	return resp, err
}

// command "writeappfile", wshserver.WriteAppFileCommand
func WriteAppFileCommand(w *ainshutil.WshRpc, data ainshrpc.CommandWriteAppFileData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "writeappfile", data, opts)
	return err
}

// command "writeappgofile", wshserver.WriteAppGoFileCommand
func WriteAppGoFileCommand(w *ainshutil.WshRpc, data ainshrpc.CommandWriteAppGoFileData, opts *ainshrpc.RpcOpts) (*ainshrpc.CommandWriteAppGoFileRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*ainshrpc.CommandWriteAppGoFileRtnData](w, "writeappgofile", data, opts)
	return resp, err
}

// command "writeappsecretbindings", wshserver.WriteAppSecretBindingsCommand
func WriteAppSecretBindingsCommand(w *ainshutil.WshRpc, data ainshrpc.CommandWriteAppSecretBindingsData, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "writeappsecretbindings", data, opts)
	return err
}

// command "writetempfile", wshserver.WriteTempFileCommand
func WriteTempFileCommand(w *ainshutil.WshRpc, data ainshrpc.CommandWriteTempFileData, opts *ainshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "writetempfile", data, opts)
	return resp, err
}

// command "wshactivity", wshserver.WshActivityCommand
func WshActivityCommand(w *ainshutil.WshRpc, data map[string]int, opts *ainshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "wshactivity", data, opts)
	return err
}

// command "wsldefaultdistro", wshserver.WslDefaultDistroCommand
func WslDefaultDistroCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "wsldefaultdistro", nil, opts)
	return resp, err
}

// command "wsllist", wshserver.WslListCommand
func WslListCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "wsllist", nil, opts)
	return resp, err
}

// command "wslstatus", wshserver.WslStatusCommand
func WslStatusCommand(w *ainshutil.WshRpc, opts *ainshrpc.RpcOpts) ([]ainshrpc.ConnStatus, error) {
	resp, err := sendRpcRequestCallHelper[[]ainshrpc.ConnStatus](w, "wslstatus", nil, opts)
	return resp, err
}


