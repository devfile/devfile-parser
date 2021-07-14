package v2

import (
	"fmt"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func TestDevfile200_AddEvents(t *testing.T) {
	multipleDupError := fmt.Sprintf("%s\n%s", "event field pre start already exists in devfile", "event field post stop already exists in devfile")

	tests := []struct {
		name          string
		currentEvents *v1.Events
		newEvents     v1.Events
		wantEvents    v1.Events
		wantErr       *string
	}{
		{
			name: "successfully add the events",
			currentEvents: &v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart: []string{"preStart1"},
				},
			},
			newEvents: v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PostStart: []string{"postStart1"},
				},
			},
			wantEvents: v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart:  []string{"preStart1"},
					PostStart: []string{"postStart1"},
				},
			},
		},
		{
			name: "successfully add the events to empty devfile event",
			newEvents: v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PostStart: []string{"postStart1"},
				},
			},
			wantEvents: v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PostStart: []string{"postStart1"},
				},
			},
		},
		{
			name: "event already present",
			currentEvents: &v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart: []string{"preStart1"},
					PostStop: []string{"postStop1"},
				},
			},
			newEvents: v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart: []string{"preStart2"},
					PostStop: []string{"postStop2"},
					PreStop:  []string{"preStop"},
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
							Events: tt.currentEvents,
						},
					},
				},
			}

			err := d.AddEvents(tt.newEvents)

			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestDevfile200_AddEvents() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestDevfile200_AddEvents(): Error message should match")
			} else {
				if !reflect.DeepEqual(*d.Events, tt.wantEvents) {
					t.Errorf("TestDevfile200_AddEvents() wanted: %v, got: %v, difference at %v", tt.wantEvents, *d.Events, pretty.Compare(tt.wantEvents, *d.Events))
				}
			}

		})
	}
}

func TestDevfile200_UpdateEvents(t *testing.T) {

	tests := []struct {
		name          string
		currentEvents *v1.Events
		newEvents     v1.Events
	}{
		{
			name: "successfully add/update the events",
			currentEvents: &v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart: []string{"preStart1"},
				},
			},
			newEvents: v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart:  []string{"preStart2"},
					PostStart: []string{"postStart2"},
				},
			},
		},
		{
			name: "successfully update the events to empty",
			currentEvents: &v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart: []string{"preStart1"},
				},
			},
			newEvents: v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart: []string{""},
				},
			},
		},
		{
			name:          "successfully add the events to empty devfile events",
			currentEvents: nil,
			newEvents: v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart:  []string{"preStart2"},
					PostStart: []string{"postStart2"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Events: tt.currentEvents,
						},
					},
				},
			}

			d.UpdateEvents(tt.newEvents.PostStart, tt.newEvents.PostStop, tt.newEvents.PreStart, tt.newEvents.PreStop)

			events := d.GetEvents()
			if !reflect.DeepEqual(events, tt.newEvents) {
				t.Errorf("TestDevfile200_UpdateEvents() error: events did not get updated. got - %+v, wanted - %+v", events, tt.newEvents)
			}

		})
	}
}
