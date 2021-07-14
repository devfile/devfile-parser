package v2

import (
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/testingutil"
	"github.com/stretchr/testify/assert"
)

func TestDevfile200_AddVolumeMount(t *testing.T) {
	image0 := "some-image-0"

	container0 := "container0"
	container1 := "container1"

	volume0 := "volume0"
	volume1 := "volume1"

	samePathPresentErr := "unable to mount volume .*, as another volume .* is mounted to the same path .* in the container .*"
	missingContainerErr := "container component .* is not found in the devfile"

	type args struct {
		componentName string
		volumeMounts  []v1.VolumeMount
	}
	tests := []struct {
		name              string
		currentComponents []v1.Component
		wantComponents    []v1.Component
		args              args
		wantErr           *string
	}{
		{
			name: "add the volume mount when other mounts are present",
			currentComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
				{
					Name: container1,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
			},
			args: args{
				volumeMounts: []v1.VolumeMount{
					testingutil.GetFakeVolumeMount(volume0, "/path0"),
				},
				componentName: container0,
			},
			wantComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
									testingutil.GetFakeVolumeMount(volume0, "/path0"),
								},
							},
						},
					},
				},
				{
					Name: container1,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "error out when same path is present in the container",
			currentComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume0, "/data0"),
									testingutil.GetFakeVolumeMount(volume1, "/data1"),
								},
							},
						},
					},
				},
			},
			args: args{
				volumeMounts: []v1.VolumeMount{
					testingutil.GetFakeVolumeMount(volume0, "/data1"),
					testingutil.GetFakeVolumeMount(volume1, "/data0"),
				},
				componentName: container0,
			},
			wantErr: &samePathPresentErr,
		},
		{
			name: "error out when the specified container is not found",
			currentComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
			},
			args: args{
				volumeMounts: []v1.VolumeMount{
					testingutil.GetFakeVolumeMount(volume0, "/data"),
				},
				componentName: container1,
			},
			wantErr: &missingContainerErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Components: tt.currentComponents,
						},
					},
				},
			}

			err := d.AddVolumeMounts(tt.args.componentName, tt.args.volumeMounts)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_AddVolumeMount() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				assert.Equal(t, tt.wantComponents, d.Components, "TestDevfile200_AddVolumeMount(): The two values should be the same.")
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_AddVolumeMount(): Error message should match")
			}
		})
	}
}

func TestDevfile200_DeleteVolumeMounts(t *testing.T) {

	d := &DevfileV2{
		v1.Devfile{
			DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
					Components: []v1.Component{
						{
							Name: "comp1",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										VolumeMounts: []v1.VolumeMount{
											testingutil.GetFakeVolumeMount("comp2", "/path"),
											testingutil.GetFakeVolumeMount("comp2", "/path2"),
											testingutil.GetFakeVolumeMount("comp3", "/path"),
											testingutil.GetFakeVolumeMount("comp2", "/path3"),
										},
									},
								},
							},
						},
						{
							Name: "comp4",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										VolumeMounts: []v1.VolumeMount{
											testingutil.GetFakeVolumeMount("comp2", "/path"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	missingMountErr := "volume mount .* is not found in the devfile"

	tests := []struct {
		name             string
		volMountToDelete string
		wantComponents   []v1.Component
		wantErr          *string
	}{
		{
			name:             "Volume Component with mounts",
			volMountToDelete: "comp2",
			wantComponents: []v1.Component{
				{
					Name: "comp1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount("comp3", "/path"),
								},
							},
						},
					},
				},
				{
					Name: "comp4",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{},
							},
						},
					},
				},
			},
		},
		{
			name:             "Missing mount name",
			volMountToDelete: "comp1",
			wantErr:          &missingMountErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := d.DeleteVolumeMount(tt.volMountToDelete)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_DeleteVolumeMounts() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				assert.Equal(t, tt.wantComponents, d.Components, "TestDevfile200_DeleteVolumeMounts(): The two values should be the same.")
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_DeleteVolumeMounts(): Error message should match")
			}
		})
	}

}

func TestDevfile200_GetVolumeMountPaths(t *testing.T) {

	volumeNotMountedErr := "volume .* not mounted to component .*"
	missingContainerErr := "container component .* is not found in the devfile"

	tests := []struct {
		name              string
		currentComponents []v1.Component
		mountName         string
		componentName     string
		wantPaths         []string
		wantErr           *string
	}{
		{
			name: "vol is mounted on the specified container component",
			currentComponents: []v1.Component{
				{
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount("volume1", "/path"),
									testingutil.GetFakeVolumeMount("volume1", "/path2"),
								},
							},
						},
					},
					Name: "component1",
				},
			},
			wantPaths:     []string{"/path", "/path2"},
			mountName:     "volume1",
			componentName: "component1",
		},
		{
			name: "vol is not mounted on the specified container component",
			currentComponents: []v1.Component{
				{
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount("volume1", "/path"),
								},
							},
						},
					},
					Name: "component1",
				},
			},
			mountName:     "volume2",
			componentName: "component1",
			wantErr:       &volumeNotMountedErr,
		},
		{
			name: "invalid specified container",
			currentComponents: []v1.Component{
				{
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount("volume1", "/path"),
								},
							},
						},
					},
					Name: "component1",
				},
			},
			mountName:     "volume1",
			componentName: "component2",
			wantErr:       &missingContainerErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Components: tt.currentComponents,
						},
					},
				},
			}
			gotPaths, err := d.GetVolumeMountPaths(tt.mountName, tt.componentName)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_GetVolumeMountPaths() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				if len(gotPaths) != len(tt.wantPaths) {
					t.Errorf("TestDevfile200_GetVolumeMountPaths() error: mount paths length mismatch, expected %v, actual %v", len(tt.wantPaths), len(gotPaths))
				}

				for _, wantPath := range tt.wantPaths {
					matched := false
					for _, gotPath := range gotPaths {
						if wantPath == gotPath {
							matched = true
						}
					}

					if !matched {
						t.Errorf("TestDevfile200_GetVolumeMountPaths() error: unable to find the wanted mount path %s in the actual mount paths slice", wantPath)
					}
				}
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_DeleteVolumeMounts(): Error message should match")
			}
		})
	}
}
