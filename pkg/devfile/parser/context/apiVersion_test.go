package parser

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSetDevfileAPIVersion(t *testing.T) {

	const (
		schemaVersion          = "2.0.0"
		validJson              = `{"schemaVersion": "2.0.0"}`
		emptyJson              = "{}"
		emptySchemaVersionJson = `{"schemaVersion": ""}`
		devfilePath            = "/testpath/devfile.yaml"
		devfileURL             = "http://server/devfile.yaml"
	)

	// test table
	tests := []struct {
		name       string
		devfileCtx DevfileCtx
		want       string
		wantErr    error
	}{
		{
			name:       "valid schemaVersion",
			devfileCtx: DevfileCtx{rawContent: []byte(validJson), absPath: devfilePath},
			want:       schemaVersion,
			wantErr:    nil,
		},
		{
			name:       "schemaVersion not present",
			devfileCtx: DevfileCtx{rawContent: []byte(emptyJson), absPath: devfilePath},
			want:       "",
			wantErr:    fmt.Errorf("schemaVersion not present in devfile: %s", devfilePath),
		},
		{
			name:       "schemaVersion empty",
			devfileCtx: DevfileCtx{rawContent: []byte(emptySchemaVersionJson), url: devfileURL},
			want:       "",
			wantErr:    fmt.Errorf("schemaVersion in devfile: %s cannot be empty", devfileURL),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// new devfile context object
			d := tt.devfileCtx

			// SetDevfileAPIVersion
			gotErr := d.SetDevfileAPIVersion()
			got := d.apiVersion

			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("TestSetDevfileAPIVersion() unexpected error: '%v', wantErr: '%v'", gotErr, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("TestSetDevfileAPIVersion() want: '%v', got: '%v'", tt.want, got)
			}
		})
	}
}

func TestGetApiVersion(t *testing.T) {

	const (
		apiVersion = "2.0.0"
	)

	t.Run("get apiVersion", func(t *testing.T) {

		var (
			d    = DevfileCtx{apiVersion: apiVersion}
			want = apiVersion
			got  = d.GetApiVersion()
		)

		if got != want {
			t.Errorf("TestGetApiVersion() want: '%v', got: '%v'", want, got)
		}
	})
}
