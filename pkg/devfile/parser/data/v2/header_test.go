package v2

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
)

func TestDevfile200_GetSchemaVersion(t *testing.T) {

	tests := []struct {
		name                  string
		expectedSchemaVersion string
		devfilev2             *DevfileV2
	}{
		{
			name: "Get the schema version",
			devfilev2: &DevfileV2{
				v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.0.0",
					},
				},
			},
			expectedSchemaVersion: "2.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := tt.devfilev2.GetSchemaVersion()
			if version != tt.expectedSchemaVersion {
				t.Errorf("TestDevfile200_GetSchemaVersion() error: schema version did not match. Expected %s, got %s", tt.expectedSchemaVersion, version)
			}
		})
	}
}

func TestDevfile200_SetSchemaVersion(t *testing.T) {

	tests := []struct {
		name              string
		schemaVersion     string
		devfilev2         *DevfileV2
		expectedDevfilev2 *DevfileV2
	}{
		{
			name: "empty header",
			devfilev2: &DevfileV2{
				v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{},
				},
			},
			schemaVersion: "2.0.0",
			expectedDevfilev2: &DevfileV2{
				v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.0.0",
					},
				},
			},
		},
		{
			name: "override existing header",
			devfilev2: &DevfileV2{
				v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "1.0.0",
					},
				},
			},
			schemaVersion: "2.0.0",
			expectedDevfilev2: &DevfileV2{
				v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.0.0",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.devfilev2.SetSchemaVersion(tt.schemaVersion)
			if !reflect.DeepEqual(tt.devfilev2, tt.expectedDevfilev2) {
				t.Errorf("TestDevfile200_SetSchemaVersion() error: expected %v, got %v", tt.expectedDevfilev2, tt.devfilev2)
			}
		})
	}
}

func TestDevfile200_GetMetadata(t *testing.T) {

	tests := []struct {
		name                   string
		devfilev2              *DevfileV2
		expectedName           string
		expectedVersion        string
		expectedAttribute      string
		expectedDockerfilePath string
	}{
		{
			name: "Get the metadata",
			devfilev2: &DevfileV2{
				v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						Metadata: devfilepkg.DevfileMetadata{
							Name:    "nodejs",
							Version: "1.2.3",
							Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
								"alpha.build-dockerfile": "/relative/path/to/Dockerfile",
							}),
						},
					},
				},
			},
			expectedName:           "nodejs",
			expectedVersion:        "1.2.3",
			expectedDockerfilePath: "/relative/path/to/Dockerfile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := tt.devfilev2.GetMetadata()
			if metadata.Name != tt.expectedName {
				t.Errorf("TestDevfile200_GetMetadata() error: name mismatch, expected %v, got %v", tt.expectedName, metadata.Name)
			}
			if metadata.Version != tt.expectedVersion {
				t.Errorf("TestDevfile200_GetMetadata() error: version mismatch, expected %v, got %v", tt.expectedVersion, metadata.Version)
			}
			if metadata.Attributes.GetString("alpha.build-dockerfile", nil) != tt.expectedDockerfilePath {
				t.Errorf("TestDevfile200_GetMetadata() error: dockor file path mismatch, expected %v, got %v", tt.expectedDockerfilePath, metadata.Attributes.GetString("alpha.build-dockerfile", nil))
			}
		})
	}
}

func TestDevfile200_SetSetMetadata(t *testing.T) {

	tests := []struct {
		name              string
		metadata          devfilepkg.DevfileMetadata
		devfilev2         *DevfileV2
		expectedDevfilev2 *DevfileV2
	}{
		{
			name: "empty header",
			devfilev2: &DevfileV2{
				v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{},
				},
			},
			metadata: devfilepkg.DevfileMetadata{
				Name:    "nodejs",
				Version: "2.0.0",
			},
			expectedDevfilev2: &DevfileV2{
				v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						Metadata: devfilepkg.DevfileMetadata{
							Name:    "nodejs",
							Version: "2.0.0",
						},
					},
				},
			},
		},
		{
			name: "override existing header",
			devfilev2: &DevfileV2{
				v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.0.0",
					},
				},
			},
			metadata: devfilepkg.DevfileMetadata{
				Name:    "nodejs",
				Version: "2.1.0",
				Attributes: attributes.Attributes{}.FromMap(map[string]interface{}{
					"xyz": "xyz",
				}, nil),
				DisplayName:       "display",
				Description:       "decription",
				Tags:              []string{"tag1"},
				Icon:              "icon",
				GlobalMemoryLimit: "globalmemorylimit",
				ProjectType:       "projectype",
				Language:          "language",
				Website:           "website",
			},
			expectedDevfilev2: &DevfileV2{
				v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: "2.0.0",
						Metadata: devfilepkg.DevfileMetadata{
							Name:    "nodejs",
							Version: "2.1.0",
							Attributes: attributes.Attributes{}.FromMap(map[string]interface{}{
								"xyz": "xyz",
							}, nil),
							DisplayName:       "display",
							Description:       "decription",
							Tags:              []string{"tag1"},
							Icon:              "icon",
							GlobalMemoryLimit: "globalmemorylimit",
							ProjectType:       "projectype",
							Language:          "language",
							Website:           "website",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.devfilev2.SetMetadata(tt.metadata)
			if !reflect.DeepEqual(tt.devfilev2, tt.expectedDevfilev2) {
				t.Errorf("TestDevfile200_SetSchemaVersion() error: expected %v, got %v", tt.expectedDevfilev2, tt.devfilev2)
			}
		})
	}
}
