package fssh

import (
	"reflect"
	"testing"
)

func TestIsCurrentPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{
			path: "dir",
			want: true,
		}, {
			path: "s3://bucket",
			want: false,
		},
	}
	for i, test := range tests {
		got := IsCurrentPath(test.path)
		if got != test.want {
			t.Errorf("tests[%d]: got %v; want %v", i, got, test.want)
		}
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		line string
		want []string
	}{
		{
			line: "a b c",
			want: []string{"a", "b", "c"},
		}, {
			line: "a  b c",
			want: []string{"a", "b", "c"},
		}, {
			line: `a 'b' "c"`,
			want: []string{"a", "b", "c"},
		},
	}
	for i, test := range tests {
		got := ParseArgs(test.line)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("tests[%d]: got %v; want %v", i, got, test.want)
		}
	}
}

func TestParseDirURL(t *testing.T) {
	osUserHomeDirOrg := osUserHomeDir
	defer func() { osUserHomeDir = osUserHomeDirOrg }()
	osUserHomeDir = func() (string, error) {
		return "/home", nil
	}

	tests := []struct {
		dirUrl       string
		wantProtocol string
		wantHost     string
		wantDir      string
		errstr       string
	}{
		{
			dirUrl:       "",
			wantProtocol: "",
			wantHost:     ".",
			wantDir:      ".",
		}, {
			dirUrl:       "dir",
			wantProtocol: "",
			wantHost:     ".",
			wantDir:      "dir",
		}, {
			dirUrl:       "file://dir1/dir2",
			wantProtocol: "",
			wantHost:     "dir1",
			wantDir:      "dir2",
		}, {
			dirUrl:       "s3://BUCKET/DIR",
			wantProtocol: "s3://",
			wantHost:     "BUCKET",
			wantDir:      "DIR",
		}, {
			dirUrl: ":",
			errstr: `parse ":": missing protocol scheme`,
		}, {
			dirUrl:       "~~/current",
			wantProtocol: "",
			wantHost:     ".",
			wantDir:      "current",
		}, {
			dirUrl:       "~/Downloads",
			wantProtocol: "",
			wantHost:     "/home",
			wantDir:      "Downloads",
		},
	}
	for i, test := range tests {
		gotProtocol, gotHost, gotDir, err := ParseDirURL(test.dirUrl)
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

func TestDisplaySize(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{
			size: 1,
			want: "   1B",
		}, {
			size: unitKb,
			want: "   1K",
		}, {
			size: 9999 * unitKb,
			want: "  10M",
		}, {
			size: unitMb,
			want: "   1M",
		}, {
			size: unitGb,
			want: "   1G",
		}, {
			size: unitTb,
			want: "   1T",
		}, {
			size: unitPb,
			want: "   1P",
		},
	}
	for i, test := range tests {
		got := DisplaySize(test.size)
		if got != test.want {
			t.Errorf("tests[%d]: got %q; want %q", i, got, test.want)
		}
	}
}

func TestIsGlobPattern(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{
			pattern: "abc",
			want:    false,
		}, {
			pattern: "*.txt",
			want:    true,
		}, {
			pattern: "**/*.txt",
			want:    true,
		}, {
			pattern: "?.txt",
			want:    true,
		}, {
			pattern: "[a-z].txt",
			want:    true,
		}, {
			pattern: "[].txt",
			want:    true,
		},
	}
	for i, test := range tests {
		got := IsGlobPattern(test.pattern)
		if got != test.want {
			t.Errorf("tests[%d]: got %v; want %v", i, got, test.want)
		}
	}
}

func TestWithPrefixes(t *testing.T) {
	tests := []struct {
		items  []string
		prefix string
		want   []string
	}{
		{
			items:  []string{"1", "2", "3"},
			prefix: "",
			want:   []string{"1", "2", "3"},
		}, {
			items:  []string{"1", "2", "3"},
			prefix: "prefix-",
			want:   []string{"prefix-1", "prefix-2", "prefix-3"},
		},
	}
	for i, test := range tests {
		got := WithPrefixes(test.items, test.prefix)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("tests[%d]: got %v; want %v", i, got, test.want)
		}
	}
}

func TestWithSuffix(t *testing.T) {
	tests := []struct {
		items  []string
		suffix string
		want   []string
	}{
		{
			items:  []string{"1", "2", "3"},
			suffix: "",
			want:   []string{"1", "2", "3"},
		}, {
			items:  []string{"1", "2", "3"},
			suffix: ".txt",
			want:   []string{"1.txt", "2.txt", "3.txt"},
		},
	}
	for i, test := range tests {
		got := WithSuffixes(test.items, test.suffix)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("tests[%d]: got %v; want %v", i, got, test.want)
		}
	}
}

func TestSliceClone(t *testing.T) {
	intTests := []struct {
		src  []int
		want []int
	}{
		{
			src:  []int{1, 2, 3},
			want: []int{1, 2, 3},
		},
	}
	for i, test := range intTests {
		got := SliceClone(test.src)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("tests[%d]: got %v; want %v", i, got, test.want)
		}
	}
	stringTests := []struct {
		src  []string
		want []string
	}{
		{
			src:  []string{"1", "2", "3"},
			want: []string{"1", "2", "3"},
		},
	}
	for i, test := range stringTests {
		got := SliceClone(test.src)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("tests[%d]: got %v; want %v", i, got, test.want)
		}
	}
}
