package fdk_test

import (
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func TestFn(t *testing.T) {
	type (
		inputs struct {
			fnID      string
			fnVersion string
		}

		wantFn func(t *testing.T, gotFnID string, gotVersion int)
	)

	tests := []struct {
		name   string
		inputs inputs
		wants  wantFn
	}{
		{
			name: "fn-id set with version 1",
			inputs: inputs{
				fnID:      "fn-id",
				fnVersion: "1",
			},
			wants: func(t *testing.T, gotFnID string, gotVersion int) {
				equalVals(t, "fn-id", gotFnID)
				equalVals(t, 1, gotVersion)
			},
		},
		{
			name: "fn-id set without version",
			inputs: inputs{
				fnID:      "fn-id",
				fnVersion: "",
			},
			wants: func(t *testing.T, gotFnID string, gotVersion int) {
				equalVals(t, "fn-id", gotFnID)
				equalVals(t, 0, gotVersion)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CS_FN_ID", tt.inputs.fnID)
			t.Setenv("CS_FN_VERSION", tt.inputs.fnVersion)

			fn := fdk.Fn()
			tt.wants(t, fn.ID, fn.Version)
		})
	}

}
