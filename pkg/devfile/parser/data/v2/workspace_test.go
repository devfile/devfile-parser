package v2

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

var devworkspaceContent = v1.DevWorkspaceTemplateSpecContent{
	Components: []v1.Component{
		{
			Name: "component1",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{},
			},
		},
		{
			Name: "component2",
			ComponentUnion: v1.ComponentUnion{
				Volume: &v1.VolumeComponent{},
			},
		},
	},
}

func TestDevfile200_SetDevfileWorkspaceSpecContent(t *testing.T) {

	devfilev2 := &DevfileV2{
		v1.Devfile{},
	}

	tests := []struct {
		name                 string
		workspaceSpecContent v1.DevWorkspaceTemplateSpecContent
		expectedDevfilev2    *DevfileV2
	}{
		{
			name:                 "set workspace",
			workspaceSpecContent: devworkspaceContent,
			expectedDevfilev2: &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: devworkspaceContent,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devfilev2.SetDevfileWorkspaceSpecContent(tt.workspaceSpecContent)
			if !reflect.DeepEqual(devfilev2, tt.expectedDevfilev2) {
				t.Errorf("TestDevfile200_SetDevfileWorkspaceSpecContent() error: expected %v, got %v", tt.expectedDevfilev2, devfilev2)
			}
		})
	}
}

func TestDevfile200_SetDevfileWorkspaceSpec(t *testing.T) {

	devfilev2 := &DevfileV2{
		v1.Devfile{},
	}

	tests := []struct {
		name              string
		workspaceSpec     v1.DevWorkspaceTemplateSpec
		expectedDevfilev2 *DevfileV2
	}{
		{
			name: "set workspace spec",
			workspaceSpec: v1.DevWorkspaceTemplateSpec{
				Parent: &v1.Parent{
					ImportReference: v1.ImportReference{
						ImportReferenceUnion: v1.ImportReferenceUnion{
							Uri: "uri",
						},
					},
				},
				DevWorkspaceTemplateSpecContent: devworkspaceContent,
			},
			expectedDevfilev2: &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						Parent: &v1.Parent{
							ImportReference: v1.ImportReference{
								ImportReferenceUnion: v1.ImportReferenceUnion{
									Uri: "uri",
								},
							},
						},
						DevWorkspaceTemplateSpecContent: devworkspaceContent,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devfilev2.SetDevfileWorkspaceSpec(tt.workspaceSpec)
			if !reflect.DeepEqual(devfilev2, tt.expectedDevfilev2) {
				t.Errorf("TestDevfile200_SetDevfileWorkspaceSpec() error: expected %v, got %v", tt.expectedDevfilev2, devfilev2)
			}
		})
	}
}
