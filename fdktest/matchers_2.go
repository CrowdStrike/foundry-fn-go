package fdktest

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

type WantFn func(t *testing.T, got fdk.Response)

func WantErrs2(wants ...fdk.APIError) WantFn {
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

func WantNoErrs2() WantFn {
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

func WantFileMatch2(want fdk.File) WantFn {
	line := src(1)
	return func(t *testing.T, got fdk.Response) {
		t.Helper()

		t.Cleanup(func() { _ = want.Contents.Close })

		f := WantFile(t, got.Body)

		eq(t, line, want.ContentType, f.ContentType, "ContentType")
		eq(t, line, want.Encoding, f.Encoding, "Encoding")
		eq(t, line, want.Filename, f.Filename, "Filename")

		wantContents, err := io.ReadAll(want.Contents)
		mustNoErr(t, err)

		gotContents, err := io.ReadAll(f.Contents)
		mustNoErr(t, err)

		eq(t, line, string(wantContents), string(gotContents), "Contents")
	}
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
		out += "[" + fnName + "]"
	}

	return out
}
