package parser

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestAddSourceAttributesForOverrideAndMerge(t *testing.T) {
	importReference := v1.ImportReference{
		ImportReferenceUnion: v1.ImportReferenceUnion{
			Uri: "127.0.0.1:8080",
		},
	}
	uriImportAttribute := attributes.Attributes{}.PutString(importSourceAttribute, resolveImportReference(importReference))
	pluginOverrideImportAttribute := attributes.Attributes{}.PutString(pluginOverrideAttribute, "main devfile")
	parentOverrideImportAttribute := attributes.Attributes{}.PutString(parentOverrideAttribute, "main devfile")

	nilTemplateErr := "cannot add source attributes to nil"
	invalidTemplateTypeErr := "unknown template type"

	tests := []struct {
		name            string
		wantErr         *string
		importReference v1.ImportReference
		template        interface{}
		wantResult      interface{}
	}{
		{
			name:     "should fail if template is nil",
			template: nil,
			wantErr:  &nilTemplateErr,
		},
		{
			name:     "should fail if template is a not support type",
			template: "invalid template",
			wantErr:  &invalidTemplateTypeErr,
		},
		{
			name:            "template is with type *DevWorkspaceTemplateSpecContent",
			importReference: importReference,
			template: &v1.DevWorkspaceTemplateSpecContent{
				Components: []v1.Component{
					{
						Name: "nodejs",
						ComponentUnion: v1.ComponentUnion{
							Container: &v1.ContainerComponent{
								Container: v1.Container{
									Image: "quay.io/nodejs-10",
								},
							},
						},
					},
				},
			},
			wantResult: &v1.DevWorkspaceTemplateSpecContent{
				Components: []v1.Component{
					{
						Attributes: uriImportAttribute,
						Name:       "nodejs",
						ComponentUnion: v1.ComponentUnion{
							Container: &v1.ContainerComponent{
								Container: v1.Container{
									Image: "quay.io/nodejs-10",
								},
							},
						},
					},
				},
			},
		},
		{
			name:            "template is with type *PluginOverrides",
			importReference: v1.ImportReference{},
			template: &v1.PluginOverrides{
				Components: []v1.ComponentPluginOverride{
					{
						Name: "nodejs",
						ComponentUnionPluginOverride: v1.ComponentUnionPluginOverride{
							Container: &v1.ContainerComponentPluginOverride{
								ContainerPluginOverride: v1.ContainerPluginOverride{
									Image: "quay.io/nodejs-10",
								},
							},
						},
					},
				},
			},
			wantResult: &v1.PluginOverrides{
				Components: []v1.ComponentPluginOverride{
					{
						Name:       "nodejs",
						Attributes: pluginOverrideImportAttribute,
						ComponentUnionPluginOverride: v1.ComponentUnionPluginOverride{
							Container: &v1.ContainerComponentPluginOverride{
								ContainerPluginOverride: v1.ContainerPluginOverride{
									Image: "quay.io/nodejs-10",
								},
							},
						},
					},
				},
			},
		},
		{
			name:            "template is with type *ParentOverrides",
			importReference: v1.ImportReference{},
			template: &v1.ParentOverrides{
				Components: []v1.ComponentParentOverride{
					{
						Name: "nodejs",
						ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
							Container: &v1.ContainerComponentParentOverride{
								ContainerParentOverride: v1.ContainerParentOverride{
									Image: "quay.io/nodejs-10",
								},
							},
						},
					},
				},
			},
			wantResult: &v1.ParentOverrides{
				Components: []v1.ComponentParentOverride{
					{
						Name:       "nodejs",
						Attributes: parentOverrideImportAttribute,
						ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
							Container: &v1.ContainerComponentParentOverride{
								ContainerParentOverride: v1.ContainerParentOverride{
									Image: "quay.io/nodejs-10",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := addSourceAttributesForOverrideAndMerge(tt.importReference, tt.template)

			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_AddSourceAttributesForOverrideAndMerge() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(tt.template, tt.wantResult) {
				t.Errorf("TestAddSourceAttributesForOverrideAndMerge() error: wanted: %v, got: %v, difference at %v", tt.wantResult, tt.template, pretty.Compare(tt.template, tt.wantResult))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestAddSourceAttributesForOverrideAndMerge(): Error message should match")
			}

		})
	}

}
