package fssh

import (
	"reflect"
	"testing"

	"github.com/jarxorg/gcsfs"
	"github.com/jarxorg/s3fs"
	"github.com/jarxorg/wfs/memfs"
	"github.com/jarxorg/wfs/osfs"
)

func TestNewFS(t *testing.T) {
	tests := []struct {
		dirUrl       string
		wantType     reflect.Type
		wantProtocol string
		wantHost     string
		wantDir      string
		errstr       string
	}{
		{
			dirUrl:       "",
			wantType:     reflect.TypeOf(osfs.New("")),
			wantProtocol: "",
			wantHost:     ".",
			wantDir:      ".",
		}, {
			dirUrl: ":",
			errstr: `parse ":": missing protocol scheme`,
		}, {
			dirUrl: "not-found",
			errstr: "stat not-found: no such file or directory",
		}, {
			dirUrl: "fs_test.go",
			errstr: "not directory: fs_test.go",
		},
	}
	for i, test := range tests {
		gotFS, gotProtocol, gotHost, gotDir, err := NewFS(test.dirUrl)
		if test.errstr != "" {
			if err == nil {
				t.Fatalf("tests[%d]: no error; want %s", i, test.errstr)
			}
			if err.Error() != test.errstr {
				t.Errorf("tests[%d]: got err %v; want %s", i, err, test.errstr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("tests[%d]: err %v", i, err)
			continue
		}
		gotType := reflect.TypeOf(gotFS)
		if gotType != test.wantType {
			t.Errorf("tests[%d]: got fs %v, want %v", i, gotType, test.wantType)
		}
		if gotProtocol != test.wantProtocol {
			t.Errorf("tests[%d]: got protocol %v; want %v", i, gotProtocol, test.wantProtocol)
		}
		if gotHost != test.wantHost {
			t.Errorf("tests[%d]: got host %v; want %v", i, gotHost, test.wantHost)
		}
		if gotDir != test.wantDir {
			t.Errorf("tests[%d]: got dir %v; want %v", i, gotDir, test.wantDir)
		}
	}
}

func Test_newFS(t *testing.T) {
	tests := []struct {
		dirUrl       string
		wantType     reflect.Type
		wantProtocol string
		wantHost     string
		wantDir      string
		errstr       string
	}{
		{
			dirUrl:       "",
			wantType:     reflect.TypeOf(osfs.New("")),
			wantProtocol: "",
			wantHost:     ".",
			wantDir:      ".",
		}, {
			dirUrl:       "dir",
			wantType:     reflect.TypeOf(osfs.New("")),
			wantProtocol: "",
			wantHost:     ".",
			wantDir:      "dir",
		}, {
			dirUrl:       "file://dir1/dir2",
			wantType:     reflect.TypeOf(osfs.New("")),
			wantProtocol: "",
			wantHost:     "dir1",
			wantDir:      "dir2",
		}, {
			dirUrl:       "mem://",
			wantType:     reflect.TypeOf(memfs.New()),
			wantProtocol: "mem://",
			wantHost:     "",
			wantDir:      ".",
		}, {
			dirUrl:       "s3://BUCKET/DIR",
			wantType:     reflect.TypeOf(s3fs.New("")),
			wantProtocol: "s3://",
			wantHost:     "BUCKET",
			wantDir:      "DIR",
		}, {
			dirUrl:       "gs://BUCKET/DIR",
			wantType:     reflect.TypeOf(gcsfs.New("")),
			wantProtocol: "gs://",
			wantHost:     "BUCKET",
			wantDir:      "DIR",
		}, {
			dirUrl: ":",
			errstr: `parse ":": missing protocol scheme`,
		},
	}
	for i, test := range tests {
		gotFS, gotProtocol, gotHost, gotDir, err := newFS(test.dirUrl)
		if test.errstr != "" {
			if err == nil {
				t.Fatalf("tests[%d]: no error; want %s", i, test.errstr)
			}
			if err.Error() != test.errstr {
				t.Errorf("tests[%d]: got err %v; want %s", i, err, test.errstr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("tests[%d]: err %v", i, err)
			continue
		}
		gotType := reflect.TypeOf(gotFS)
		if gotType != test.wantType {
			t.Errorf("tests[%d]: got fs %v, want %v", i, gotType, test.wantType)
		}
		if gotProtocol != test.wantProtocol {
			t.Errorf("tests[%d]: got protocol %v; want %v", i, gotProtocol, test.wantProtocol)
		}
		if gotHost != test.wantHost {
			t.Errorf("tests[%d]: got host %v; want %v", i, gotHost, test.wantHost)
		}
		if gotDir != test.wantDir {
			t.Errorf("tests[%d]: got dir %v; want %v", i, gotDir, test.wantDir)
		}
	}
}
