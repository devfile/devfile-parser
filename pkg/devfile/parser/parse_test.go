package parser

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/testingutil"
	"github.com/kylelemons/godebug/pretty"
	kubev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const schemaV200 = "2.0.0"

func Test_parseParentAndPluginFromURI(t *testing.T) {
	const uri1 = "127.0.0.1:8080"
	const uri2 = "127.0.0.1:9090"
	importFromUri1 := attributes.Attributes{}.PutString(importSourceAttribute, fmt.Sprintf("uri: http://%s", uri1))
	importFromUri2 := attributes.Attributes{}.PutString(importSourceAttribute, fmt.Sprintf("uri: http://%s", uri2))
	parentOverridesFromMainDevfile := attributes.Attributes{}.PutString(importSourceAttribute,
		fmt.Sprintf("uri: http://%s", uri1)).PutString(parentOverrideAttribute, "main devfile")
	pluginOverridesFromMainDevfile := attributes.Attributes{}.PutString(importSourceAttribute,
		fmt.Sprintf("uri: http://%s", uri2)).PutString(pluginOverrideAttribute, "main devfile")

	parentDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Commands: []v1.Command{
							{
								Id: "devrun",
								CommandUnion: v1.CommandUnion{
									Exec: &v1.ExecCommand{
										WorkingDir:  "/projects",
										CommandLine: "npm run",
									},
								},
							},
						},
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
						Events: &v1.Events{
							DevWorkspaceEvents: v1.DevWorkspaceEvents{
								PostStart: []string{"post-start-0"},
							},
						},
						Projects: []v1.Project{
							{
								ClonePath: "/data",
								ProjectSource: v1.ProjectSource{
									Git: &v1.GitProjectSource{
										GitLikeProjectSource: v1.GitLikeProjectSource{
											Remotes: map[string]string{
												"master": "https://githube.com/somerepo/someproject.git",
											},
										},
									},
								},
								Name: "nodejs-starter",
							},
						},
					},
				},
			},
		},
	}

	parentCmdAlreadyDefinedErr := "Some Commands are already defined in parent.* If you want to override them, you should do it in the parent scope."
	parentCmpAlreadyDefinedErr := "Some Components are already defined in parent.* If you want to override them, you should do it in the parent scope."
	parentProjectAlreadyDefinedErr := "Some Projects are already defined in parent.* If you want to override them, you should do it in the parent scope."
	pluginCmdAlreadyDefinedErr := "Some Commands are already defined in plugin.* If you want to override them, you should do it in the plugin scope."
	pluginCmpAlreadyDefinedErr := "Some Components are already defined in plugin.* If you want to override them, you should do it in the plugin scope."
	pluginProjectAlreadyDefinedErr := "Some Projects are already defined in plugin.* If you want to override them, you should do it in the plugin scope."
	newCmdErr := "Some Commands do not override any existing element.* They should be defined in the main body, as new elements, not in the overriding section"
	newCmpErr := "Some Components do not override any existing element.* They should be defined in the main body, as new elements, not in the overriding section"
	newProjectErr := "Some Projects do not override any existing element.* They should be defined in the main body, as new elements, not in the overriding section"
	importCycleErr := "devfile has an cycle in references: main devfile -> .*"

	type args struct {
		devFileObj DevfileObj
	}
	tests := []struct {
		name                   string
		args                   args
		parentDevfile          DevfileObj
		pluginDevfile          DevfileObj
		pluginOverride         v1.PluginOverrides
		wantDevFile            DevfileObj
		wantErr                []string
		testRecursiveReference bool
	}{
		{
			name: "it should override the requested parent's data and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								Parent: &v1.Parent{
									ParentOverrides: v1.ParentOverrides{
										Commands: []v1.CommandParentOverride{
											{
												Id: "devrun",
												CommandUnionParentOverride: v1.CommandUnionParentOverride{
													Exec: &v1.ExecCommandParentOverride{
														WorkingDir: "/projects/nodejs-starter",
													},
												},
											},
										},
										Components: []v1.ComponentParentOverride{
											{
												Name: "nodejs",
												ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
													Container: &v1.ContainerComponentParentOverride{
														ContainerParentOverride: v1.ContainerParentOverride{
															Image: "quay.io/nodejs-12",
														},
													},
												},
											},
										},
										Projects: []v1.ProjectParentOverride{
											{
												ClonePath: "/projects",
												Name:      "nodejs-starter",
											},
										},
									},
								},
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Commands: []v1.Command{
										{
											Id: "devbuild",
											CommandUnion: v1.CommandUnion{
												Exec: &v1.ExecCommand{
													WorkingDir: "/projects/nodejs-starter",
												},
											},
										},
									},
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
									Events: &v1.Events{
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
											PostStop: []string{"post-stop"},
										},
									},
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: parentDevfile,
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Attributes: parentOverridesFromMainDevfile,
										Id:         "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects/nodejs-starter",
											},
										},
									},
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Attributes: parentOverridesFromMainDevfile,
										Name:       "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{},
									},
								},
								Projects: []v1.Project{
									{
										Attributes: parentOverridesFromMainDevfile,
										ClonePath:  "/projects",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "handle a parent's data without any local override and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Commands: []v1.Command{
										{
											Id: "devbuild",
											CommandUnion: v1.CommandUnion{
												Exec: &v1.ExecCommand{
													WorkingDir: "/projects/nodejs-starter",
												},
											},
										},
									},
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
									Events: &v1.Events{
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
											PostStop: []string{"post-stop"},
										},
									},
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: parentDevfile,
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Attributes: importFromUri1,
										Id:         "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects",
											},
										},
									},
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Attributes: importFromUri1,
										Name:       "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-10",
												},
											},
										},
									},
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{},
									},
								},
								Projects: []v1.Project{
									{
										Attributes: importFromUri1,
										ClonePath:  "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "it should error out when the override is invalid",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								Parent: &v1.Parent{
									ParentOverrides: v1.ParentOverrides{
										Commands: []v1.CommandParentOverride{
											{
												Id: "devrun",
												CommandUnionParentOverride: v1.CommandUnionParentOverride{
													Exec: &v1.ExecCommandParentOverride{
														WorkingDir: "/projects/nodejs-starter",
													},
												},
											},
										},
										Components: []v1.ComponentParentOverride{
											{
												Name: "nodejs",
												ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
													Container: &v1.ContainerComponentParentOverride{
														ContainerParentOverride: v1.ContainerParentOverride{
															Image: "quay.io/nodejs-12",
														},
													},
												},
											},
										},
										Projects: []v1.ProjectParentOverride{
											{
												ClonePath: "/projects",
												Name:      "nodejs-starter",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands:   []v1.Command{},
								Components: []v1.Component{},
								Projects:   []v1.Project{},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{newCmpErr, newCmdErr, newProjectErr},
		},
		{
			name: "error out if the same parent command is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Commands: []v1.Command{
										{
											Id: "devbuild",
											CommandUnion: v1.CommandUnion{
												Exec: &v1.ExecCommand{
													WorkingDir: "/projects/nodejs-starter",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{parentCmdAlreadyDefinedErr},
		},
		{
			name: "error out if the same parent component is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{parentCmpAlreadyDefinedErr},
		},
		{
			name: "should not have error if the same event is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Events: &v1.Events{
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
											PostStop: []string{"post-stop"},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStop: []string{"post-stop"},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStop:  []string{"post-stop"},
										PreStart:  []string{},
										PreStop:   []string{},
										PostStart: []string{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "error out if the parent project is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
											ProjectSource: v1.ProjectSource{
												Git: &v1.GitProjectSource{
													GitLikeProjectSource: v1.GitLikeProjectSource{
														Remotes: map[string]string{
															"origin": "url",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"origin": "url",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{parentProjectAlreadyDefinedErr},
		},
		{
			name: "it should merge the plugin's uri data and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Commands: []v1.Command{
										{
											Id: "devbuild",
											CommandUnion: v1.CommandUnion{
												Exec: &v1.ExecCommand{
													WorkingDir: "/projects/nodejs-starter",
												},
											},
										},
									},
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
									Events: &v1.Events{
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
											PostStop: []string{"post-stop"},
										},
									},
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
										},
									},
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir:  "/projects",
												CommandLine: "npm run",
											},
										},
									},
								},
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
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Attributes: importFromUri2,
										Id:         "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects",
											},
										},
									},
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Attributes: importFromUri2,
										Name:       "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-10",
												},
											},
										},
									},
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{},
									},
								},
								Projects: []v1.Project{
									{
										Attributes: importFromUri2,
										ClonePath:  "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "it should override the plugin's data with local overrides and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Commands: []v1.Command{
										{
											Id: "devbuild",
											CommandUnion: v1.CommandUnion{
												Exec: &v1.ExecCommand{
													WorkingDir: "/projects/nodejs-starter",
												},
											},
										},
									},
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
									Events: &v1.Events{
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
											PostStop: []string{"post-stop-1"},
										},
									},
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
										},
									},
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir:  "/projects",
												CommandLine: "npm run",
											},
										},
									},
								},
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
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop-2"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
								},
							},
						},
					},
				},
			},
			pluginOverride: v1.PluginOverrides{
				OverridesBase: v1.OverridesBase{},
				Components: []v1.ComponentPluginOverride{
					{
						Name: "nodejs",
						ComponentUnionPluginOverride: v1.ComponentUnionPluginOverride{
							Container: &v1.ContainerComponentPluginOverride{
								ContainerPluginOverride: v1.ContainerPluginOverride{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
				},
				Commands: []v1.CommandPluginOverride{
					{
						Id: "devrun",
						CommandUnionPluginOverride: v1.CommandUnionPluginOverride{
							Exec: &v1.ExecCommandPluginOverride{
								WorkingDir:  "/projects-new",
								CommandLine: "npm build",
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Attributes: pluginOverridesFromMainDevfile,
										Id:         "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm build",
												WorkingDir:  "/projects-new",
											},
										},
									},
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Attributes: pluginOverridesFromMainDevfile,
										Name:       "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop-1", "post-stop-2"},
										PreStop:   []string{},
										PreStart:  []string{},
									},
								},
								Projects: []v1.Project{
									{
										Attributes: importFromUri2,
										ClonePath:  "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "it should error out when the plugin devfile is invalid",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands:   []v1.Command{},
								Components: []v1.Component{},
								Projects:   []v1.Project{},
							},
						},
					},
				},
			},
			pluginOverride: v1.PluginOverrides{
				Commands: []v1.CommandPluginOverride{
					{
						Id: "devrun",
						CommandUnionPluginOverride: v1.CommandUnionPluginOverride{
							Exec: &v1.ExecCommandPluginOverride{
								WorkingDir: "/projects/nodejs-starter",
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{newCmdErr},
		},
		{
			name: "error out if the same plugin command is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Commands: []v1.Command{
										{
											Id: "devbuild",
											CommandUnion: v1.CommandUnion{
												Exec: &v1.ExecCommand{
													WorkingDir: "/projects/nodejs-starter",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{pluginCmdAlreadyDefinedErr},
		},
		{
			name: "error out if the same plugin component is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{pluginCmpAlreadyDefinedErr},
		},
		{
			name: "error out if the plugin project is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
											ProjectSource: v1.ProjectSource{
												Git: &v1.GitProjectSource{
													GitLikeProjectSource: v1.GitLikeProjectSource{
														Remotes: map[string]string{
															"origin": "url",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"origin": "url",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{pluginProjectAlreadyDefinedErr},
		},
		{
			name: "error out if the same project is defined in the both plugin devfile and parent",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
											ProjectSource: v1.ProjectSource{
												Git: &v1.GitProjectSource{
													GitLikeProjectSource: v1.GitLikeProjectSource{
														Remotes: map[string]string{
															"origin": "url",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"origin": "url",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"origin": "url",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{pluginProjectAlreadyDefinedErr},
		},
		{
			name: "error out if the same command is defined in both plugin devfile and parent devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Commands: []v1.Command{
										{
											Id: "devbuild",
											CommandUnion: v1.CommandUnion{
												Exec: &v1.ExecCommand{
													WorkingDir: "/projects/nodejs-starter",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{pluginCmdAlreadyDefinedErr},
		},
		{
			name: "error out if the same component is defined in both plugin devfile and parent devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "build",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "build",
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
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{pluginCmpAlreadyDefinedErr},
		},
		{
			name: "it should override the requested parent's data and plugin's data, and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								Parent: &v1.Parent{
									ParentOverrides: v1.ParentOverrides{
										Commands: []v1.CommandParentOverride{
											{
												Id: "devrun",
												CommandUnionParentOverride: v1.CommandUnionParentOverride{
													Exec: &v1.ExecCommandParentOverride{
														WorkingDir: "/projects/nodejs-starter",
													},
												},
											},
										},
										Projects: []v1.ProjectParentOverride{
											{
												ClonePath: "/projects",
												Name:      "nodejs-starter",
											},
										},
									},
								},
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Commands: []v1.Command{
										{
											Id: "devbuild",
											CommandUnion: v1.CommandUnion{
												Exec: &v1.ExecCommand{
													WorkingDir: "/projects/nodejs-starter",
												},
											},
										},
									},
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
									Events: &v1.Events{
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
											PostStop: []string{"post-stop"},
										},
									},
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir:  "/projects",
												CommandLine: "npm run",
											},
										},
									},
								},
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devdebug",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir:  "/projects",
												CommandLine: "npm debug",
											},
										},
									},
								},
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
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PreStart: []string{"pre-start-0"},
									},
								},
							},
						},
					},
				},
			},
			pluginOverride: v1.PluginOverrides{
				Components: []v1.ComponentPluginOverride{
					{
						Name: "nodejs",
						ComponentUnionPluginOverride: v1.ComponentUnionPluginOverride{
							Container: &v1.ContainerComponentPluginOverride{
								ContainerPluginOverride: v1.ContainerPluginOverride{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Attributes: parentOverridesFromMainDevfile,
										Id:         "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects/nodejs-starter",
											},
										},
									},
									{
										Attributes: importFromUri2,
										Id:         "devdebug",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir:  "/projects",
												CommandLine: "npm debug",
											},
										},
									},
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Attributes: pluginOverridesFromMainDevfile,
										Name:       "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{"pre-start-0"},
									},
								},
								Projects: []v1.Project{
									{
										Attributes: parentOverridesFromMainDevfile,
										ClonePath:  "/projects",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "error out if the plugin component is defined with a different component type in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{pluginCmpAlreadyDefinedErr},
		},
		{
			name: "it should override with no errors if the plugin component is defined with a different component type in the plugin override",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{},
					},
				},
			},
			pluginOverride: v1.PluginOverrides{
				Components: []v1.ComponentPluginOverride{
					{
						Name: "runtime",
						ComponentUnionPluginOverride: v1.ComponentUnionPluginOverride{
							Container: &v1.ContainerComponentPluginOverride{
								ContainerPluginOverride: v1.ContainerPluginOverride{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Attributes: pluginOverridesFromMainDevfile,
										Name:       "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "error out if the parent component is defined with a different component type in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{parentCmpAlreadyDefinedErr},
		},
		{
			name: "it should override with no errors if the parent component is defined with a different component type in the parent override",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								Parent: &v1.Parent{
									ParentOverrides: v1.ParentOverrides{
										Components: []v1.ComponentParentOverride{
											{
												Name: "runtime",
												ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
													Container: &v1.ContainerComponentParentOverride{
														ContainerParentOverride: v1.ContainerParentOverride{
															Image: "quay.io/nodejs-12",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Attributes: parentOverridesFromMainDevfile,
										Name:       "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "error out if the URI is recursively referenced",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: v1.ImportReference{
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Uri: "http://" + uri2,
									},
								},
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr:                []string{importCycleErr},
			testRecursiveReference: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parentTestServer *httptest.Server
			var pluginTestServer *httptest.Server
			if !reflect.DeepEqual(tt.parentDevfile, DevfileObj{}) {
				parentTestServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					data, err := yaml.Marshal(tt.parentDevfile.Data)
					if err != nil {
						t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while doing yaml marshal: %v", err)
					}
					_, err = w.Write(data)
					if err != nil {
						t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while writing data: %v", err)
					}
				}))
				// create a listener with the desired port.
				l1, err := net.Listen("tcp", uri1)
				if err != nil {
					t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while creating listener: %v", err)
				}

				// NewUnstartedServer creates a listener. Close that listener and replace
				// with the one we created.
				parentTestServer.Listener.Close()
				parentTestServer.Listener = l1

				parentTestServer.Start()
				defer parentTestServer.Close()

				parent := tt.args.devFileObj.Data.GetParent()
				if parent == nil {
					parent = &v1.Parent{}
				}
				parent.Uri = parentTestServer.URL

				tt.args.devFileObj.Data.SetParent(parent)
			}
			if !reflect.DeepEqual(tt.pluginDevfile, DevfileObj{}) {

				pluginTestServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					data, err := yaml.Marshal(tt.pluginDevfile.Data)
					if err != nil {
						t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while doing yaml marshal: %v", err)
					}
					_, err = w.Write(data)
					if err != nil {
						t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while writing data: %v", err)
					}
				}))
				l, err := net.Listen("tcp", uri2)
				if err != nil {
					t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while creating listener: %v", err)
				}

				// NewUnstartedServer creates a listener. Close that listener and replace
				// with the one we created.
				pluginTestServer.Listener.Close()
				pluginTestServer.Listener = l

				pluginTestServer.Start()
				defer pluginTestServer.Close()

				plugincomp := []v1.Component{
					{
						Name: "plugincomp",
						ComponentUnion: v1.ComponentUnion{
							Plugin: &v1.PluginComponent{
								ImportReference: v1.ImportReference{
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Uri: pluginTestServer.URL,
									},
								},
								PluginOverrides: tt.pluginOverride,
							},
						},
					},
				}
				tt.args.devFileObj.Data.AddComponents(plugincomp)

			}
			err := parseParentAndPlugin(tt.args.devFileObj, &resolutionContextTree{}, resolverTools{})

			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseParentAndPluginFromURI() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(tt.args.devFileObj.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseParentAndPluginFromURI() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile.Data, tt.args.devFileObj.Data, pretty.Compare(tt.args.devFileObj.Data, tt.wantDevFile.Data))
			} else if err != nil {
				for _, wantErr := range tt.wantErr {
					assert.Regexp(t, wantErr, err.Error(), "Test_parseParentAndPluginFromURI(): Error message should match")
				}
			}
		})
	}
}

func Test_parseParentAndPlugin_RecursivelyReference(t *testing.T) {
	const uri1 = "127.0.0.1:8080"
	const uri2 = "127.0.0.1:8090"
	const httpPrefix = "http://"
	const name = "testcrd"
	const namespace = "defaultnamespace"

	devFileObj := DevfileObj{
		Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					Parent: &v1.Parent{
						ImportReference: v1.ImportReference{
							ImportReferenceUnion: v1.ImportReferenceUnion{
								Uri: httpPrefix + uri1,
							},
						},
					},
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime2",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size: "500Mi",
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

	parentDevfile1 := DevfileObj{
		Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					Parent: &v1.Parent{
						ImportReference: v1.ImportReference{
							ImportReferenceUnion: v1.ImportReferenceUnion{
								Kubernetes: &v1.KubernetesCustomResourceImportReference{
									Name:      name,
									Namespace: namespace,
								},
							},
						},
					},
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size: "500Mi",
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

	parentDevfile2 := DevfileObj{
		Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					Parent: &v1.Parent{
						ImportReference: v1.ImportReference{
							ImportReferenceUnion: v1.ImportReferenceUnion{
								Uri: httpPrefix + uri1,
							},
						},
					},
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "test",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size: "500Mi",
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

	testServer1 := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := yaml.Marshal(parentDevfile1.Data)
		if err != nil {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while doing yaml marshal: %v", err)
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l1, err := net.Listen("tcp", uri1)
	if err != nil {
		t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while creating listener: %v", err)
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer1.Listener.Close()
	testServer1.Listener = l1

	testServer1.Start()
	defer testServer1.Close()

	testServer2 := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		if strings.Contains(r.URL.Path, "/devfiles/nodejs") {
			data, err = yaml.Marshal(parentDevfile2.Data)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while writing data: %v", err)
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l3, err := net.Listen("tcp", uri2)
	if err != nil {
		t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while creating listener: %v", err)
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer2.Listener.Close()
	testServer2.Listener = l3

	testServer2.Start()
	defer testServer2.Close()

	parentSpec := v1.DevWorkspaceTemplateSpec{
		Parent: &v1.Parent{
			ImportReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: "nodejs",
				},
				RegistryUrl: httpPrefix + uri2,
			},
		},
		DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
			Components: []v1.Component{
				{
					Name: "crdcomponent",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{
							Volume: v1.Volume{
								Size: "500Mi",
							},
						},
					},
				},
			},
		},
	}
	devWorkspaceResources := map[string]v1.DevWorkspaceTemplate{
		name: {
			TypeMeta: kubev1.TypeMeta{
				Kind:       "DevWorkspaceTemplate",
				APIVersion: "testgroup/v1alpha2",
			},
			Spec: parentSpec,
		},
	}

	t.Run("it should error out if import reference has a cycle", func(t *testing.T) {
		testK8sClient := &testingutil.FakeK8sClient{
			DevWorkspaceResources: devWorkspaceResources,
		}
		tool := resolverTools{
			k8sClient: testK8sClient,
			context:   context.Background(),
		}

		err := parseParentAndPlugin(devFileObj, &resolutionContextTree{}, tool)
		// devfile has a cycle in references: main devfile -> uri: http://127.0.0.1:8080 -> name: testcrd, namespace: defaultnamespace -> id: nodejs, registryURL: http://127.0.0.1:8090 -> uri: http://127.0.0.1:8080
		expectedErr := fmt.Sprintf("devfile has an cycle in references: main devfile -> uri: %s%s -> name: %s, namespace: %s -> id: nodejs, registryURL: %s%s -> uri: %s%s", httpPrefix, uri1, name, namespace,
			httpPrefix, uri2, httpPrefix, uri1)
		// Unexpected error
		if err == nil || !reflect.DeepEqual(expectedErr, err.Error()) {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error: %v", err)

			return
		}

	})
}

func Test_parseParentFromRegistry(t *testing.T) {
	const validRegistry = "127.0.0.1:8080"
	const invalidRegistry = "invalid-registry.io"
	tool := resolverTools{
		registryURLs: []string{"http://" + validRegistry},
	}

	invalidURLErr := "the provided registryURL: .* is not a valid URL"
	idNotFoundErr := "failed to get id: .* from registry URLs provided"

	parentDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "parent-runtime",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size: "500Mi",
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
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var data []byte
		var err error
		if strings.Contains(r.URL.Path, "/devfiles/nodejs") {
			data, err = yaml.Marshal(parentDevfile.Data)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			t.Errorf("Test_parseParentFromRegistry() unexpected error while doing yaml marshal: %v", err)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Test_parseParentFromRegistry() unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", validRegistry)
	if err != nil {
		t.Errorf("Test_parseParentFromRegistry() unexpected error while creating listener: %v", err)
		return
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	mainDevfileContent := v1.Devfile{
		DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
			Parent: &v1.Parent{
				ImportReference: v1.ImportReference{
					RegistryUrl: "http://" + validRegistry,
					ImportReferenceUnion: v1.ImportReferenceUnion{
						Id: "nodejs",
					},
				},
				ParentOverrides: v1.ParentOverrides{
					Components: []v1.ComponentParentOverride{
						{
							Name: "parent-runtime",
							ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
								Container: &v1.ContainerComponentParentOverride{
									ContainerParentOverride: v1.ContainerParentOverride{
										Image: "quay.io/nodejs-12",
									},
								},
							},
						},
					},
				},
			},
			DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
				Commands: []v1.Command{
					{
						Id: "devbuild",
						CommandUnion: v1.CommandUnion{
							Exec: &v1.ExecCommand{
								WorkingDir: "/projects/nodejs-starter",
							},
						},
					},
				},
				Components: []v1.Component{
					{
						Name: "runtime2",
						ComponentUnion: v1.ComponentUnion{
							Container: &v1.ContainerComponent{
								Container: v1.Container{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
				},
				Events: &v1.Events{
					DevWorkspaceEvents: v1.DevWorkspaceEvents{
						PostStop: []string{"post-stop"},
					},
				},
				Projects: []v1.Project{
					{
						ClonePath: "/projects",
						Name:      "nodejs-starter-build",
					},
				},
			},
		},
	}

	importFromRegistry := attributes.Attributes{}.PutString(importSourceAttribute, resolveImportReference(mainDevfileContent.Parent.ImportReference))
	parentOverridesFromMainDevfile := attributes.Attributes{}.PutString(importSourceAttribute,
		resolveImportReference(mainDevfileContent.Parent.ImportReference)).PutString(parentOverrideAttribute, "main devfile")

	wantDevfileContent := v1.Devfile{
		DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
			DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
				Commands: []v1.Command{
					{
						Id: "devbuild",
						CommandUnion: v1.CommandUnion{
							Exec: &v1.ExecCommand{
								WorkingDir: "/projects/nodejs-starter",
							},
						},
					},
				},
				Components: []v1.Component{
					{
						Attributes: parentOverridesFromMainDevfile,
						Name:       "parent-runtime",
						ComponentUnion: v1.ComponentUnion{
							Container: &v1.ContainerComponent{
								Container: v1.Container{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
					{
						Name: "runtime2",
						ComponentUnion: v1.ComponentUnion{
							Container: &v1.ContainerComponent{
								Container: v1.Container{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
				},
				Events: &v1.Events{
					DevWorkspaceEvents: v1.DevWorkspaceEvents{
						PostStart: []string{},
						PostStop:  []string{"post-stop"},
						PreStop:   []string{},
						PreStart:  []string{},
					},
				},
				Projects: []v1.Project{
					{
						ClonePath: "/projects",
						Name:      "nodejs-starter-build",
					},
				},
			},
		},
	}

	tests := []struct {
		name                   string
		mainDevfile            DevfileObj
		registryURI            string
		wantDevFile            DevfileObj
		wantErr                *string
		testRecursiveReference bool
	}{
		{
			name: "it should override the requested parent's data from provided registryURL and add the local devfile's data",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: mainDevfileContent,
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: wantDevfileContent,
				},
			},
		},
		{
			name: "it should override the requested parent's data from registryURLs set in context and add the local devfile's data",
			mainDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: mainDevfileContent,
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: wantDevfileContent,
				},
			},
		},
		{
			name: "it should merge the requested parent's data from provided registryURL if no override is set",
			mainDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: v1.ImportReference{
									RegistryUrl: "http://" + validRegistry,
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Id: "nodejs",
									},
								},
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Name: "runtime2",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Attributes: importFromRegistry,
										Name:       "parent-runtime",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
												},
											},
										},
									},
									{
										Name: "runtime2",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "it should error out with invalid registry provided",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: v1.ImportReference{
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Id: "nodejs",
									},
									RegistryUrl: invalidRegistry,
								},
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{},
						},
					},
				},
			},
			wantErr: &invalidURLErr,
		},
		{
			name: "it should error out with non-exist registry id provided",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: v1.ImportReference{
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Id: "not-exist",
									},
								},
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{},
						},
					},
				},
			},
			wantErr: &idNotFoundErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := parseParentAndPlugin(tt.mainDevfile, &resolutionContextTree{}, tool)

			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseParentFromRegistry() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(tt.mainDevfile.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseParentFromRegistry() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile.Data, tt.mainDevfile.Data, pretty.Compare(tt.mainDevfile.Data, tt.wantDevFile.Data))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Test_parseParentFromRegistry(): Error message should match")
			}

		})
	}
}

func Test_parseParentFromKubeCRD(t *testing.T) {

	const (
		namespace  = "default"
		name       = "test-parent-k8s"
		apiVersion = "testgroup/v1alpha2"
	)

	kubeCRDReference := v1.ImportReference{
		ImportReferenceUnion: v1.ImportReferenceUnion{
			Kubernetes: &v1.KubernetesCustomResourceImportReference{
				Name:      name,
				Namespace: namespace,
			},
		},
	}

	importFromKubeCRD := attributes.Attributes{}.PutString(importSourceAttribute, resolveImportReference(kubeCRDReference))
	parentOverridesFromMainDevfile := attributes.Attributes{}.PutString(importSourceAttribute,
		resolveImportReference(kubeCRDReference)).PutString(parentOverrideAttribute, "main devfile")

	parentSpec := v1.DevWorkspaceTemplateSpec{
		DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
			Components: []v1.Component{
				{
					Name: "parent-runtime",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{
							Volume: v1.Volume{
								Size: "500Mi",
							},
						},
					},
				},
			},
		},
	}

	crdNotFoundErr := "not found"

	tests := []struct {
		name                  string
		devWorkspaceResources map[string]v1.DevWorkspaceTemplate
		errors                map[string]string
		mainDevfile           DevfileObj
		wantDevFile           DevfileObj
		wantErr               *string
	}{
		{
			name: "should successfully override the parent data",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: kubeCRDReference,
								ParentOverrides: v1.ParentOverrides{
									Components: []v1.ComponentParentOverride{
										{
											Name: "parent-runtime",
											ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
												Container: &v1.ContainerComponentParentOverride{
													ContainerParentOverride: v1.ContainerParentOverride{
														Image: "quay.io/nodejs-12",
													},
												},
											},
										},
									},
								},
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Attributes: parentOverridesFromMainDevfile,
										Name:       "parent-runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			devWorkspaceResources: map[string]v1.DevWorkspaceTemplate{
				name: {
					TypeMeta: kubev1.TypeMeta{
						Kind:       "DevWorkspaceTemplate",
						APIVersion: apiVersion,
					},
					Spec: parentSpec,
				},
			},
		},
		{
			name: "should successfully merge the parent data without override defined",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: kubeCRDReference,
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Attributes: importFromKubeCRD,
										Name:       "parent-runtime",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
												},
											},
										},
									},
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			devWorkspaceResources: map[string]v1.DevWorkspaceTemplate{
				name: {
					TypeMeta: kubev1.TypeMeta{
						Kind:       "DevWorkspaceTemplate",
						APIVersion: apiVersion,
					},
					Spec: parentSpec,
				},
			},
		},
		{
			name: "should fail if kclient get returns error",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: kubeCRDReference,
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{},
						},
					},
				},
			},
			devWorkspaceResources: map[string]v1.DevWorkspaceTemplate{},
			errors: map[string]string{
				name: crdNotFoundErr,
			},
			wantErr: &crdNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testK8sClient := &testingutil.FakeK8sClient{
				DevWorkspaceResources: tt.devWorkspaceResources,
				Errors:                tt.errors,
			}
			tool := resolverTools{
				k8sClient: testK8sClient,
				context:   context.Background(),
			}
			err := parseParentAndPlugin(tt.mainDevfile, &resolutionContextTree{}, tool)

			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseParentFromKubeCRD() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(tt.mainDevfile.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseParentFromKubeCRD() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile.Data, tt.mainDevfile.Data, pretty.Compare(tt.mainDevfile.Data, tt.wantDevFile.Data))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Test_parseParentFromKubeCRD(): Error message should match")
			}

		})
	}
}

func Test_parseFromURI(t *testing.T) {
	const (
		uri1             = "127.0.0.1:8080"
		httpPrefix       = "http://"
		localRelativeURI = "testTmp/dir/devfile.yaml"
		notExistURI      = "notexist/devfile.yaml"
		invalidURL       = "http//invalid.com"
	)
	uri2 := path.Join(uri1, localRelativeURI)

	localDevfile := DevfileObj{
		Ctx: devfileCtx.NewDevfileCtx(localRelativeURI),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime",
								ComponentUnion: v1.ComponentUnion{
									Container: &v1.ContainerComponent{
										Container: v1.Container{
											Image: "nodejs",
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

	invalidFilePathErr := "the provided path is not a valid yaml filepath, and devfile.yaml or .devfile.yaml not found in the provided path.*"
	readDevfileErr := "failed to read devfile from path.*"
	URLNotFoundErr := "error getting devfile info from url: failed to retrieve .*, 404: Not Found"
	invalidURLErr := "parse .* invalid URI for request"

	// prepare for local file
	err := os.MkdirAll(path.Dir(localRelativeURI), 0755)
	if err != nil {
		t.Errorf("Test_parseFromURI() error: failed to create folder: %v, error: %v", path.Dir(localRelativeURI), err)
		return
	}
	yamlData, err := yaml.Marshal(localDevfile.Data)
	if err != nil {
		t.Errorf("Test_parseFromURI() error: failed to marshall devfile data: %v", err)
		return
	}
	err = ioutil.WriteFile(localRelativeURI, yamlData, 0644)
	if err != nil {
		t.Errorf("Test_parseFromURI() error: fail to write to file: %v", err)
		return
	}
	defer os.RemoveAll("testTmp/")

	parentDevfile := DevfileObj{
		Ctx: devfileCtx.NewURLDevfileCtx(httpPrefix + uri1),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					Parent: &v1.Parent{
						ImportReference: v1.ImportReference{
							ImportReferenceUnion: v1.ImportReferenceUnion{
								Uri: localRelativeURI,
							},
						},
					},
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime2",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size: "500Mi",
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
	relativeParentDevfile := DevfileObj{
		Ctx: devfileCtx.NewURLDevfileCtx(httpPrefix + uri2),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size: "500Mi",
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

	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "notexist") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var data []byte
		var err error
		if strings.Contains(r.URL.Path, "devfile.yaml") {
			data, err = yaml.Marshal(relativeParentDevfile.Data)
		} else {
			data, err = yaml.Marshal(parentDevfile.Data)
		}
		if err != nil {
			t.Errorf("Test_parseFromURI() unexpected while doing yaml marshal: %v", err)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Test_parseFromURI() unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", uri1)
	if err != nil {
		t.Errorf("Test_parseFromURI() unexpected error while creating listener: %v", err)
		return
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	tests := []struct {
		name            string
		curDevfileCtx   devfileCtx.DevfileCtx
		importReference v1.ImportReference
		wantDevFile     DevfileObj
		wantErr         *string
	}{
		{
			name:          "should be able to parse from relative uri on local disk",
			curDevfileCtx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
			wantDevFile:   localDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: localRelativeURI,
				},
			},
		},
		{
			name:          "should be able to parse relative uri from URL",
			curDevfileCtx: parentDevfile.Ctx,
			wantDevFile:   relativeParentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: localRelativeURI,
				},
			},
		},
		{
			name:          "should fail if no path or url has been set for devfile ctx",
			curDevfileCtx: devfileCtx.DevfileCtx{},
			wantErr:       &invalidFilePathErr,
		},
		{
			name:          "should fail if file not exist",
			curDevfileCtx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: notExistURI,
				},
			},
			wantErr: &readDevfileErr,
		},
		{
			name:          "should fail if url not exist",
			curDevfileCtx: devfileCtx.NewURLDevfileCtx(httpPrefix + uri1),
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: notExistURI,
				},
			},
			wantErr: &URLNotFoundErr,
		},
		{
			name:          "should fail if with invalid URI format",
			curDevfileCtx: devfileCtx.NewURLDevfileCtx(OutputDevfileYamlPath),
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: invalidURL,
				},
			},
			wantErr: &invalidURLErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if the main devfile is from local, need to set absolute path
			if tt.curDevfileCtx.GetURL() == "" {
				err := tt.curDevfileCtx.SetAbsPath()
				if err != nil {
					t.Errorf("Test_parseFromURI() unexpected error: %v", err)
					return
				}
			}
			got, err := parseFromURI(tt.importReference, tt.curDevfileCtx, &resolutionContextTree{}, resolverTools{})
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseFromURI() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(got.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseFromURI() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, got, pretty.Compare(tt.wantDevFile, got))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Test_parseFromURI(): Error message should match")
			}
		})
	}
}

func Test_parseFromRegistry(t *testing.T) {
	const (
		registry        = "127.0.0.1:8080"
		httpPrefix      = "http://"
		notExistId      = "notexist"
		invalidRegistry = "http//invalid.com"
		registryId      = "nodejs"
	)

	parentDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime2",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size: "500Mi",
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

	invalidURLErr := "the provided registryURL: .* is not a valid URL"
	URLNotFoundErr := "failed to retrieve .*, 404: Not Found"
	missingRegistryURLErr := "failed to fetch from registry, registry URL is not provided"
	invalidRegistryURLErr := "Get .* dial tcp: lookup http: .*"

	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		var err error
		if strings.Contains(r.URL.Path, "/devfiles/"+registryId) {
			data, err = yaml.Marshal(parentDevfile.Data)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			t.Errorf("Test_parseFromRegistry() unexpected error while doing yaml marshal: %v", err)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Test_parseFromRegistry() unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", registry)
	if err != nil {
		t.Errorf("Test_parseFromRegistry() unexpected error while creating listener: %v", err)
		return
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	tests := []struct {
		name            string
		curDevfileCtx   devfileCtx.DevfileCtx
		importReference v1.ImportReference
		tool            resolverTools
		wantDevFile     DevfileObj
		wantErr         *string
	}{
		{
			name:        "should fail if provided registryUrl does not have protocol prefix",
			wantDevFile: parentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
				RegistryUrl: registry,
			},
			wantErr: &invalidURLErr,
		},
		{
			name:        "should be able to parse from provided registryUrl with prefix",
			wantDevFile: parentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
				RegistryUrl: httpPrefix + registry,
			},
		},
		{
			name:        "should be able to parse from registry URL defined in tool",
			wantDevFile: parentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
			},
			tool: resolverTools{
				registryURLs: []string{"http://" + registry},
			},
		},
		{
			name: "should fail if registryId does not exist",
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: notExistId,
				},
				RegistryUrl: httpPrefix + registry,
			},
			wantErr: &URLNotFoundErr,
		},
		{
			name: "should fail if registryUrl is not provided, and no registry URLs has been set in tool",
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
			},
			wantErr: &missingRegistryURLErr,
		},
		{
			name: "should fail if registryUrl is invalid",
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: notExistId,
				},
				RegistryUrl: httpPrefix + invalidRegistry,
			},
			wantErr: &invalidRegistryURLErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFromRegistry(tt.importReference, &resolutionContextTree{}, tt.tool)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseFromRegistry() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(got.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseFromRegistry() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, got, pretty.Compare(tt.wantDevFile, got))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Test_parseFromRegistry(): Error message should match")
			}
		})
	}
}

func Test_parseFromKubeCRD(t *testing.T) {
	const (
		namespace  = "default"
		name       = "test-parent-k8s"
		apiVersion = "testgroup/v1alpha2"
	)
	parentSpec := v1.DevWorkspaceTemplateSpec{
		DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
			Components: []v1.Component{
				{
					Name: "runtime",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{
							Volume: v1.Volume{
								Size: "500Mi",
							},
						},
					},
				},
			},
		},
	}
	parentDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevWorkspaceTemplateSpec: parentSpec,
			},
		},
	}

	crdNotFoundErr := "not found"

	tests := []struct {
		name                  string
		curDevfileCtx         devfileCtx.DevfileCtx
		importReference       v1.ImportReference
		devWorkspaceResources map[string]v1.DevWorkspaceTemplate
		errors                map[string]string
		wantDevFile           DevfileObj
		wantErr               *string
	}{
		{
			name:        "should successfully parse the parent with namespace specified in devfile",
			wantDevFile: parentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Kubernetes: &v1.KubernetesCustomResourceImportReference{
						Name:      name,
						Namespace: namespace,
					},
				},
			},
			devWorkspaceResources: map[string]v1.DevWorkspaceTemplate{
				name: {
					TypeMeta: kubev1.TypeMeta{
						Kind:       "DevWorkspaceTemplate",
						APIVersion: apiVersion,
					},
					Spec: parentSpec,
				},
			},
		},
		{
			name:        "should fail if kclient get returns error",
			wantDevFile: parentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Kubernetes: &v1.KubernetesCustomResourceImportReference{
						Name:      name,
						Namespace: namespace,
					},
				},
			},
			devWorkspaceResources: map[string]v1.DevWorkspaceTemplate{},
			errors: map[string]string{
				name: crdNotFoundErr,
			},
			wantErr: &crdNotFoundErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testK8sClient := &testingutil.FakeK8sClient{
				DevWorkspaceResources: tt.devWorkspaceResources,
				Errors:                tt.errors,
			}
			tool := resolverTools{
				k8sClient: testK8sClient,
				context:   context.Background(),
			}
			got, err := parseFromKubeCRD(tt.importReference, &resolutionContextTree{}, tool)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseFromKubeCRD() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(got.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseFromKubeCRD() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, got, pretty.Compare(tt.wantDevFile, got))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Test_parseFromKubeCRD(): Error message should match")
			}
		})
	}
}
