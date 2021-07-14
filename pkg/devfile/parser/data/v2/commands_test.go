package v2

import (
	"fmt"
	"github.com/kylelemons/godebug/pretty"
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/stretchr/testify/assert"
)

func TestDevfile200_GetCommands(t *testing.T) {

	invalidCmdTypeErr := "unknown command type"

	tests := []struct {
		name            string
		currentCommands []v1.Command
		filterOptions   common.DevfileOptions
		wantCommands    []string
		wantErr         *string
	}{
		{
			name: "Get all the commands",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
				{
					Id: "command2",
					CommandUnion: v1.CommandUnion{
						Composite: &v1.CompositeCommand{},
					},
				},
			},
			wantCommands: []string{"command1", "command2"},
		},
		{
			name: "Get the filtered commands",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString":  "firstStringValue",
						"secondString": "secondStringValue",
					}),
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{
							LabeledCommand: v1.LabeledCommand{
								BaseCommand: v1.BaseCommand{
									Group: &v1.CommandGroup{
										Kind: v1.BuildCommandGroupKind,
									},
								},
							},
						},
					},
				},
				{
					Id: "command2",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					CommandUnion: v1.CommandUnion{
						Composite: &v1.CompositeCommand{},
					},
				},
				{
					Id: "command3",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					CommandUnion: v1.CommandUnion{
						Composite: &v1.CompositeCommand{
							LabeledCommand: v1.LabeledCommand{
								BaseCommand: v1.BaseCommand{
									Group: &v1.CommandGroup{
										Kind: v1.BuildCommandGroupKind,
									},
								},
							},
						},
					},
				},
				{
					Id: "command4",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"thirdString": "thirdStringValue",
					}),
					CommandUnion: v1.CommandUnion{
						Apply: &v1.ApplyCommand{
							LabeledCommand: v1.LabeledCommand{
								BaseCommand: v1.BaseCommand{
									Group: &v1.CommandGroup{
										Kind: v1.BuildCommandGroupKind,
									},
								},
							},
						},
					},
				},
				{
					Id: "command5",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					CommandUnion: v1.CommandUnion{
						Composite: &v1.CompositeCommand{
							LabeledCommand: v1.LabeledCommand{
								BaseCommand: v1.BaseCommand{
									Group: &v1.CommandGroup{
										Kind: v1.RunCommandGroupKind,
									},
								},
							},
						},
					},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstString": "firstStringValue",
				},
				CommandOptions: common.CommandOptions{
					CommandGroupKind: v1.BuildCommandGroupKind,
					CommandType:      v1.CompositeCommandType,
				},
			},
			wantCommands: []string{"command3"},
		},
		{
			name: "Wrong filter for commands",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString":  "firstStringValue",
						"secondString": "secondStringValue",
					}),
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
				{
					Id: "command2",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					CommandUnion: v1.CommandUnion{
						Composite: &v1.CompositeCommand{},
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
			name: "Invalid command type",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
					}),
					CommandUnion: v1.CommandUnion{},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstString": "firstStringValue",
				},
			},
			wantErr: &invalidCmdTypeErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Commands: tt.currentCommands,
						},
					},
				},
			}

			commands, err := d.GetCommands(tt.filterOptions)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_GetCommands() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				// confirm the length of actual vs expected
				if len(commands) != len(tt.wantCommands) {
					t.Errorf("TestDevfile200_GetCommands() error: length of expected commands is not the same as the length of actual commands")
					return
				}

				// compare the command slices for content
				for _, wantCommand := range tt.wantCommands {
					matched := false
					for _, command := range commands {
						if wantCommand == command.Id {
							matched = true
						}
					}

					if !matched {
						t.Errorf("TestDevfile200_GetCommands() error: command %s not found in the devfile", wantCommand)
					}
				}
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_GetCommands(): Error message should match")
			}
		})
	}
}

func TestDevfile200_AddCommands(t *testing.T) {
	multipleDupError := fmt.Sprintf("%s\n%s", "command command1 already exists in devfile", "command command2 already exists in devfile")

	tests := []struct {
		name            string
		currentCommands []v1.Command
		newCommands     []v1.Command
		wantErr         *string
	}{
		{
			name: "Command does not exist",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
			},
			newCommands: []v1.Command{
				{
					Id: "command2",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
				{
					Id: "command3",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Multiple duplicate commands",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
				{
					Id: "command2",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
			},
			newCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
				{
					Id: "command2",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
				{
					Id: "command3",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
			},
			wantErr: &multipleDupError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Commands: tt.currentCommands,
						},
					},
				},
			}

			err := d.AddCommands(tt.newCommands)
			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_AddCommands() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_AddCommands(): Error message should match")
			} else {
				wantCommands := append(tt.currentCommands, tt.newCommands...)
				if !reflect.DeepEqual(d.Commands, wantCommands) {
					t.Errorf("TestDevfile200_AddCommands() wanted: %v, got: %v, difference at %v", wantCommands, d.Commands, pretty.Compare(wantCommands, d.Commands))
				}
			}

		})
	}
}

func TestDevfile200_UpdateCommands(t *testing.T) {
	invalidCmdErr := "update command failed: command .* not found"

	tests := []struct {
		name            string
		currentCommands []v1.Command
		newCommand      v1.Command
		wantErr         *string
	}{
		{
			name: "successfully update the command",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{
							Component: "component1",
						},
					},
				},
				{
					Id: "command2",
					CommandUnion: v1.CommandUnion{
						Composite: &v1.CompositeCommand{},
					},
				},
			},
			newCommand: v1.Command{
				Id: "command1",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{
						Component: "component1new",
					},
				},
			},
		},
		{
			name: "fail to update the command if not exist",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{
							Component: "component1",
						},
					},
				},
			},
			newCommand: v1.Command{
				Id: "command2",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{
						Component: "component1new",
					},
				},
			},
			wantErr: &invalidCmdErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Commands: tt.currentCommands,
						},
					},
				},
			}

			err := d.UpdateCommand(tt.newCommand)
			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_UpdateCommands() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				commands, err := d.GetCommands(common.DevfileOptions{})
				if err != nil {
					t.Errorf("TestDevfile200_UpdateCommands() unxpected error: %v", err)
					return
				}

				matched := false
				for _, devfileCommand := range commands {
					if tt.newCommand.Id == devfileCommand.Id {
						matched = true
						if !reflect.DeepEqual(devfileCommand, tt.newCommand) {
							t.Errorf("TestDevfile200_UpdateCommands() error: command mismatch, wanted %+v, got %+v", tt.newCommand, devfileCommand)
						}
					}
				}

				if !matched {
					t.Errorf("TestDevfile200_UpdateCommands() error: command mismatch, did not find command with id %s", tt.newCommand.Id)
				}
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_UpdateCommands(): Error message should match")
			}
		})
	}
}

func TestDeleteCommands(t *testing.T) {
	missingCmdErr := "command .* is not found in the devfile"

	d := &DevfileV2{
		v1.Devfile{
			DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
					Commands: []v1.Command{
						{
							Id: "command1",
							CommandUnion: v1.CommandUnion{
								Exec: &v1.ExecCommand{},
							},
						},
						{
							Id: "command2",
							CommandUnion: v1.CommandUnion{
								Exec: &v1.ExecCommand{},
							},
						},
						{
							Id: "command3",
							CommandUnion: v1.CommandUnion{
								Composite: &v1.CompositeCommand{
									Commands: []string{"command1", "command2", "command1"},
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name            string
		commandToDelete string
		wantCommands    []v1.Command
		wantErr         *string
	}{
		{
			name:            "Successfully delete command",
			commandToDelete: "command1",
			wantCommands: []v1.Command{
				{
					Id: "command2",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
				{
					Id: "command3",
					CommandUnion: v1.CommandUnion{
						Composite: &v1.CompositeCommand{
							Commands: []string{"command1", "command2", "command1"},
						},
					},
				},
			},
		},
		{
			name:            "Missing Command",
			commandToDelete: "command34",
			wantErr:         &missingCmdErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := d.DeleteCommand(tt.commandToDelete)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDeleteCommands() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				assert.Equal(t, tt.wantCommands, d.Commands, "TestDeleteCommands(): The two values should be the same.")
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDeleteCommands(): Error message should match")
			}
		})
	}

}
