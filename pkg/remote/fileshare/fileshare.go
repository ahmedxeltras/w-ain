package fileshare

import (
	"context"
	"fmt"
	"log"

	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainshutil"
	"github.com/wavetermdev/ainterm/pkg/remote/awsconn"
	"github.com/wavetermdev/ainterm/pkg/remote/connparse"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/fstype"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/s3fs"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/wavefs"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/wshfs"
	"github.com/wavetermdev/ainterm/pkg/util/iochan/iochantypes"
)

const (
	ErrorParsingConnection = "error creating fileshare client, could not parse connection %s"
)

// CreateFileShareClient creates a fileshare client based on the connection string
// Returns the client and the parsed connection
func CreateFileShareClient(ctx context.Context, connection string) (fstype.FileShareClient, *connparse.Connection) {
	conn, err := connparse.ParseURIAndReplaceCurrentHost(ctx, connection)
	if err != nil {
		log.Printf("error parsing connection: %v", err)
		return nil, nil
	}
	conntype := conn.GetType()
	if conntype == connparse.ConnectionTypeS3 {
		config, err := awsconn.GetConfig(ctx, connection)
		if err != nil {
			log.Printf("error getting aws config: %v", err)
			return nil, nil
		}
		return s3fs.NewS3Client(config), conn
	} else if conntype == connparse.ConnectionTypeWave {
		return wavefs.NewWaveClient(), conn
	} else if conntype == connparse.ConnectionTypeWsh {
		return wshfs.NewWshClient(), conn
	} else {
		log.Printf("unsupported connection type: %s", conntype)
		return nil, nil
	}
}

func Read(ctx context.Context, data ainshrpc.FileData) (*ainshrpc.FileData, error) {
	log.Printf("Read: %v", data.Info.Path)
	client, conn := CreateFileShareClient(ctx, data.Info.Path)
	if conn == nil || client == nil {
		return nil, fmt.Errorf(ErrorParsingConnection, data.Info.Path)
	}
	return client.Read(ctx, conn, data)
}

func ReadStream(ctx context.Context, data ainshrpc.FileData) <-chan ainshrpc.RespOrErrorUnion[ainshrpc.FileData] {
	log.Printf("ReadStream: %v", data.Info.Path)
	client, conn := CreateFileShareClient(ctx, data.Info.Path)
	if conn == nil || client == nil {
		return ainshutil.SendErrCh[ainshrpc.FileData](fmt.Errorf(ErrorParsingConnection, data.Info.Path))
	}
	return client.ReadStream(ctx, conn, data)
}

func ReadTarStream(ctx context.Context, data ainshrpc.CommandRemoteStreamTarData) <-chan ainshrpc.RespOrErrorUnion[iochantypes.Packet] {
	log.Printf("ReadTarStream: %v", data.Path)
	client, conn := CreateFileShareClient(ctx, data.Path)
	if conn == nil || client == nil {
		return ainshutil.SendErrCh[iochantypes.Packet](fmt.Errorf(ErrorParsingConnection, data.Path))
	}
	return client.ReadTarStream(ctx, conn, data.Opts)
}

func ListEntries(ctx context.Context, path string, opts *ainshrpc.FileListOpts) ([]*ainshrpc.FileInfo, error) {
	log.Printf("ListEntries: %v", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		return nil, fmt.Errorf(ErrorParsingConnection, path)
	}
	return client.ListEntries(ctx, conn, opts)
}

func ListEntriesStream(ctx context.Context, path string, opts *ainshrpc.FileListOpts) <-chan ainshrpc.RespOrErrorUnion[ainshrpc.CommandRemoteListEntriesRtnData] {
	log.Printf("ListEntriesStream: %v", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		return ainshutil.SendErrCh[ainshrpc.CommandRemoteListEntriesRtnData](fmt.Errorf(ErrorParsingConnection, path))
	}
	return client.ListEntriesStream(ctx, conn, opts)
}

func Stat(ctx context.Context, path string) (*ainshrpc.FileInfo, error) {
	log.Printf("Stat: %v", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		return nil, fmt.Errorf(ErrorParsingConnection, path)
	}
	return client.Stat(ctx, conn)
}

func PutFile(ctx context.Context, data ainshrpc.FileData) error {
	log.Printf("PutFile: %v", data.Info.Path)
	client, conn := CreateFileShareClient(ctx, data.Info.Path)
	if conn == nil || client == nil {
		return fmt.Errorf(ErrorParsingConnection, data.Info.Path)
	}
	return client.PutFile(ctx, conn, data)
}

func Mkdir(ctx context.Context, path string) error {
	log.Printf("Mkdir: %v", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		return fmt.Errorf(ErrorParsingConnection, path)
	}
	return client.Mkdir(ctx, conn)
}

func Move(ctx context.Context, data ainshrpc.CommandFileCopyData) error {
	opts := data.Opts
	if opts == nil {
		opts = &ainshrpc.FileCopyOpts{}
	}
	log.Printf("Move: srcuri: %v, desturi: %v, opts: %v", data.SrcUri, data.DestUri, opts)
	srcClient, srcConn := CreateFileShareClient(ctx, data.SrcUri)
	if srcConn == nil || srcClient == nil {
		return fmt.Errorf("error creating fileshare client, could not parse source connection %s", data.SrcUri)
	}
	destClient, destConn := CreateFileShareClient(ctx, data.DestUri)
	if destConn == nil || destClient == nil {
		return fmt.Errorf("error creating fileshare client, could not parse destination connection %s", data.DestUri)
	}
	if srcConn.Host != destConn.Host {
		isDir, err := destClient.CopyRemote(ctx, srcConn, destConn, srcClient, opts)
		if err != nil {
			return fmt.Errorf("cannot copy %q to %q: %w", data.SrcUri, data.DestUri, err)
		}
		return srcClient.Delete(ctx, srcConn, opts.Recursive && isDir)
	} else {
		return srcClient.MoveInternal(ctx, srcConn, destConn, opts)
	}
}

func Copy(ctx context.Context, data ainshrpc.CommandFileCopyData) error {
	opts := data.Opts
	if opts == nil {
		opts = &ainshrpc.FileCopyOpts{}
	}
	opts.Recursive = true
	log.Printf("Copy: srcuri: %v, desturi: %v, opts: %v", data.SrcUri, data.DestUri, opts)
	srcClient, srcConn := CreateFileShareClient(ctx, data.SrcUri)
	if srcConn == nil || srcClient == nil {
		return fmt.Errorf("error creating fileshare client, could not parse source connection %s", data.SrcUri)
	}
	destClient, destConn := CreateFileShareClient(ctx, data.DestUri)
	if destConn == nil || destClient == nil {
		return fmt.Errorf("error creating fileshare client, could not parse destination connection %s", data.DestUri)
	}
	if srcConn.Host != destConn.Host {
		_, err := destClient.CopyRemote(ctx, srcConn, destConn, srcClient, opts)
		return err
	} else {
		_, err := srcClient.CopyInternal(ctx, srcConn, destConn, opts)
		return err
	}
}

func Delete(ctx context.Context, data ainshrpc.CommandDeleteFileData) error {
	log.Printf("Delete: %v", data)
	client, conn := CreateFileShareClient(ctx, data.Path)
	if conn == nil || client == nil {
		return fmt.Errorf(ErrorParsingConnection, data.Path)
	}
	return client.Delete(ctx, conn, data.Recursive)
}

func Join(ctx context.Context, path string, parts ...string) (*ainshrpc.FileInfo, error) {
	log.Printf("Join: %v", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		return nil, fmt.Errorf(ErrorParsingConnection, path)
	}
	return client.Join(ctx, conn, parts...)
}

func Append(ctx context.Context, data ainshrpc.FileData) error {
	log.Printf("Append: %v", data.Info.Path)
	client, conn := CreateFileShareClient(ctx, data.Info.Path)
	if conn == nil || client == nil {
		return fmt.Errorf(ErrorParsingConnection, data.Info.Path)
	}
	return client.AppendFile(ctx, conn, data)
}

func GetCapability(ctx context.Context, path string) (ainshrpc.FileShareCapability, error) {
	log.Printf("GetCapability: %v", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		return ainshrpc.FileShareCapability{}, fmt.Errorf(ErrorParsingConnection, path)
	}
	return client.GetCapability(), nil
}
