package fdk_test

import (
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func TestCompressGzip(t *testing.T) {
	f := fdk.CompressGzip(fdk.File{
		ContentType: "text/yaml",
		Encoding:    "",
	})
	fdk.EqualVals(t, "text/yaml", f.ContentType)
	fdk.EqualVals(t, "gzip", f.Encoding)

	f = fdk.CompressGzip(fdk.File{
		ContentType: "application/json",
		Encoding:    "gzip",
	})
	fdk.EqualVals(t, "application/json", f.ContentType)
	fdk.EqualVals(t, "gzip", f.Encoding)

	f = fdk.CompressGzip(fdk.File{
		ContentType: "text/plain",
		Encoding:    "deflate",
	})
	fdk.EqualVals(t, "text/plain", f.ContentType)
	fdk.EqualVals(t, "deflate, gzip", f.Encoding)

	f = fdk.CompressGzip(fdk.File{
		ContentType: "text/plain",
		Encoding:    "gzip, deflate",
	})
	fdk.EqualVals(t, "text/plain", f.ContentType)
	fdk.EqualVals(t, "gzip, deflate", f.Encoding)

	f = fdk.CompressGzip(fdk.File{
		ContentType: "application/octet-stream",
		Encoding:    "kstd, deflate",
	})
	fdk.EqualVals(t, "application/octet-stream", f.ContentType)
	fdk.EqualVals(t, "kstd, deflate, gzip", f.Encoding)
}
