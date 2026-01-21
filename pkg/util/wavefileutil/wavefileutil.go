package wavefileutil

import (
	"fmt"

	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/filestore"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/fsutil"
	"github.com/wavetermdev/ainterm/pkg/util/fileutil"
)

const (
	WaveFilePathPattern = "wavefile://%s/%s"
)

func WaveFileToFileInfo(wf *filestore.WaveFile) *ainshrpc.FileInfo {
	path := fmt.Sprintf(WaveFilePathPattern, wf.ZoneId, wf.Name)
	rtn := &ainshrpc.FileInfo{
		Path:          path,
		Dir:           fsutil.GetParentPathString(path),
		Name:          wf.Name,
		Opts:          &wf.Opts,
		Size:          wf.Size,
		Meta:          &wf.Meta,
		SupportsMkdir: false,
	}
	fileutil.AddMimeTypeToFileInfo(path, rtn)
	return rtn
}

func WaveFileListToFileInfoList(wfList []*filestore.WaveFile) []*ainshrpc.FileInfo {
	var fileInfoList []*ainshrpc.FileInfo
	for _, wf := range wfList {
		fileInfoList = append(fileInfoList, WaveFileToFileInfo(wf))
	}
	return fileInfoList
}
