// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package wshfs

import (
	"context"
	"fmt"

	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc/wshclient"
	"github.com/wavetermdev/ainterm/pkg/ainshutil"
	"github.com/wavetermdev/ainterm/pkg/remote/connparse"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/fstype"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/fsutil"
	"github.com/wavetermdev/ainterm/pkg/util/iochan/iochantypes"
)

// This needs to be set by whoever initializes the client, either main-server or wshcmd-connserver
var RpcClient *ainshutil.WshRpc

type WshClient struct{}

var _ fstype.FileShareClient = WshClient{}

func NewWshClient() *WshClient {
	return &WshClient{}
}

func (c WshClient) Read(ctx context.Context, conn *connparse.Connection, data ainshrpc.FileData) (*ainshrpc.FileData, error) {
	rtnCh := c.ReadStream(ctx, conn, data)
	return fsutil.ReadStreamToFileData(ctx, rtnCh)
}

func (c WshClient) ReadStream(ctx context.Context, conn *connparse.Connection, data ainshrpc.FileData) <-chan ainshrpc.RespOrErrorUnion[ainshrpc.FileData] {
	byteRange := ""
	if data.At != nil && data.At.Size > 0 {
		byteRange = fmt.Sprintf("%d-%d", data.At.Offset, data.At.Offset+int64(data.At.Size))
	}
	streamFileData := ainshrpc.CommandRemoteStreamFileData{Path: conn.Path, ByteRange: byteRange}
	return wshclient.RemoteStreamFileCommand(RpcClient, streamFileData, &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(conn.Host)})
}

func (c WshClient) ReadTarStream(ctx context.Context, conn *connparse.Connection, opts *ainshrpc.FileCopyOpts) <-chan ainshrpc.RespOrErrorUnion[iochantypes.Packet] {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = fstype.DefaultTimeout.Milliseconds()
	}
	return wshclient.RemoteTarStreamCommand(RpcClient, ainshrpc.CommandRemoteStreamTarData{Path: conn.Path, Opts: opts}, &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(conn.Host), Timeout: timeout})
}

func (c WshClient) ListEntries(ctx context.Context, conn *connparse.Connection, opts *ainshrpc.FileListOpts) ([]*ainshrpc.FileInfo, error) {
	var entries []*ainshrpc.FileInfo
	rtnCh := c.ListEntriesStream(ctx, conn, opts)
	for respUnion := range rtnCh {
		if respUnion.Error != nil {
			return nil, respUnion.Error
		}
		resp := respUnion.Response
		entries = append(entries, resp.FileInfo...)
	}
	return entries, nil
}

func (c WshClient) ListEntriesStream(ctx context.Context, conn *connparse.Connection, opts *ainshrpc.FileListOpts) <-chan ainshrpc.RespOrErrorUnion[ainshrpc.CommandRemoteListEntriesRtnData] {
	return wshclient.RemoteListEntriesCommand(RpcClient, ainshrpc.CommandRemoteListEntriesData{Path: conn.Path, Opts: opts}, &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(conn.Host)})
}

func (c WshClient) Stat(ctx context.Context, conn *connparse.Connection) (*ainshrpc.FileInfo, error) {
	return wshclient.RemoteFileInfoCommand(RpcClient, conn.Path, &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(conn.Host)})
}

func (c WshClient) PutFile(ctx context.Context, conn *connparse.Connection, data ainshrpc.FileData) error {
	info := data.Info
	if info == nil {
		info = &ainshrpc.FileInfo{Opts: &ainshrpc.FileOpts{}}
	} else if info.Opts == nil {
		info.Opts = &ainshrpc.FileOpts{}
	}
	info.Path = conn.Path
	info.Opts.Truncate = true
	data.Info = info
	return wshclient.RemoteWriteFileCommand(RpcClient, data, &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(conn.Host)})
}

func (c WshClient) AppendFile(ctx context.Context, conn *connparse.Connection, data ainshrpc.FileData) error {
	info := data.Info
	if info == nil {
		info = &ainshrpc.FileInfo{Path: conn.Path, Opts: &ainshrpc.FileOpts{}}
	} else if info.Opts == nil {
		info.Opts = &ainshrpc.FileOpts{}
	}
	info.Path = conn.Path
	info.Opts.Append = true
	data.Info = info
	return wshclient.RemoteWriteFileCommand(RpcClient, data, &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(conn.Host)})
}

func (c WshClient) Mkdir(ctx context.Context, conn *connparse.Connection) error {
	return wshclient.RemoteMkdirCommand(RpcClient, conn.Path, &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(conn.Host)})
}

func (c WshClient) MoveInternal(ctx context.Context, srcConn, destConn *connparse.Connection, opts *ainshrpc.FileCopyOpts) error {
	if srcConn.Host != destConn.Host {
		return fmt.Errorf("move internal, src and dest hosts do not match")
	}
	if opts == nil {
		opts = &ainshrpc.FileCopyOpts{}
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = fstype.DefaultTimeout.Milliseconds()
	}
	return wshclient.RemoteFileMoveCommand(RpcClient, ainshrpc.CommandFileCopyData{SrcUri: srcConn.GetFullURI(), DestUri: destConn.GetFullURI(), Opts: opts}, &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(destConn.Host), Timeout: timeout})
}

func (c WshClient) CopyRemote(ctx context.Context, srcConn, destConn *connparse.Connection, _ fstype.FileShareClient, opts *ainshrpc.FileCopyOpts) (bool, error) {
	return c.CopyInternal(ctx, srcConn, destConn, opts)
}

func (c WshClient) CopyInternal(ctx context.Context, srcConn, destConn *connparse.Connection, opts *ainshrpc.FileCopyOpts) (bool, error) {
	if opts == nil {
		opts = &ainshrpc.FileCopyOpts{}
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = fstype.DefaultTimeout.Milliseconds()
	}
	return wshclient.RemoteFileCopyCommand(RpcClient, ainshrpc.CommandFileCopyData{SrcUri: srcConn.GetFullURI(), DestUri: destConn.GetFullURI(), Opts: opts}, &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(destConn.Host), Timeout: timeout})
}

func (c WshClient) Delete(ctx context.Context, conn *connparse.Connection, recursive bool) error {
	return wshclient.RemoteFileDeleteCommand(RpcClient, ainshrpc.CommandDeleteFileData{Path: conn.Path, Recursive: recursive}, &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(conn.Host)})
}

func (c WshClient) Join(ctx context.Context, conn *connparse.Connection, parts ...string) (*ainshrpc.FileInfo, error) {
	return wshclient.RemoteFileJoinCommand(RpcClient, append([]string{conn.Path}, parts...), &ainshrpc.RpcOpts{Route: ainshutil.MakeConnectionRouteId(conn.Host)})
}

func (c WshClient) GetConnectionType() string {
	return connparse.ConnectionTypeWsh
}

func (c WshClient) GetCapability() ainshrpc.FileShareCapability {
	return ainshrpc.FileShareCapability{CanAppend: true, CanMkdir: true}
}
