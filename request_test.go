package fdk

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRequest_Fields(t *testing.T) {
	type fields struct {
		Body    json.RawMessage
		Params  *Params
		URL     string
		Method  string
		Context json.RawMessage
	}
	tests := []struct {
		name   string
		fields fields
		want   []*Field
	}{
		{
			name: "fields are correctly set in context",
			fields: fields{
				Context: json.RawMessage(`{"fields": [{"name": "a", "display": "A", "kind": "foo", "value": "baz"}]}`),
			},
			want: []*Field{{
				Name:    "a",
				Display: "A",
				Kind:    "foo",
				Value:   "baz",
			}},
		},
		{
			name: "multiple fields are correctly set in context",
			fields: fields{
				Context: json.RawMessage(`{"fields": [{"name": "a", "display": "A", "kind": "foo", "value": "baz"}, {"name": "b", "display": "B", "kind": "bar", "value": "baz"}]}`),
			},
			want: []*Field{{
				Name:    "a",
				Display: "A",
				Kind:    "foo",
				Value:   "baz",
			}, {
				Name:    "b",
				Display: "B",
				Kind:    "bar",
				Value:   "baz",
			}},
		},
		{
			name: "field name is missing",
			fields: fields{
				Context: json.RawMessage(`{"fields": [{"name": "", "display": "A", "kind": "foo", "value": "baz"}]}`),
			},
			want: []*Field{},
		},
		{
			name: "field display is missing",
			fields: fields{
				Context: json.RawMessage(`{"fields": [{"name": "a", "display": "", "kind": "foo", "value": "baz"}]}`),
			},
			want: []*Field{},
		},
		{
			name: "field kind is missing",
			fields: fields{
				Context: json.RawMessage(`{"fields": [{"name": "a", "display": "A", "kind": "", "value": "baz"}]}`),
			},
			want: []*Field{},
		},
		{
			name: "field value is missing",
			fields: fields{
				Context: json.RawMessage(`{"fields": [{"name": "a", "display": "A", "kind": "foo", "value": null}]}`),
			},
			want: []*Field{},
		},
		{
			name: "fields is not set",
			fields: fields{
				Context: json.RawMessage(`{"not_fields": 1}`),
			},
			want: []*Field{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Request{
				Body:    tt.fields.Body,
				Params:  tt.fields.Params,
				URL:     tt.fields.URL,
				Method:  tt.fields.Method,
				Context: tt.fields.Context,
			}
			if got := r.Fields(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Fields() = %v, want %v", got, tt.want)
			}
		})
	}
}
