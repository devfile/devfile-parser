package common

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func TestGetGroup(t *testing.T) {

	tests := []struct {
		name    string
		command v1.Command
		want    *v1.CommandGroup
	}{
		{
			name: "Exec command group",
			command: v1.Command{
				Id: "exec1",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{
						LabeledCommand: v1.LabeledCommand{
							BaseCommand: v1.BaseCommand{
								Group: &v1.CommandGroup{
									IsDefault: true,
									Kind:      v1.RunCommandGroupKind,
								},
							},
						},
					},
				},
			},
			want: &v1.CommandGroup{
				IsDefault: true,
				Kind:      v1.RunCommandGroupKind,
			},
		},
		{
			name: "Composite command group",
			command: v1.Command{
				Id: "composite1",
				CommandUnion: v1.CommandUnion{
					Composite: &v1.CompositeCommand{
						LabeledCommand: v1.LabeledCommand{
							BaseCommand: v1.BaseCommand{
								Group: &v1.CommandGroup{
									IsDefault: true,
									Kind:      v1.BuildCommandGroupKind,
								},
							},
						},
					},
				},
			},
			want: &v1.CommandGroup{
				IsDefault: true,
				Kind:      v1.BuildCommandGroupKind,
			},
		},
		{
			name:    "Empty command",
			command: v1.Command{},
			want:    nil,
		},
		{
			name: "Apply command group",
			command: v1.Command{
				Id: "apply1",
				CommandUnion: v1.CommandUnion{
					Apply: &v1.ApplyCommand{
						LabeledCommand: v1.LabeledCommand{
							BaseCommand: v1.BaseCommand{
								Group: &v1.CommandGroup{
									IsDefault: true,
									Kind:      v1.TestCommandGroupKind,
								},
							},
						},
					},
				},
			},
			want: &v1.CommandGroup{
				IsDefault: true,
				Kind:      v1.TestCommandGroupKind,
			},
		},
		{
			name: "Custom command group",
			command: v1.Command{
				Id: "custom1",
				CommandUnion: v1.CommandUnion{
					Custom: &v1.CustomCommand{
						LabeledCommand: v1.LabeledCommand{
							BaseCommand: v1.BaseCommand{
								Group: &v1.CommandGroup{
									IsDefault: true,
									Kind:      v1.BuildCommandGroupKind,
								},
							},
						},
					},
				},
			},
			want: &v1.CommandGroup{
				IsDefault: true,
				Kind:      v1.BuildCommandGroupKind,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandGroup := GetGroup(tt.command)
			if !reflect.DeepEqual(commandGroup, tt.want) {
				t.Errorf("TestGetGroup() error: expected %v, actual %v", tt.want, commandGroup)
			}
		})
	}

}

func TestGetExecComponent(t *testing.T) {

	tests := []struct {
		name    string
		command v1.Command
		want    string
	}{
		{
			name: "Exec component present",
			command: v1.Command{
				Id: "exec1",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{
						Component: "component1",
					},
				},
			},
			want: "component1",
		},
		{
			name: "Exec component absent",
			command: v1.Command{
				Id: "exec1",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{},
				},
			},
			want: "",
		},
		{
			name:    "Empty command",
			command: v1.Command{},
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := GetExecComponent(tt.command)
			if component != tt.want {
				t.Errorf("TestGetExecComponent() error: expected %v, actual %v", tt.want, component)
			}
		})
	}

}

func TestGetExecCommandLine(t *testing.T) {

	tests := []struct {
		name    string
		command v1.Command
		want    string
	}{
		{
			name: "Exec command line present",
			command: v1.Command{
				Id: "exec1",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{
						CommandLine: "commandline1",
					},
				},
			},
			want: "commandline1",
		},
		{
			name: "Exec command line absent",
			command: v1.Command{
				Id: "exec1",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{},
				},
			},
			want: "",
		},
		{
			name:    "Empty command",
			command: v1.Command{},
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandLine := GetExecCommandLine(tt.command)
			if commandLine != tt.want {
				t.Errorf("TestGetExecCommandLine() error: expected %v, actual %v", tt.want, commandLine)
			}
		})
	}

}

func TestGetExecWorkingDir(t *testing.T) {

	tests := []struct {
		name    string
		command v1.Command
		want    string
	}{
		{
			name: "Exec working dir present",
			command: v1.Command{
				Id: "exec1",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{
						WorkingDir: "workingdir1",
					},
				},
			},
			want: "workingdir1",
		},
		{
			name: "Exec working dir absent",
			command: v1.Command{
				Id: "exec1",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{},
				},
			},
			want: "",
		},
		{
			name:    "Empty command",
			command: v1.Command{},
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workingDir := GetExecWorkingDir(tt.command)
			if workingDir != tt.want {
				t.Errorf("TestGetExecWorkingDir() error: expected %v, actual %v", tt.want, workingDir)
			}
		})
	}

}

func TestGetCommandType(t *testing.T) {

	cmdTypeErr := "unknown command type"

	tests := []struct {
		name        string
		command     v1.Command
		wantErr     *string
		commandType v1.CommandType
	}{
		{
			name: "Exec command",
			command: v1.Command{
				Id: "exec1",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{},
				},
			},
			commandType: v1.ExecCommandType,
		},
		{
			name: "Composite command",
			command: v1.Command{
				Id: "comp1",
				CommandUnion: v1.CommandUnion{
					Composite: &v1.CompositeCommand{},
				},
			},
			commandType: v1.CompositeCommandType,
		},
		{
			name: "Apply command",
			command: v1.Command{
				Id: "apply1",
				CommandUnion: v1.CommandUnion{
					Apply: &v1.ApplyCommand{},
				},
			},
			commandType: v1.ApplyCommandType,
		},
		{
			name: "Custom command",
			command: v1.Command{
				Id: "custom",
				CommandUnion: v1.CommandUnion{
					Custom: &v1.CustomCommand{},
				},
			},
			commandType: v1.CustomCommandType,
		},
		{
			name: "Unknown command",
			command: v1.Command{
				Id:           "unknown",
				CommandUnion: v1.CommandUnion{},
			},
			wantErr: &cmdTypeErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCommandType(tt.command)
			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestGetCommandType() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && got != tt.commandType {
				t.Errorf("TestGetCommandType() error: command type mismatch, expected: %v got: %v", tt.commandType, got)
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestGetCommandType(): Error message should match")
			}
		})
	}

}

func TestGetCommandsFromEvent(t *testing.T) {

	execCommands := []v1.Command{
		{
			Id: "exec1",
			CommandUnion: v1.CommandUnion{
				Exec: &v1.ExecCommand{},
			},
		},
		{
			Id: "exec2",
			CommandUnion: v1.CommandUnion{
				Exec: &v1.ExecCommand{},
			},
		},
		{
			Id: "exec3",
			CommandUnion: v1.CommandUnion{
				Exec: &v1.ExecCommand{},
			},
		},
	}

	compCommands := []v1.Command{
		{
			Id: "comp1",
			CommandUnion: v1.CommandUnion{
				Composite: &v1.CompositeCommand{
					Commands: []string{
						"exec1",
						"exec3",
					},
				},
			},
		},
	}

	commandsMap := map[string]v1.Command{
		compCommands[0].Id: compCommands[0],
		execCommands[0].Id: execCommands[0],
		execCommands[1].Id: execCommands[1],
		execCommands[2].Id: execCommands[2],
	}

	tests := []struct {
		name         string
		eventName    string
		wantCommands []string
	}{
		{
			name:      "composite event",
			eventName: "comp1",
			wantCommands: []string{
				"exec1",
				"exec3",
			},
		},
		{
			name:      "exec event",
			eventName: "exec2",
			wantCommands: []string{
				"exec2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := GetCommandsFromEvent(commandsMap, tt.eventName)
			if !reflect.DeepEqual(tt.wantCommands, commands) {
				t.Errorf("TestGetCommandsFromEvent() error: got %v expected %v", commands, tt.wantCommands)
			}
		})
	}

}
