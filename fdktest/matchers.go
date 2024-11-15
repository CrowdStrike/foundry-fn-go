package fdktest

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

// WantFn defines the assertions for a response.
type WantFn func(t *testing.T, got fdk.Response)

// Want runs a set of want fns upon hte response type.
func Want(t *testing.T, got fdk.Response, wants ...WantFn) {
	t.Helper()

	for _, want := range wants {
		want(t, got)
	}
}

// WantErrs compares verifies the errors received from teh response match the desired.
func WantErrs(wants ...fdk.APIError) WantFn {
	line := src(1)
	return func(t *testing.T, got fdk.Response) {
		t.Helper()
		if len(got.Errors) != len(wants) {
			t.Fatalf("number of errors mismatched\n\t\twant:\t%+v\n\t\tgot:\t%+v\n\t\tsource: %s", wants, got, line)
		}

		for i, want := range wants {
			err := got.Errors[i]
			if err != want {
				t.Errorf("err[%d] does not match:\n\t\twant:\t%+v\n\t\tgot:\t%+v\n\t\tsource: %s", i, want, err, line)
			}
		}
	}
}

// WantNoErrs verifies there are no errors returned.
func WantNoErrs() WantFn {
	line := src(1)
	return func(t *testing.T, got fdk.Response) {
		t.Helper()

		if len(got.Errors) == 0 {
			return
		}

		b, _ := json.MarshalIndent(got.Errors, "", "  ")
		t.Errorf("recieved unexpected errors:\n\t\tgot:\t%s\n\t\tsource: %s", string(b), line)
	}
}

// WantCode verifies the status matches desired.
func WantCode(want int) WantFn {
	line := src(1)
	return func(t *testing.T, got fdk.Response) {
		t.Helper()

		if want == got.Code {
			return
		}

		eq(t, line, want, got.Code, "Code")
	}
}

// WantFileMatch verifies the contents of the response body are
// a fdk.File type and have matching contents.
func WantFileMatch(want fdk.File) WantFn {
	line := src(1)
	return func(t *testing.T, got fdk.Response) {
		t.Helper()

		f := wantFileBody(t, line, got.Body)
		wantFileInfo(t, line, want, f)

		wantContents, err := io.ReadAll(want.Contents)
		mustNoErr(t, line, err)

		gotContents, err := io.ReadAll(f.Contents)
		mustNoErr(t, line, err)

		eq(t, line, string(wantContents), string(gotContents), "Contents")
	}
}

// WantGzipFileMatch compares the wanted file (if encoded will decode) and
// then compare it with the decoded returned file.
func WantGzipFileMatch(want fdk.File) WantFn {
	line := src(1)
	return func(t *testing.T, got fdk.Response) {
		t.Helper()

		t.Cleanup(func() { _ = want.Contents.Close() })

		wantGR, err := gzip.NewReader(want.Contents)
		mustNoErr(t, line, err)
		want.Contents = wantGR

		f := wantFileBody(t, line, got.Body)
		t.Cleanup(func() { _ = f.Contents.Close() })

		gotGr, err := gzip.NewReader(f.Contents)
		mustNoErr(t, line, err)
		f.Contents = gotGr

		got.Body = f

		wantFile(t, line, want, got)
	}
}

func wantFile(t *testing.T, line string, want fdk.File, got fdk.Response) {
	t.Helper()

	t.Cleanup(func() { _ = want.Contents.Close })

	f := wantFileBody(t, line, got.Body)
	wantFileInfo(t, line, want, f)

	wantContents, err := io.ReadAll(want.Contents)
	mustNoErr(t, line, err)

	gotContents, err := io.ReadAll(f.Contents)
	mustNoErr(t, line, err)

	eq(t, line, string(wantContents), string(gotContents), "Contents")
}

func wantFileInfo(t *testing.T, line string, want, got fdk.File) {
	t.Helper()

	eq(t, line, want.ContentType, got.ContentType, "ContentType")
	eq(t, line, want.Encoding, got.Encoding, "Encoding")
	eq(t, line, want.Filename, got.Filename, "Filename")
}

func src(deltas ...int) string {
	skip := 1
	for _, v := range deltas {
		skip += v
	}
	// Source identifies the file:line of the callsite.
	pc, file, line, _ := runtime.Caller(skip)
	if skip != 1 {
		pc, _, _, _ = runtime.Caller(1)
	}

	out := fmt.Sprintf("%s:%d", file, line)
	if fn := runtime.FuncForPC(pc); fn != nil {
		fnName := strings.TrimPrefix(fn.Name(), "github.com/CrowdStrike/foundry-fn-go/")
		out += " [" + fnName + "]"
	}

	return out
}

func eq[T comparable](t *testing.T, line string, want, got T, label string) {
	t.Helper()

	if want == got {
		return
	}

	t.Errorf("%s values do not match:\n\t\twant:\t%+v\n\t\tgot:\t%+v\n\t\tsource: %s", label, want, got, line)
}

func mustNoErr(t *testing.T, line string, err error) {
	t.Helper()

	if err == nil {
		return
	}

	t.Fatalf("recieved unexpected error:\n\t\tgot:\t%s\n\t\tsource:\t%s", err, line)
}

// wantFileBody returns the file from the fdk.Response.Body field. If
// it is not a fdk.File it will fail. Additionally, this will close
// the file in t.Cleanup. Safe to ignore closing the contents in tests.
func wantFileBody(t *testing.T, line string, body json.Marshaler) fdk.File {
	t.Helper()

	f, ok := body.(fdk.File)
	if !ok {
		t.Fatalf("did not receive correct type back; got:\t%t\n\t\tsource:\t%s", body, line)
	}

	t.Cleanup(func() {
		// ignore the error here as this may cause issues when user
		// utilizes it, we just want to make sure this is closing
		// at least once
		_ = f.Contents.Close()
	})
	return f
}
