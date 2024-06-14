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
	equalVals(t, "text/yaml", f.ContentType)
	equalVals(t, "gzip", f.Encoding)

	f = fdk.CompressGzip(fdk.File{
		ContentType: "application/json",
		Encoding:    "gzip",
	})
	equalVals(t, "application/json", f.ContentType)
	equalVals(t, "gzip", f.Encoding)

	f = fdk.CompressGzip(fdk.File{
		ContentType: "text/plain",
		Encoding:    "deflate",
	})
	equalVals(t, "text/plain", f.ContentType)
	equalVals(t, "deflate, gzip", f.Encoding)

	f = fdk.CompressGzip(fdk.File{
		ContentType: "text/plain",
		Encoding:    "gzip, deflate",
	})
	equalVals(t, "text/plain", f.ContentType)
	equalVals(t, "gzip, deflate", f.Encoding)

	f = fdk.CompressGzip(fdk.File{
		ContentType: "application/octet-stream",
		Encoding:    "kstd, deflate",
	})
	equalVals(t, "application/octet-stream", f.ContentType)
	equalVals(t, "kstd, deflate, gzip", f.Encoding)
}
