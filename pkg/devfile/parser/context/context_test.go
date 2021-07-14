package parser

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPopulateFromBytes(t *testing.T) {
	failedToConvertYamlErr := "failed to convert devfile yaml to json: yaml: mapping values are not allowed in this context"

	tests := []struct {
		name        string
		dataFunc    func() []byte
		expectError *string
	}{
		{
			name:     "valid data passed",
			dataFunc: validJsonRawContent200,
		},
		{
			name:        "invalid data passed",
			dataFunc:    invalidJsonRawContent200,
			expectError: &failedToConvertYamlErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write(tt.dataFunc())
				if err != nil {
					t.Error(err)
				}
			}))
			var (
				d = DevfileCtx{
					url: testServer.URL,
				}
			)
			defer testServer.Close()
			err := d.PopulateFromURL()
			if (tt.expectError != nil) != (err != nil) {
				t.Errorf("TestPopulateFromBytes(): unexpected error: %v, wantErr: %v", err, tt.expectError)
			} else if tt.expectError != nil {
				assert.Regexp(t, *tt.expectError, err.Error(), "TestPopulateFromBytes(): Error message should match")
			}
		})
	}
}

func TestPopulateFromInvalidURL(t *testing.T) {
	expectError := ".*invalid URI for request"
	t.Run("Populate from invalid URL", func(t *testing.T) {
		var (
			d = DevfileCtx{
				url: "blah",
			}
		)

		err := d.PopulateFromURL()

		if err == nil {
			t.Errorf("TestPopulateFromInvalidURL(): expected an error, didn't get one")
		} else {
			assert.Regexp(t, expectError, err.Error(), "TestPopulateFromInvalidURL(): Error message should match")
		}
	})
}

func invalidJsonRawContent200() []byte {
	return []byte(InvalidDevfileContent)
}
