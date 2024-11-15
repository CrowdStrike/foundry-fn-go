package fdktest

import (
	"encoding/json"
	"io"
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

// WantErrs compares verifies the errors received from teh response match the desired.
func WantErrs(t *testing.T, got []fdk.APIError, wants ...fdk.APIError) {
	t.Helper()

	if len(got) != len(wants) {
		t.Fatalf("number of errors mismatched\n\t\twant:\t%+v\n\t\tgot:\t%+v", wants, got)
	}

	for i, want := range wants {
		err := got[i]
		if err != want {
			t.Errorf("err[%d] does not match:\n\t\twant:\t%+v\n\t\tgot:\t%+v", i, want, err)
		}
	}
}

// WantNoErrs verifies there are no errors returned.
func WantNoErrs(t *testing.T, got []fdk.APIError) {
	t.Helper()

	if len(got) == 0 {
		return
	}

	b, _ := json.MarshalIndent(got, "", "  ")
	t.Errorf("recieved unexpected errors:\n\t\tgot:\t%s", string(b))
}

// WantStatus verifies the status matches desired.
func WantStatus(t *testing.T, want, got int) {
	t.Helper()

	if want == got {
		return
	}

	t.Errorf("status does not match:\n\t\twant:\t%d\n\t\tgot:\t%d", want, got)
}

// WantFile returns the file from the fdk.Response.Body field. If
// it is not a fdk.File it will fail. Additionally, this will close
// the file in t.Cleanup. Safe to ignore closing the contents in tests.
func WantFile(t *testing.T, body json.Marshaler) fdk.File {
	t.Helper()

	f, ok := body.(fdk.File)
	if !ok {
		t.Fatalf("did not receive correct type back; got:\t%t", body)
	}

	t.Cleanup(func() {
		// ignore the error here as this may cause issues when user
		// utilizes it, we just want to make sure this is closing
		// at least once
		_ = f.Contents.Close()
	})
	return f
}

func WantFileMatch(t *testing.T, body json.Marshaler, want fdk.File) {
	t.Helper()

	t.Cleanup(func() { _ = want.Contents.Close })

	f := WantFile(t, body)

	eq(t, want.ContentType, f.ContentType, "ContentType")
	eq(t, want.Encoding, f.Encoding, "Encoding")
	eq(t, want.Filename, f.Filename, "Filename")

	wantContents, err := io.ReadAll(want.Contents)
	mustNoErr(t, err)

	gotContents, err := io.ReadAll(f.Contents)
	mustNoErr(t, err)

	eq(t, string(wantContents), string(gotContents), "Contents")
}

func eq[T comparable](t *testing.T, want, got T, label string) {
	t.Helper()

	if want == got {
		return
	}

	t.Errorf("%s values do not match:\n\t\twant:\t%+v\n\t\tgot:\t%+v", label, want, got)
}

func mustNoErr(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		return
	}

	t.Fatalf("recieved unexpected error:\n\t\tgot:\t%s", err)
}
