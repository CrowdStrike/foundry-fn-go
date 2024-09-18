package fdk

import (
	"fmt"
	"testing"
	"time"
)

func TestNormalizeFile(t *testing.T) {
	var start = time.Time{}.UTC().Add(time.Hour)

	tests := []struct {
		name  string
		nowFn func() time.Time
		in    File
		want  File
	}{
		{
			name: "file with all values set should remain the same",
			in: File{
				ContentType: "application/json",
				Encoding:    "gzip",
				Filename:    "test.json",
			},
			want: File{
				ContentType: "application/json",
				Encoding:    "gzip",
				Filename:    "test.json",
			},
		},
		{
			name: "file missing encoding should remain the same",
			in: File{
				ContentType: "application/json",
				Filename:    "test.json",
			},
			want: File{
				ContentType: "application/json",
				Filename:    "test.json",
			},
		},
		{
			name: "file missing content type for filename with json extension should add content type to application/ld+json",
			in: File{
				Filename: "test.jsonld",
			},
			want: File{
				ContentType: "application/ld+json",
				Filename:    "test.jsonld",
			},
		},
		{
			name: "file missing content type for filename with jsonld extension should add content type to application/json",
			in: File{
				Filename: "test.json",
			},
			want: File{
				ContentType: "application/json",
				Filename:    "test.json",
			},
		},
		{
			name: "file missing content type for filename with jsonld extension should add content type to application/json",
			in: File{
				Filename: "test.js",
			},
			want: File{
				ContentType: "text/javascript; charset=utf-8",
				Filename:    "test.js",
			},
		},
		{
			name: "file missing content type for filename with xml extension should add content type to application/xml",
			in: File{
				Filename: "test.xml",
			},
			want: File{
				ContentType: "application/xml",
				Filename:    "test.xml",
			},
		},
		{
			name: "file missing content type for filename with tar extension should add content type to application/x-tar",
			in: File{
				Filename: "test.tar",
			},
			want: File{
				ContentType: "application/x-tar",
				Filename:    "test.tar",
			},
		},
		{
			name: "file missing content type for filename with zip extension should add content type to application/zip",
			in: File{
				Filename: "test.zip",
			},
			want: File{
				ContentType: "application/zip",
				Filename:    "test.zip",
			},
		},
		{
			name: "file missing content type for filename with html extension should add content type to text/html",
			in: File{
				Filename: "test.html",
			},
			want: File{
				ContentType: "text/html; charset=utf-8",
				Filename:    "test.html",
			},
		},
		{
			name: "file missing content type for filename with yaml extension should add content type to text/yaml",
			in: File{
				Filename: "test.yaml",
			},
			want: File{
				ContentType: "text/yaml; charset=utf-8",
				Filename:    "test.yaml",
			},
		},
		{
			name: "file missing content type for filename with yml extension should add content type to text/yaml",
			in: File{
				Filename: "test.yml",
			},
			want: File{
				ContentType: "text/yaml; charset=utf-8",
				Filename:    "test.yml",
			},
		},
		{
			name: "file missing content type for filename with txt extension should add content type to text/plain",
			in: File{
				Filename: "test.txt",
			},
			want: File{
				ContentType: "text/plain; charset=utf-8",
				Filename:    "test.txt",
			},
		},
		{
			name: "file missing content encoding for gzipped json filename should add content type and encoding",
			in: File{
				Filename: "test.json.gz",
			},
			want: File{
				ContentType: "application/json",
				Encoding:    "gzip",
				Filename:    "test.json.gz",
			},
		},
		{
			name: "file missing content encoding for gzipped yaml filename should add content type and encoding",
			in: File{
				Filename: "test.yaml.gz",
			},
			want: File{
				ContentType: "text/yaml; charset=utf-8",
				Encoding:    "gzip",
				Filename:    "test.yaml.gz",
			},
		},
		{
			name: "file missing content type and filename should set filename to timestamp",
			nowFn: func() time.Time {
				return start
			},
			in: File{},
			want: File{
				ContentType: "application/octet-stream",
				Filename:    "upload_" + start.Format(time.RFC3339),
			},
		},
		{
			name: "file missing filename with content type json set filename to timestamp.json file",
			nowFn: func() time.Time {
				return start
			},
			in: File{
				ContentType: "application/json",
			},
			want: File{
				ContentType: "application/json",
				Filename:    "upload_" + start.Format(time.RFC3339) + ".json",
			},
		},
		{
			name: "file missing filename with content type json and gz encoding set filename to timestamp.json.gz file",
			nowFn: func() time.Time {
				return start
			},
			in: File{
				ContentType: "application/json",
				Encoding:    "gzip",
			},
			want: File{
				ContentType: "application/json",
				Encoding:    "gzip",
				Filename:    "upload_" + start.Format(time.RFC3339) + ".json.gz",
			},
		},
		{
			name: "file missing filename with content type json and gz zstd encodings set filename to timestamp.json.gz file",
			nowFn: func() time.Time {
				return start
			},
			in: File{
				ContentType: "application/json",
				Encoding:    "zstd, gzip",
			},
			want: File{
				ContentType: "application/json",
				Encoding:    "zstd, gzip",
				Filename:    "upload_" + start.Format(time.RFC3339) + ".json.zst.gz",
			},
		},
		{
			name: "file missing filename with content type plain/javascript and gz zstd encodings set filename to timestamp.json.gz file",
			nowFn: func() time.Time {
				return start
			},
			in: File{
				ContentType: "text/javascript; charset=utf-8",
				Encoding:    "zstd, gzip",
			},
			want: File{
				ContentType: "text/javascript; charset=utf-8",
				Encoding:    "zstd, gzip",
				Filename:    "upload_" + start.Format(time.RFC3339) + ".js.zst.gz",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nowFn != nil {
				old := nowFn
				nowFn = tt.nowFn
				t.Cleanup(func() { nowFn = old })
			}

			got := NormalizeFile(tt.in)
			EqualVals(t, tt.want.Filename, got.Filename)
			EqualVals(t, tt.want.ContentType, got.ContentType)
			EqualVals(t, tt.want.Encoding, got.Encoding)
		})
	}
}

func EqualVals[T comparable](t testing.TB, want, got T, args ...any) bool {
	t.Helper()

	var errMsg string
	if len(args) > 0 {
		format, ok := args[0].(string)
		if ok {
			errMsg = fmt.Sprintf(format, args[1:]...)
		}
	}

	match := want == got
	if !match {
		msg := "values not equal:\n\twant:\t%#v\n\tgot:\t%#v"
		if errMsg != "" {
			msg += "\n\n\t" + errMsg
		}
		t.Errorf(msg, want, got)
	}
	return match
}
