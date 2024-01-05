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

// NewFS parses dirUrl and creates a new FS according to the protocol.
func NewFS(dirUrl string) (fsys FS, protocol string, host string, dir string, err error) {
	fsys, protocol, host, dir, err = newFS(dirUrl)
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

func newFS(dirUrl string) (fsys FS, protocol string, host string, dir string, err error) {
	protocol, host, dir, err = ParseDirURL(dirUrl)
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
		err = fsys.MkdirAll(path.Join(host, dir), os.ModePerm)
	default:
		fsys = osfs.New(host)
	}
	return
}
