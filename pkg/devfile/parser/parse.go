package parser

import (
	"encoding/json"

	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"k8s.io/klog"

	"reflect"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	apiOverride "github.com/devfile/api/pkg/utils/overriding"
	"github.com/pkg/errors"
)

// ParseDevfile func validates the devfile integrity.
// Creates devfile context and runtime objects
func parseDevfile(d DevfileObj, flattenedDevfile bool) (DevfileObj, error) {

	// Validate devfile
	err := d.Ctx.Validate()
	if err != nil {
		return d, err
	}

	// Create a new devfile data object
	d.Data, err = data.NewDevfileData(d.Ctx.GetApiVersion())
	if err != nil {
		return d, err
	}

	// Unmarshal devfile content into devfile struct
	err = json.Unmarshal(d.Ctx.GetDevfileContent(), &d.Data)
	if err != nil {
		return d, errors.Wrapf(err, "failed to decode devfile content")
	}

	if flattenedDevfile {
		err = parseParentAndPlugin(d)
		if err != nil {
			return DevfileObj{}, err
		}
	}

	// Successful
	return d, nil
}

// Parse func populates the flattened devfile data, parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func Parse(path string) (d DevfileObj, err error) {

	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.Populate()
	if err != nil {
		return d, err
	}
	return parseDevfile(d, true)
}

// ParseRawDevfile populates the raw devfile data witout overring and merging
func ParseRawDevfile(path string) (d DevfileObj, err error) {
	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.Populate()
	if err != nil {
		return d, err
	}
	return parseDevfile(d, false)
}

// ParseFromURL func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func ParseFromURL(url string) (d DevfileObj, err error) {
	d.Ctx = devfileCtx.NewURLDevfileCtx(url)

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.PopulateFromURL()
	if err != nil {
		return d, err
	}
	return parseDevfile(d, true)
}

// ParseFromData func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func ParseFromData(data []byte) (d DevfileObj, err error) {
	d.Ctx = devfileCtx.DevfileCtx{}
	err = d.Ctx.SetDevfileContentFromBytes(data)
	if err != nil {
		return d, errors.Wrap(err, "failed to set devfile content from bytes")
	}
	err = d.Ctx.PopulateFromRaw()
	if err != nil {
		return d, err
	}

	return parseDevfile(d, true)
}

func parseParentAndPlugin(d DevfileObj) (err error) {
	flattenedParent := &v1.DevWorkspaceTemplateSpecContent{}
	if d.Data.GetParent() != nil {
		if !reflect.DeepEqual(d.Data.GetParent(), &v1.Parent{}) && d.Data.GetParent().Uri != "" {
			parent := d.Data.GetParent()

			parentData, err := ParseFromURL(parent.Uri)
			if err != nil {
				return err
			}

			parentWorkspaceContent := parentData.Data.GetDevfileWorkspace()
			if !reflect.DeepEqual(parent.ParentOverrides, v1.ParentOverrides{}) {
				flattenedParent, err = apiOverride.OverrideDevWorkspaceTemplateSpec(parentWorkspaceContent, parent.ParentOverrides)
				if err != nil {
					return err
				}
			} else {
				flattenedParent = parentWorkspaceContent
			}

			klog.V(4).Infof("adding data of devfile with URI: %v", parent.Uri)
		}
	}
	plugins := []*v1.DevWorkspaceTemplateSpecContent{}
	components, err := d.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Plugin != nil && !reflect.DeepEqual(component.Plugin, &v1.PluginComponent{}) {
			plugin := component.Plugin
			var pluginData DevfileObj
			if plugin.Uri != "" {
				pluginData, err = ParseFromURL(plugin.Uri)
				if err != nil {
					return err
				}
			}
			pluginWorkspaceContent := pluginData.Data.GetDevfileWorkspace()
			result := pluginWorkspaceContent
			if !reflect.DeepEqual(plugin.PluginOverrides, v1.PluginOverrides{}) {
				result, err = apiOverride.OverrideDevWorkspaceTemplateSpec(pluginWorkspaceContent, plugin.PluginOverrides)
				if err != nil {
					return err
				}
			}
			plugins = append(plugins, result)
		}
	}
	mergedContent, err := apiOverride.MergeDevWorkspaceTemplateSpec(d.Data.GetDevfileWorkspace(), flattenedParent, plugins...)
	if err != nil {
		return err
	}
	d.Data.SetDevfileWorkspace(*mergedContent)
	// remove parent from flatterned devfile
	d.Data.SetParent(nil)

	return nil
}
