package v2

import (
	"fmt"
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
)

func TestDevfile200_GetProjects(t *testing.T) {
	invalidProjectSrcType := "unknown project source type"

	tests := []struct {
		name            string
		currentProjects []v1.Project
		filterOptions   common.DevfileOptions
		wantProjects    []string
		wantErr         *string
	}{
		{
			name: "Get all the projects",
			currentProjects: []v1.Project{
				{
					Name: "project1",
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{},
					},
				},
				{
					Name: "project2",
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{},
					},
				},
			},
			filterOptions: common.DevfileOptions{},
			wantProjects:  []string{"project1", "project2"},
		},
		{
			name: "Get the filtered projects",
			currentProjects: []v1.Project{
				{
					Name: "project1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString":  "firstStringValue",
						"secondString": "secondStringValue",
					}),
					ClonePath: "/project",
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{},
					},
				},
				{
					Name: "project2",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					ClonePath: "/project",
					ProjectSource: v1.ProjectSource{
						Zip: &v1.ZipProjectSource{},
					},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstString": "firstStringValue",
				},
				ProjectOptions: common.ProjectOptions{
					ProjectSourceType: v1.GitProjectSourceType,
				},
			},
			wantProjects: []string{"project1"},
		},
		{
			name: "Wrong filter for projects",
			currentProjects: []v1.Project{
				{
					Name: "project1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString":  "firstStringValue",
						"secondString": "secondStringValue",
					}),
					ClonePath: "/project",
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{},
					},
				},
				{
					Name: "project2",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					ClonePath: "/project",
					ProjectSource: v1.ProjectSource{
						Zip: &v1.ZipProjectSource{},
					},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstStringIsWrong": "firstStringValue",
				},
			},
		},
		{
			name: "Invalid project src type",
			currentProjects: []v1.Project{
				{
					Name: "project1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
					}),
					ProjectSource: v1.ProjectSource{},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstString": "firstStringValue",
				},
			},
			wantErr: &invalidProjectSrcType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Projects: tt.currentProjects,
						},
					},
				},
			}

			projects, err := d.GetProjects(tt.filterOptions)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_GetProjects() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				// confirm the length of actual vs expected
				if len(projects) != len(tt.wantProjects) {
					t.Errorf("TestDevfile200_GetProjects() error: length of expected projects is not the same as the length of actual projects")
					return
				}

				// compare the project slices for content
				for _, wantProject := range tt.wantProjects {
					matched := false
					for _, project := range projects {
						if wantProject == project.Name {
							matched = true
						}
					}

					if !matched {
						t.Errorf("TestDevfile200_GetProjects() error: project %s not found in the devfile", wantProject)
					}
				}
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_GetProjects(): Error message should match")
			}
		})
	}
}

func TestDevfile200_AddProjects(t *testing.T) {
	currentProject := []v1.Project{
		{
			Name:      "java-starter",
			ClonePath: "/project",
		},
		{
			Name:      "quarkus-starter",
			ClonePath: "/test",
		},
	}

	d := &DevfileV2{
		v1.Devfile{
			DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
					Projects: currentProject,
				},
			},
		},
	}

	multipleDupError := fmt.Sprintf("%s\n%s", "project java-starter already exists in devfile", "project quarkus-starter already exists in devfile")

	tests := []struct {
		name    string
		args    []v1.Project
		wantErr *string
	}{
		{
			name: "It should add project",
			args: []v1.Project{
				{
					Name: "nodejs",
				},
				{
					Name: "spring-pet-clinic",
				},
			},
			wantErr: nil,
		},

		{
			name: "It should give error if tried to add already present starter project",
			args: []v1.Project{
				{
					Name: "java-starter",
				},
				{
					Name: "quarkus-starter",
				},
			},
			wantErr: &multipleDupError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := d.AddProjects(tt.args)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_AddProjects() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_AddProjects(): Error message should match")
			} else if err == nil {
				wantProjects := append(currentProject, tt.args...)

				if !reflect.DeepEqual(d.Projects, wantProjects) {
					t.Errorf("TestDevfile200_AddProjects() error: wanted: %v, got: %v, difference at %v", wantProjects, d.Projects, pretty.Compare(wantProjects, d.Projects))
				}
			}
		})
	}

}

func TestDevfile200_UpdateProject(t *testing.T) {

	missingProjectErr := "update project failed: project .* not found"

	tests := []struct {
		name              string
		args              v1.Project
		devfilev2         *DevfileV2
		expectedDevfilev2 *DevfileV2
		wantErr           *string
	}{
		{
			name: "It should update project for existing project",
			args: v1.Project{
				Name:      "nodejs",
				ClonePath: "/test",
			},
			devfilev2: &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Projects: []v1.Project{
								{
									Name:      "nodejs",
									ClonePath: "/project",
								},
								{
									Name:      "java-starter",
									ClonePath: "/project",
								},
							},
						},
					},
				},
			},
			expectedDevfilev2: &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Projects: []v1.Project{
								{
									Name:      "nodejs",
									ClonePath: "/test",
								},
								{
									Name:      "java-starter",
									ClonePath: "/project",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "It should fail to update project for non existing project",
			args: v1.Project{
				Name:      "quarkus-starter",
				ClonePath: "/project",
			},
			devfilev2: &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Projects: []v1.Project{
								{
									Name:      "nodejs",
									ClonePath: "/project",
								},
								{
									Name:      "java-starter",
									ClonePath: "/project",
								},
							},
						},
					},
				},
			},
			wantErr: &missingProjectErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devfilev2.UpdateProject(tt.args)
			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_UpdateProject() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(tt.devfilev2, tt.expectedDevfilev2) {
				t.Errorf("TestDevfile200_UpdateProject() error: wanted: %v, got: %v, difference at %v", tt.expectedDevfilev2, tt.devfilev2, pretty.Compare(tt.expectedDevfilev2, tt.devfilev2))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_UpdateProject(): Error message should match")
			}
		})
	}
}

func TestDevfile200_DeleteProject(t *testing.T) {
	missingProjectErr := "project .* is not found in the devfile"

	d := &DevfileV2{
		v1.Devfile{
			DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
					Projects: []v1.Project{
						{
							Name:      "nodejs",
							ClonePath: "/project",
						},
						{
							Name:      "java",
							ClonePath: "/project2",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name            string
		projectToDelete string
		wantProjects    []v1.Project
		wantErr         *string
	}{
		{
			name:            "Project successfully deleted",
			projectToDelete: "nodejs",
			wantProjects: []v1.Project{
				{
					Name:      "java",
					ClonePath: "/project2",
				},
			},
		},
		{
			name:            "Project not found",
			projectToDelete: "nodejs1",
			wantErr:         &missingProjectErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := d.DeleteProject(tt.projectToDelete)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_DeleteProject() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				assert.Equal(t, tt.wantProjects, d.Projects, "TestDevfile200_DeleteProject(): The two values should be the same.")
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_DeleteProject(): Error message should match")
			}
		})
	}

}

func TestDevfile200_GetStarterProjects(t *testing.T) {

	invalidStarterProjectSrcTypeErr := "unknown project source type"

	tests := []struct {
		name                   string
		currentStarterProjects []v1.StarterProject
		filterOptions          common.DevfileOptions
		wantStarterProjects    []string
		wantErr                *string
	}{
		{
			name: "Get all the starter projects",
			currentStarterProjects: []v1.StarterProject{
				{
					Name: "project1",
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{},
					},
				},
				{
					Name: "project2",
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{},
					},
				},
			},
			filterOptions:       common.DevfileOptions{},
			wantStarterProjects: []string{"project1", "project2"},
		},
		{
			name: "Get the filtered starter projects",
			currentStarterProjects: []v1.StarterProject{
				{
					Name: "project1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString":  "firstStringValue",
						"secondString": "secondStringValue",
					}),
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{},
					},
				},
				{
					Name: "project2",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					ProjectSource: v1.ProjectSource{
						Zip: &v1.ZipProjectSource{},
					},
				},
				{
					Name: "project3",
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{},
					},
				},
			},
			filterOptions: common.DevfileOptions{
				ProjectOptions: common.ProjectOptions{
					ProjectSourceType: v1.GitProjectSourceType,
				},
			},
			wantStarterProjects: []string{"project1", "project3"},
		},
		{
			name: "Wrong filter for starter projects",
			currentStarterProjects: []v1.StarterProject{
				{
					Name: "project1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString":  "firstStringValue",
						"secondString": "secondStringValue",
					}),
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{},
					},
				},
				{
					Name: "project2",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					ProjectSource: v1.ProjectSource{
						Zip: &v1.ZipProjectSource{},
					},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstStringIsWrong": "firstStringValue",
				},
			},
		},
		{
			name: "Invalid starter project src type",
			currentStarterProjects: []v1.StarterProject{
				{
					Name: "project1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
					}),
					ProjectSource: v1.ProjectSource{},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstString": "firstStringValue",
				},
			},
			wantErr: &invalidStarterProjectSrcTypeErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							StarterProjects: tt.currentStarterProjects,
						},
					},
				},
			}

			starterProjects, err := d.GetStarterProjects(tt.filterOptions)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_GetStarterProjects() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				// confirm the length of actual vs expected
				if len(starterProjects) != len(tt.wantStarterProjects) {
					t.Errorf("TestDevfile200_GetStarterProjects() error: length of expected starter projects is not the same as the length of actual starter projects")
					return
				}

				// compare the starter project slices for content
				for _, wantProject := range tt.wantStarterProjects {
					matched := false

					for _, starterProject := range starterProjects {
						if wantProject == starterProject.Name {
							matched = true
						}
					}

					if !matched {
						t.Errorf("TestDevfile200_GetStarterProjects() error: starter project %s not found in the devfile", wantProject)
					}
				}
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_GetStarterProjects(): Error message should match")
			}
		})
	}
}

func TestDevfile200_AddStarterProjects(t *testing.T) {
	currentProject := []v1.StarterProject{
		{
			Name:        "java-starter",
			Description: "starter project for java",
		},
		{
			Name:        "quarkus-starter",
			Description: "starter project for quarkus",
		},
	}

	d := &DevfileV2{
		v1.Devfile{
			DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
					StarterProjects: currentProject,
				},
			},
		},
	}
	multipleDupError := fmt.Sprintf("%s\n%s", "starterProject java-starter already exists in devfile", "starterProject quarkus-starter already exists in devfile")

	tests := []struct {
		name    string
		args    []v1.StarterProject
		wantErr *string
	}{
		{
			name: "It should add starter project",
			args: []v1.StarterProject{
				{
					Name:        "nodejs",
					Description: "starter project for nodejs",
				},
				{
					Name:        "spring-pet-clinic",
					Description: "starter project for springboot",
				},
			},
		},

		{
			name: "It should give error if tried to add already present starter project",
			args: []v1.StarterProject{
				{
					Name:        "java-starter",
					Description: "starter project for java",
				},
				{
					Name:        "quarkus-starter",
					Description: "starter project for quarkus",
				},
			},
			wantErr: &multipleDupError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := d.AddStarterProjects(tt.args)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_AddStarterProjects() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_AddStarterProjects(): Error message should match")
			} else if err == nil {
				wantProjects := append(currentProject, tt.args...)

				if !reflect.DeepEqual(d.StarterProjects, wantProjects) {
					t.Errorf("TestDevfile200_AddStarterProjects() error: wanted: %v, got: %v, difference at %v", wantProjects, d.StarterProjects, pretty.Compare(wantProjects, d.StarterProjects))
				}
			}
		})
	}

}

func TestDevfile200_UpdateStarterProject(t *testing.T) {

	missingStarterProjectErr := "update starter project failed: starter project .* not found"

	tests := []struct {
		name              string
		args              v1.StarterProject
		devfilev2         *DevfileV2
		expectedDevfilev2 *DevfileV2
		wantErr           *string
	}{
		{
			name: "It should update project for existing project",
			args: v1.StarterProject{
				Name:   "nodejs",
				SubDir: "/test",
			},
			devfilev2: &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							StarterProjects: []v1.StarterProject{
								{
									Name:   "nodejs",
									SubDir: "/project",
								},
								{
									Name:   "java-starter",
									SubDir: "/project",
								},
							},
						},
					},
				},
			},
			expectedDevfilev2: &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							StarterProjects: []v1.StarterProject{
								{
									Name:   "nodejs",
									SubDir: "/test",
								},
								{
									Name:   "java-starter",
									SubDir: "/project",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "It should fail to update project for non existing project",
			args: v1.StarterProject{
				Name:   "quarkus-starter",
				SubDir: "/project",
			},
			devfilev2: &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							StarterProjects: []v1.StarterProject{
								{
									Name:   "nodejs",
									SubDir: "/project",
								},
								{
									Name:   "java-starter",
									SubDir: "/project",
								},
							},
						},
					},
				},
			},
			wantErr: &missingStarterProjectErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devfilev2.UpdateStarterProject(tt.args)
			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_UpdateStarterProject() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(tt.devfilev2, tt.expectedDevfilev2) {
				t.Errorf("TestDevfile200_UpdateStarterProject() error: wanted: %v, got: %v, difference at %v", tt.expectedDevfilev2, tt.devfilev2, pretty.Compare(tt.expectedDevfilev2, tt.devfilev2))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_UpdateStarterProject(): Error message should match")
			}
		})
	}
}

func TestDevfile200_DeleteStarterProject(t *testing.T) {

	d := &DevfileV2{
		v1.Devfile{
			DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
					StarterProjects: []v1.StarterProject{
						{
							Name:   "nodejs",
							SubDir: "/project",
						},
						{
							Name:   "java",
							SubDir: "/project2",
						},
					},
				},
			},
		},
	}

	missingStarterProjectErr := "starter project .* is not found in the devfile"

	tests := []struct {
		name                   string
		starterProjectToDelete string
		wantStarterProjects    []v1.StarterProject
		wantErr                *string
	}{
		{
			name:                   "Starter Project successfully deleted",
			starterProjectToDelete: "nodejs",
			wantStarterProjects: []v1.StarterProject{
				{
					Name:   "java",
					SubDir: "/project2",
				},
			},
		},
		{
			name:                   "Starter Project not found",
			starterProjectToDelete: "nodejs1",
			wantErr:                &missingStarterProjectErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := d.DeleteStarterProject(tt.starterProjectToDelete)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_DeleteStarterProject() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				assert.Equal(t, tt.wantStarterProjects, d.StarterProjects, "TestDevfile200_DeleteStarterProject(): The two values should be the same.")
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_DeleteStarterProject(): Error message should match")
			}
		})
	}

}
