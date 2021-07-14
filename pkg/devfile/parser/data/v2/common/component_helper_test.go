package common

import (
	"github.com/stretchr/testify/assert"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func TestIsContainer(t *testing.T) {

	tests := []struct {
		name            string
		component       v1.Component
		wantIsSupported bool
	}{
		{
			name: "Container component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Container: &v1.ContainerComponent{},
				},
			},
			wantIsSupported: true,
		},
		{
			name: "Not a container component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Openshift: &v1.OpenshiftComponent{},
				},
			},
			wantIsSupported: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSupported := IsContainer(tt.component)
			if isSupported != tt.wantIsSupported {
				t.Errorf("TestIsContainer error: component support mismatch, expected: %v got: %v", tt.wantIsSupported, isSupported)
			}
		})
	}

}

func TestIsVolume(t *testing.T) {

	tests := []struct {
		name            string
		component       v1.Component
		wantIsSupported bool
	}{
		{
			name: "Volume component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Volume: &v1.VolumeComponent{
						Volume: v1.Volume{
							Size: "size",
						},
					},
				},
			},
			wantIsSupported: true,
		},
		{
			name: "Not a volume component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Openshift: &v1.OpenshiftComponent{},
				},
			},
			wantIsSupported: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSupported := IsVolume(tt.component)
			if isSupported != tt.wantIsSupported {
				t.Errorf("TestIsVolume error: component support mismatch, expected: %v got: %v", tt.wantIsSupported, isSupported)
			}
		})
	}

}

func TestGetComponentType(t *testing.T) {
	cmpTypeErr := "unknown component type"

	tests := []struct {
		name          string
		component     v1.Component
		wantErr       *string
		componentType v1.ComponentType
	}{
		{
			name: "Volume component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Volume: &v1.VolumeComponent{
						Volume: v1.Volume{},
					},
				},
			},
			componentType: v1.VolumeComponentType,
		},
		{
			name: "Openshift component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Openshift: &v1.OpenshiftComponent{},
				},
			},
			componentType: v1.OpenshiftComponentType,
		},
		{
			name: "Kubernetes component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Kubernetes: &v1.KubernetesComponent{},
				},
			},
			componentType: v1.KubernetesComponentType,
		},
		{
			name: "Container component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Container: &v1.ContainerComponent{},
				},
			},
			componentType: v1.ContainerComponentType,
		},
		{
			name: "Plugin component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Plugin: &v1.PluginComponent{},
				},
			},
			componentType: v1.PluginComponentType,
		},
		{
			name: "Custom component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Custom: &v1.CustomComponent{},
				},
			},
			componentType: v1.CustomComponentType,
		},
		{
			name: "Unknown component",
			component: v1.Component{
				Name:           "name",
				ComponentUnion: v1.ComponentUnion{},
			},
			wantErr: &cmpTypeErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetComponentType(tt.component)
			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestGetComponentType() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && got != tt.componentType {
				t.Errorf("TestGetComponentType error: component type mismatch, expected: %v got: %v", tt.componentType, got)
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestGetComponentType(): Error message should match")
			}
		})
	}

}
