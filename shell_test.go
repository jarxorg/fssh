package fssh

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"testing"
	"time"
)

func setupTestNewShell(t *testing.T) (done func()) {
	tmpDir, err := os.MkdirTemp("", "*-test")
	if err != nil {
		t.Fatal(err)
	}

	osUserHomeDirOrg := osUserHomeDir
	osUserHomeDir = func() (string, error) {
		return tmpDir, nil
	}

	return func() {
		osUserHomeDir = osUserHomeDirOrg
		_ = os.RemoveAll(tmpDir)
	}
}

func TestNewShell(t *testing.T) {
	done := setupTestNewShell(t)
	defer done()

	tests := []struct {
		setup  func() (done func())
		dirUrl string
		errstr string
	}{
		{
			setup:  func() (done func()) { return func() {} },
			dirUrl: "",
		}, {
			setup:  func() (done func()) { return func() {} },
			dirUrl: ":",
			errstr: `parse ":": missing protocol scheme`,
		}, {
			setup: func() (done func()) {
				osUserHomeDirOrg := osUserHomeDir
				osUserHomeDir = func() (string, error) {
					return "", errors.New("test-error")
				}
				return func() { osUserHomeDir = osUserHomeDirOrg }
			},
			dirUrl: "",
			errstr: "test-error",
		},
	}
	for i, test := range tests {
		func() {
			done2 := test.setup()
			defer done2()

			sh, err := NewShell(test.dirUrl)
			if test.errstr != "" {
				if err == nil {
					t.Fatalf("tests[%d]: no error; want %s", i, test.errstr)
				}
				if err.Error() != test.errstr {
					t.Errorf("tests[%d]: got err %v; want %s", i, err, test.errstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("tests[%d]: err %v", i, err)
			}
			_ = sh.Close()
		}()
	}
}

func TestShellRun(t *testing.T) {
	done := setupTestNewShell(t)
	defer done()

	sh, err := NewShell("mem://")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(1 * time.Millisecond)
		_ = sh.Close()
	}()

	if err := sh.Run(); err != nil {
		t.Errorf("failed to run: %v", err)
	}
}

func TestShellExecCommand(t *testing.T) {
	done := setupTestNewShell(t)
	defer done()

	testCmd := &testCommand{
		name:    "test",
		flagSet: &flag.FlagSet{},
	}
	RegisterNewCommandFunc(func() Command {
		return testCmd
	})
	defer DeregisterNewCommandFunc("test")

	sh, err := NewShell("mem://")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		args   []string
		errstr string
	}{
		{
			args: []string{},
		}, {
			args: []string{"test"},
		}, {
			args:   []string{"unknown"},
			errstr: "command not found: unknown",
		}, {
			args: []string{"test", "-h"},
		}, {
			args:   []string{"test", "-a"},
			errstr: "flag provided but not defined: -a",
		},
	}
	for i, test := range tests {
		err := sh.ExecCommand(test.args)
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
		}
	}
}

func TestShellUsage(t *testing.T) {
	done := setupTestNewShell(t)
	defer done()

	testCmd := &testCommand{
		name:    "test",
		flagSet: &flag.FlagSet{},
	}
	RegisterNewCommandFunc(func() Command {
		return testCmd
	})
	defer DeregisterNewCommandFunc("test")

	sh, err := NewShell("mem://")
	if err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	sh.Usage(buf)

	want := "Commands:\n  test\n"
	got := buf.String()
	if got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestShellSubFS(t *testing.T) {
	done := setupTestNewShell(t)
	defer done()

	sh, err := NewShell("mem://root")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		dirUrl  string
		wantFS  FS
		wantDir string
		errstr  string
	}{
		{
			dirUrl:  "dir",
			wantDir: "dir",
		}, {
			dirUrl:  "mem://test",
			wantDir: ".",
		}, {
			dirUrl: "mem://invalid url",
			errstr: `parse "mem://invalid url": invalid character " " in host name`,
		},
	}
	for i, test := range tests {
		_, gotDir, err := sh.SubFS(test.dirUrl)
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
		}
		if gotDir != test.wantDir {
			t.Errorf("tests[%d]: got dir %v; want %v", i, gotDir, test.wantDir)
		}
	}
}
