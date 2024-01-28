package fssh

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/jarxorg/gcsfs"
	"github.com/jarxorg/s3fs"
	"github.com/jarxorg/wfs"
	"github.com/jarxorg/wfs/memfs"
	"github.com/jarxorg/wfs/osfs"
)

// FS is writable FS.
type FS wfs.WriteFileFS

// NewFS parses nameUrl and creates a new FS according to the protocol.
func NewFS(filenameUrl string) (fsys FS, protocol string, host string, filename string, err error) {
	protocol, host, filename, err = ParseURI(filenameUrl)
	if err != nil {
		return
	}
	switch protocol {
	case "s3://":
		fsys = s3fs.New(host)
	case "gs://":
		fsys = gcsfs.New(host)
	case "mem://":
		fsys = memfs.New()
		err = fsys.MkdirAll(path.Join(host, filename), os.ModePerm)
	default:
		fsys = osfs.New(host)
	}
	return
}

// NewDirFS parses dirUrl and creates a new FS according to the protocol.
func NewDirFS(dirUrl string) (fsys FS, protocol string, host string, dir string, err error) {
	fsys, protocol, host, dir, err = NewFS(dirUrl)
	if err != nil {
		return
	}
	var info fs.FileInfo
	info, err = fs.Stat(fsys, dir)
	if err != nil {
		return
	}
	if !info.IsDir() {
		err = fmt.Errorf("not directory: %s", dir)
		return
	}
	return
}
