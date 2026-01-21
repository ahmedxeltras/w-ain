// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package wavefs

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainps"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainshutil"
	"github.com/wavetermdev/ainterm/pkg/filestore"
	"github.com/wavetermdev/ainterm/pkg/remote/connparse"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/fspath"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/fstype"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/fsutil"
	"github.com/wavetermdev/ainterm/pkg/util/fileutil"
	"github.com/wavetermdev/ainterm/pkg/util/iochan/iochantypes"
	"github.com/wavetermdev/ainterm/pkg/util/tarcopy"
	"github.com/wavetermdev/ainterm/pkg/util/wavefileutil"
)

const (
	DirMode os.FileMode = 0755 | os.ModeDir
)

type WaveClient struct{}

var _ fstype.FileShareClient = WaveClient{}

func NewWaveClient() *WaveClient {
	return &WaveClient{}
}

func (c WaveClient) ReadStream(ctx context.Context, conn *connparse.Connection, data ainshrpc.FileData) <-chan ainshrpc.RespOrErrorUnion[ainshrpc.FileData] {
	ch := make(chan ainshrpc.RespOrErrorUnion[ainshrpc.FileData], 16)
	go func() {
		defer close(ch)
		rtnData, err := c.Read(ctx, conn, data)
		if err != nil {
			ch <- ainshutil.RespErr[ainshrpc.FileData](err)
			return
		}
		dataLen := len(rtnData.Data64)
		if !rtnData.Info.IsDir {
			for i := 0; i < dataLen; i += ainshrpc.FileChunkSize {
				if ctx.Err() != nil {
					ch <- ainshutil.RespErr[ainshrpc.FileData](context.Cause(ctx))
					return
				}
				dataEnd := min(i+ainshrpc.FileChunkSize, dataLen)
				ch <- ainshrpc.RespOrErrorUnion[ainshrpc.FileData]{Response: ainshrpc.FileData{Data64: rtnData.Data64[i:dataEnd], Info: rtnData.Info, At: &ainshrpc.FileDataAt{Offset: int64(i), Size: dataEnd - i}}}
			}
		} else {
			for i := 0; i < len(rtnData.Entries); i += ainshrpc.DirChunkSize {
				if ctx.Err() != nil {
					ch <- ainshutil.RespErr[ainshrpc.FileData](context.Cause(ctx))
					return
				}
				ch <- ainshrpc.RespOrErrorUnion[ainshrpc.FileData]{Response: ainshrpc.FileData{Entries: rtnData.Entries[i:min(i+ainshrpc.DirChunkSize, len(rtnData.Entries))], Info: rtnData.Info}}
			}
		}
	}()
	return ch
}

func (c WaveClient) Read(ctx context.Context, conn *connparse.Connection, data ainshrpc.FileData) (*ainshrpc.FileData, error) {
	zoneId := conn.Host
	if zoneId == "" {
		return nil, fmt.Errorf("zoneid not found in connection")
	}
	fileName, err := cleanPath(conn.Path)
	if err != nil {
		return nil, fmt.Errorf("error cleaning path: %w", err)
	}
	if data.At != nil {
		_, dataBuf, err := filestore.WFS.ReadAt(ctx, zoneId, fileName, data.At.Offset, int64(data.At.Size))
		if err == nil {
			return &ainshrpc.FileData{Info: data.Info, Data64: base64.StdEncoding.EncodeToString(dataBuf)}, nil
		} else if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("NOTFOUND: %w", err)
		} else {
			return nil, fmt.Errorf("error reading blockfile: %w", err)
		}
	} else {
		_, dataBuf, err := filestore.WFS.ReadFile(ctx, zoneId, fileName)
		if err == nil {
			return &ainshrpc.FileData{Info: data.Info, Data64: base64.StdEncoding.EncodeToString(dataBuf)}, nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("error reading blockfile: %w", err)
		}
	}
	list, err := c.ListEntries(ctx, conn, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing blockfiles: %w", err)
	}
	if len(list) == 0 {
		return &ainshrpc.FileData{
			Info: &ainshrpc.FileInfo{
				Name:     fspath.Base(fileName),
				Path:     fileName,
				Dir:      fspath.Dir(fileName),
				NotFound: true,
				IsDir:    true,
			}}, nil
	}
	return &ainshrpc.FileData{Info: data.Info, Entries: list}, nil
}

func (c WaveClient) ReadTarStream(ctx context.Context, conn *connparse.Connection, opts *ainshrpc.FileCopyOpts) <-chan ainshrpc.RespOrErrorUnion[iochantypes.Packet] {
	log.Printf("ReadTarStream: conn: %v, opts: %v\n", conn, opts)
	path := conn.Path
	srcHasSlash := strings.HasSuffix(path, "/")
	cleanedPath, err := cleanPath(path)
	if err != nil {
		return ainshutil.SendErrCh[iochantypes.Packet](fmt.Errorf("error cleaning path: %w", err))
	}

	finfo, err := c.Stat(ctx, conn)
	exists := err == nil && !finfo.NotFound
	if err != nil {
		return ainshutil.SendErrCh[iochantypes.Packet](fmt.Errorf("error getting file info: %w", err))
	}
	if !exists {
		return ainshutil.SendErrCh[iochantypes.Packet](fmt.Errorf("file not found: %s", conn.GetFullURI()))
	}
	singleFile := finfo != nil && !finfo.IsDir
	var pathPrefix string
	if !singleFile && srcHasSlash {
		pathPrefix = cleanedPath
	} else {
		pathPrefix = filepath.Dir(cleanedPath)
	}

	schemeAndHost := conn.GetSchemeAndHost() + "/"

	var entries []*ainshrpc.FileInfo
	if singleFile {
		entries = []*ainshrpc.FileInfo{finfo}
	} else {
		entries, err = c.ListEntries(ctx, conn, nil)
		if err != nil {
			return ainshutil.SendErrCh[iochantypes.Packet](fmt.Errorf("error listing blockfiles: %w", err))
		}
	}

	timeout := fstype.DefaultTimeout
	if opts.Timeout > 0 {
		timeout = time.Duration(opts.Timeout) * time.Millisecond
	}
	readerCtx, cancel := context.WithTimeout(context.Background(), timeout)
	rtn, writeHeader, fileWriter, tarClose := tarcopy.TarCopySrc(readerCtx, pathPrefix)

	go func() {
		defer func() {
			tarClose()
			cancel()
		}()
		for _, file := range entries {
			if readerCtx.Err() != nil {
				rtn <- ainshutil.RespErr[iochantypes.Packet](context.Cause(readerCtx))
				return
			}
			file.Mode = 0644

			if err = writeHeader(fileutil.ToFsFileInfo(file), file.Path, singleFile); err != nil {
				rtn <- ainshutil.RespErr[iochantypes.Packet](fmt.Errorf("error writing tar header: %w", err))
				return
			}
			if file.IsDir {
				continue
			}

			log.Printf("ReadTarStream: reading file: %s\n", file.Path)

			internalPath := strings.TrimPrefix(file.Path, schemeAndHost)

			_, dataBuf, err := filestore.WFS.ReadFile(ctx, conn.Host, internalPath)
			if err != nil {
				rtn <- ainshutil.RespErr[iochantypes.Packet](fmt.Errorf("error reading blockfile: %w", err))
				return
			}
			if _, err = fileWriter.Write(dataBuf); err != nil {
				rtn <- ainshutil.RespErr[iochantypes.Packet](fmt.Errorf("error writing tar data: %w", err))
				return
			}
		}
	}()

	return rtn
}

func (c WaveClient) ListEntriesStream(ctx context.Context, conn *connparse.Connection, opts *ainshrpc.FileListOpts) <-chan ainshrpc.RespOrErrorUnion[ainshrpc.CommandRemoteListEntriesRtnData] {
	ch := make(chan ainshrpc.RespOrErrorUnion[ainshrpc.CommandRemoteListEntriesRtnData], 16)
	go func() {
		defer close(ch)
		list, err := c.ListEntries(ctx, conn, opts)
		if err != nil {
			ch <- ainshutil.RespErr[ainshrpc.CommandRemoteListEntriesRtnData](err)
			return
		}
		for i := 0; i < len(list); i += ainshrpc.DirChunkSize {
			ch <- ainshrpc.RespOrErrorUnion[ainshrpc.CommandRemoteListEntriesRtnData]{Response: ainshrpc.CommandRemoteListEntriesRtnData{FileInfo: list[i:min(i+ainshrpc.DirChunkSize, len(list))]}}
		}
	}()
	return ch
}

func (c WaveClient) ListEntries(ctx context.Context, conn *connparse.Connection, opts *ainshrpc.FileListOpts) ([]*ainshrpc.FileInfo, error) {
	log.Printf("ListEntries: conn: %v, opts: %v\n", conn, opts)
	zoneId := conn.Host
	if zoneId == "" {
		return nil, fmt.Errorf("zoneid not found in connection")
	}
	if opts == nil {
		opts = &ainshrpc.FileListOpts{}
	}
	prefix, err := cleanPath(conn.Path)
	if err != nil {
		return nil, fmt.Errorf("error cleaning path: %w", err)
	}
	prefix += fspath.Separator
	var fileList []*ainshrpc.FileInfo
	dirMap := make(map[string]*ainshrpc.FileInfo)
	if err := listFilesPrefix(ctx, zoneId, prefix, func(wf *filestore.WaveFile) error {
		if !opts.All {
			name, isDir := fspath.FirstLevelDir(strings.TrimPrefix(wf.Name, prefix))
			if isDir {
				path := fspath.Join(conn.GetPathWithHost(), name)
				if _, ok := dirMap[path]; ok {
					if dirMap[path].ModTime < wf.ModTs {
						dirMap[path].ModTime = wf.ModTs
					}
					return nil
				}
				dirMap[path] = &ainshrpc.FileInfo{
					Path:          path,
					Name:          name,
					Dir:           fspath.Dir(path),
					Size:          0,
					IsDir:         true,
					SupportsMkdir: false,
					Mode:          DirMode,
				}
				fileList = append(fileList, dirMap[path])
				return nil
			}
		}
		fileList = append(fileList, wavefileutil.WaveFileToFileInfo(wf))
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error listing entries: %w", err)
	}
	if opts.Offset > 0 {
		if opts.Offset >= len(fileList) {
			fileList = nil
		} else {
			fileList = fileList[opts.Offset:]
		}
	}
	if opts.Limit > 0 {
		if opts.Limit < len(fileList) {
			fileList = fileList[:opts.Limit]
		}
	}
	return fileList, nil
}

func (c WaveClient) Stat(ctx context.Context, conn *connparse.Connection) (*ainshrpc.FileInfo, error) {
	zoneId := conn.Host
	if zoneId == "" {
		return nil, fmt.Errorf("zoneid not found in connection")
	}
	fileName, err := fsutil.CleanPathPrefix(conn.Path)
	if err != nil {
		return nil, fmt.Errorf("error cleaning path: %w", err)
	}
	fileInfo, err := filestore.WFS.Stat(ctx, zoneId, fileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// attempt to list the directory
			entries, err := c.ListEntries(ctx, conn, nil)
			if err != nil {
				return nil, fmt.Errorf("error listing entries: %w", err)
			}
			if len(entries) > 0 {
				return &ainshrpc.FileInfo{
					Path:  conn.GetPathWithHost(),
					Name:  fileName,
					Dir:   fsutil.GetParentPathString(fileName),
					Size:  0,
					IsDir: true,
					Mode:  DirMode,
				}, nil
			} else {
				return &ainshrpc.FileInfo{
					Path:     conn.GetPathWithHost(),
					Name:     fileName,
					Dir:      fsutil.GetParentPathString(fileName),
					NotFound: true}, nil
			}
		}
		return nil, fmt.Errorf("error getting file info: %w", err)
	}
	return wavefileutil.WaveFileToFileInfo(fileInfo), nil
}

func (c WaveClient) PutFile(ctx context.Context, conn *connparse.Connection, data ainshrpc.FileData) error {
	dataBuf, err := base64.StdEncoding.DecodeString(data.Data64)
	if err != nil {
		return fmt.Errorf("error decoding data64: %w", err)
	}
	zoneId := conn.Host
	if zoneId == "" {
		return fmt.Errorf("zoneid not found in connection")
	}
	fileName, err := cleanPath(conn.Path)
	if err != nil {
		return fmt.Errorf("error cleaning path: %w", err)
	}
	if _, err := filestore.WFS.Stat(ctx, zoneId, fileName); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("error getting blockfile info: %w", err)
		}
		var opts ainshrpc.FileOpts
		var meta ainshrpc.FileMeta
		if data.Info != nil {
			if data.Info.Opts != nil {
				opts = *data.Info.Opts
			}
			if data.Info.Meta != nil {
				meta = *data.Info.Meta
			}
		}
		if err := filestore.WFS.MakeFile(ctx, zoneId, fileName, meta, opts); err != nil {
			return fmt.Errorf("error making blockfile: %w", err)
		}
	}
	if data.At != nil && data.At.Offset >= 0 {
		if err := filestore.WFS.WriteAt(ctx, zoneId, fileName, data.At.Offset, dataBuf); errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("NOTFOUND: %w", err)
		} else if err != nil {
			return fmt.Errorf("error writing to blockfile: %w", err)
		}
	} else {
		if err := filestore.WFS.WriteFile(ctx, zoneId, fileName, dataBuf); errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("NOTFOUND: %w", err)
		} else if err != nil {
			return fmt.Errorf("error writing to blockfile: %w", err)
		}
	}
	ainps.Broker.Publish(ainps.WaveEvent{
		Event:  ainps.Event_BlockFile,
		Scopes: []string{ainobj.MakeORef(ainobj.OType_Block, zoneId).String()},
		Data: &ainps.WSFileEventData{
			ZoneId:   zoneId,
			FileName: fileName,
			FileOp:   ainps.FileOp_Invalidate,
		},
	})
	return nil
}

func (c WaveClient) AppendFile(ctx context.Context, conn *connparse.Connection, data ainshrpc.FileData) error {
	dataBuf, err := base64.StdEncoding.DecodeString(data.Data64)
	if err != nil {
		return fmt.Errorf("error decoding data64: %w", err)
	}
	zoneId := conn.Host
	if zoneId == "" {
		return fmt.Errorf("zoneid not found in connection")
	}
	fileName, err := cleanPath(conn.Path)
	if err != nil {
		return fmt.Errorf("error cleaning path: %w", err)
	}
	_, err = filestore.WFS.Stat(ctx, zoneId, fileName)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("error getting blockfile info: %w", err)
		}
		var opts ainshrpc.FileOpts
		var meta ainshrpc.FileMeta
		if data.Info != nil {
			if data.Info.Opts != nil {
				opts = *data.Info.Opts
			}
			if data.Info.Meta != nil {
				meta = *data.Info.Meta
			}
		}
		if err := filestore.WFS.MakeFile(ctx, zoneId, fileName, meta, opts); err != nil {
			return fmt.Errorf("error making blockfile: %w", err)
		}
	}
	err = filestore.WFS.AppendData(ctx, zoneId, fileName, dataBuf)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("NOTFOUND: %w", err)
	}
	if err != nil {
		return fmt.Errorf("error writing to blockfile: %w", err)
	}
	ainps.Broker.Publish(ainps.WaveEvent{
		Event:  ainps.Event_BlockFile,
		Scopes: []string{ainobj.MakeORef(ainobj.OType_Block, zoneId).String()},
		Data: &ainps.WSFileEventData{
			ZoneId:   zoneId,
			FileName: fileName,
			FileOp:   ainps.FileOp_Invalidate,
		},
	})
	return nil
}

// WaveFile does not support directories, only prefix-based listing
func (c WaveClient) Mkdir(ctx context.Context, conn *connparse.Connection) error {
	return errors.ErrUnsupported
}

func (c WaveClient) MoveInternal(ctx context.Context, srcConn, destConn *connparse.Connection, opts *ainshrpc.FileCopyOpts) error {
	if srcConn.Host != destConn.Host {
		return fmt.Errorf("move internal, src and dest hosts do not match")
	}
	isDir, err := c.CopyInternal(ctx, srcConn, destConn, opts)
	if err != nil {
		return fmt.Errorf("error copying blockfile: %w", err)
	}
	recursive := opts != nil && opts.Recursive && isDir
	if err := c.Delete(ctx, srcConn, recursive); err != nil {
		return fmt.Errorf("error deleting blockfile: %w", err)
	}
	return nil
}

func (c WaveClient) CopyInternal(ctx context.Context, srcConn, destConn *connparse.Connection, opts *ainshrpc.FileCopyOpts) (bool, error) {
	return fsutil.PrefixCopyInternal(ctx, srcConn, destConn, c, opts, func(ctx context.Context, zoneId, prefix string) ([]string, error) {
		entryList := make([]string, 0)
		if err := listFilesPrefix(ctx, zoneId, prefix, func(wf *filestore.WaveFile) error {
			entryList = append(entryList, wf.Name)
			return nil
		}); err != nil {
			return nil, err
		}
		return entryList, nil
	}, func(ctx context.Context, srcPath, destPath string) error {
		srcHost := srcConn.Host
		srcFileName := strings.TrimPrefix(srcPath, srcHost+fspath.Separator)
		destHost := destConn.Host
		destFileName := strings.TrimPrefix(destPath, destHost+fspath.Separator)
		_, dataBuf, err := filestore.WFS.ReadFile(ctx, srcHost, srcFileName)
		if err != nil {
			return fmt.Errorf("error reading source blockfile: %w", err)
		}
		if err := filestore.WFS.WriteFile(ctx, destHost, destFileName, dataBuf); err != nil {
			return fmt.Errorf("error writing to destination blockfile: %w", err)
		}
		ainps.Broker.Publish(ainps.WaveEvent{
			Event:  ainps.Event_BlockFile,
			Scopes: []string{ainobj.MakeORef(ainobj.OType_Block, destHost).String()},
			Data: &ainps.WSFileEventData{
				ZoneId:   destHost,
				FileName: destFileName,
				FileOp:   ainps.FileOp_Invalidate,
			},
		})
		return nil
	})
}

func (c WaveClient) CopyRemote(ctx context.Context, srcConn, destConn *connparse.Connection, srcClient fstype.FileShareClient, opts *ainshrpc.FileCopyOpts) (bool, error) {
	if srcConn.Scheme == connparse.ConnectionTypeWave && destConn.Scheme == connparse.ConnectionTypeWave {
		return c.CopyInternal(ctx, srcConn, destConn, opts)
	}
	zoneId := destConn.Host
	if zoneId == "" {
		return false, fmt.Errorf("zoneid not found in connection")
	}
	return fsutil.PrefixCopyRemote(ctx, srcConn, destConn, srcClient, c, func(zoneId, path string, size int64, reader io.Reader) error {
		dataBuf := make([]byte, size)
		if _, err := reader.Read(dataBuf); err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("error reading tar data: %w", err)
			}
		}
		if _, err := filestore.WFS.Stat(ctx, zoneId, path); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("error getting blockfile info: %w", err)
			} else {
				if err := filestore.WFS.MakeFile(ctx, zoneId, path, ainshrpc.FileMeta{}, ainshrpc.FileOpts{}); err != nil {
					return fmt.Errorf("error making blockfile: %w", err)
				}
			}
		}

		if err := filestore.WFS.WriteFile(ctx, zoneId, path, dataBuf); err != nil {
			return fmt.Errorf("error writing to blockfile: %w", err)
		}
		ainps.Broker.Publish(ainps.WaveEvent{
			Event:  ainps.Event_BlockFile,
			Scopes: []string{ainobj.MakeORef(ainobj.OType_Block, zoneId).String()},
			Data: &ainps.WSFileEventData{
				ZoneId:   zoneId,
				FileName: path,
				FileOp:   ainps.FileOp_Invalidate,
			},
		})
		return nil
	}, opts)
}

func (c WaveClient) Delete(ctx context.Context, conn *connparse.Connection, recursive bool) error {
	zoneId := conn.Host
	if zoneId == "" {
		return fmt.Errorf("zoneid not found in connection")
	}
	prefix := conn.Path

	finfo, err := c.Stat(ctx, conn)
	exists := err == nil && !finfo.NotFound
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("error getting file info: %w", err)
	}
	if !exists {
		return nil
	}

	pathsToDelete := make([]string, 0)

	if finfo.IsDir {
		if !recursive {
			return fmt.Errorf("%v is not empty, use recursive flag to delete", prefix)
		}
		if !strings.HasSuffix(prefix, fspath.Separator) {
			prefix += fspath.Separator
		}
		if err := listFilesPrefix(ctx, zoneId, prefix, func(wf *filestore.WaveFile) error {
			pathsToDelete = append(pathsToDelete, wf.Name)
			return nil
		}); err != nil {
			return fmt.Errorf("error listing blockfiles: %w", err)
		}
	} else {
		pathsToDelete = append(pathsToDelete, prefix)
	}
	if len(pathsToDelete) > 0 {
		errs := make([]error, 0)
		for _, entry := range pathsToDelete {
			if err := filestore.WFS.DeleteFile(ctx, zoneId, entry); err != nil {
				errs = append(errs, fmt.Errorf("error deleting blockfile %s/%s: %w", zoneId, entry, err))
				continue
			}
			ainps.Broker.Publish(ainps.WaveEvent{
				Event:  ainps.Event_BlockFile,
				Scopes: []string{ainobj.MakeORef(ainobj.OType_Block, zoneId).String()},
				Data: &ainps.WSFileEventData{
					ZoneId:   zoneId,
					FileName: entry,
					FileOp:   ainps.FileOp_Delete,
				},
			})
		}
		if len(errs) > 0 {
			return fmt.Errorf("error deleting blockfiles: %v", errs)
		}
	}
	return nil
}

func listFilesPrefix(ctx context.Context, zoneId, prefix string, entryCallback func(*filestore.WaveFile) error) error {
	if zoneId == "" {
		return fmt.Errorf("zoneid not found in connection")
	}
	fileListOrig, err := filestore.WFS.ListFiles(ctx, zoneId)
	if err != nil {
		return fmt.Errorf("error listing blockfiles: %w", err)
	}
	for _, wf := range fileListOrig {
		if prefix == "" || strings.HasPrefix(wf.Name, prefix) {
			entryCallback(wf)
		}
	}
	return nil
}

func (c WaveClient) Join(ctx context.Context, conn *connparse.Connection, parts ...string) (*ainshrpc.FileInfo, error) {
	newPath := fspath.Join(append([]string{conn.Path}, parts...)...)
	newPath, err := cleanPath(newPath)
	if err != nil {
		return nil, fmt.Errorf("error cleaning path: %w", err)
	}
	conn.Path = newPath
	return c.Stat(ctx, conn)
}

func (c WaveClient) GetCapability() ainshrpc.FileShareCapability {
	return ainshrpc.FileShareCapability{
		CanAppend: true,
		CanMkdir:  false,
	}
}

func cleanPath(path string) (string, error) {
	if path == "" || path == fspath.Separator {
		return "", nil
	}
	if strings.HasPrefix(path, fspath.Separator) {
		path = path[1:]
	}
	if strings.HasPrefix(path, "~") || strings.HasPrefix(path, ".") || strings.HasPrefix(path, "..") {
		return "", fmt.Errorf("wavefile path cannot start with ~, ., or ..")
	}
	var newParts []string
	for _, part := range strings.Split(path, fspath.Separator) {
		if part == ".." {
			if len(newParts) > 0 {
				newParts = newParts[:len(newParts)-1]
			}
		} else if part != "." {
			newParts = append(newParts, part)
		}
	}
	return fspath.Join(newParts...), nil
}

func (c WaveClient) GetConnectionType() string {
	return connparse.ConnectionTypeWave
}
