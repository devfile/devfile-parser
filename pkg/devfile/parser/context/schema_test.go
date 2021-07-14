package parser

import (
	"github.com/stretchr/testify/assert"
	"testing"

	v200 "github.com/devfile/library/pkg/devfile/parser/data/v2/2.0.0"
)

const (
	validJson200 = `{
		"schemaVersion": "2.0.0",
		"metadata": {
		   "name": "nodejs-stack"
		},
		"projects": [
		   {
			  "name": "project",
			  "git": {
				 "remotes": {
					 "origin": "https://github.com/che-samples/web-nodejs-sample.git"
				 }
			  }
		   }
		],
		"components": [
		   {
			  "name": "che-theia-plugin",
			  "plugin": {
				 "id": "eclipse/che-theia/7.1.0"
			  }
		   },
		   {
			  "name": "che-exec-plugin",
			  "plugin": {
				 "id": "eclipse/che-machine-exec-plugin/7.1.0"
			  }
		   },
		   {
			  "name": "typescript-plugin",
			  "plugin": {
				 "id": "che-incubator/typescript/1.30.2",
				 "components": [
					{
					   "name": "somecontainer",
					   "container": {
						  "memoryLimit": "512Mi"
					   }
					}
				 ]
			  }
		   },
		   {
			  "name": "nodejs",
			  "container": {
				 "image": "quay.io/eclipse/che-nodejs10-ubi:nightly",
				 "memoryLimit": "512Mi",
				 "endpoints": [
					{
					   "name": "nodejs",
					   "protocol": "http",
					   "targetPort": 3000
					}
				 ],
				 "mountSources": true
			  }
		   }
		],
		"commands": [
		   {
			  "id": "download-dependencies",
			  "exec": {
				 "component": "nodejs",
				 "commandLine": "npm install",
				 "workingDir": "${PROJECTS_ROOT}/project/app",
				 "group": {
					"kind": "build"
				 }
			  }
		   },
		   {
			  "id": "run-the-app",
			  "exec": {
				 "component": "nodejs",
				 "commandLine": "nodemon app.js",
				 "workingDir": "${PROJECTS_ROOT}/project/app",
				 "group": {
					"kind": "run",
					"isDefault": true
				 }
			  }
		   },
		   {
			  "id": "run-the-app-debugging-enabled", 
			  "exec": {
				 "component": "nodejs",
				 "commandLine": "nodemon --inspect app.js",
				 "workingDir": "${PROJECTS_ROOT}/project/app",
				 "group": {
					"kind": "run"
				 }
			  }
		   },
		   {
			  "id": "stop-the-app",
			  "exec": {
				 "component": "nodejs",
				 "commandLine": "node_server_pids=$(pgrep -fx '.*nodemon (--inspect )?app.js' | tr \"\\\\n\" \" \") && echo \"Stopping node server with PIDs: ${node_server_pids}\" &&  kill -15 ${node_server_pids} &>/dev/null && echo 'Done.'"
			  }
		   },
		   {
			  "id": "attach-remote-debugger", 
			  "vscodeLaunch": {
				 "inlined": "{\n  \"version\": \"0.2.0\",\n  \"configurations\": [\n    {\n      \"type\": \"node\",\n      \"request\": \"attach\",\n      \"name\": \"Attach to Remote\",\n      \"address\": \"localhost\",\n      \"port\": 9229,\n      \"localRoot\": \"${workspaceFolder}\",\n      \"remoteRoot\": \"${workspaceFolder}\"\n    }\n  ]\n}\n"
			  }
		   }
		]
	 }`
)

func TestValidateDevfileSchema(t *testing.T) {

	t.Run("valid 2.0.0 json schema", func(t *testing.T) {

		var (
			d = DevfileCtx{
				jsonSchema: v200.JsonSchema200,
				rawContent: validJsonRawContent200(),
			}
		)

		err := d.ValidateDevfileSchema()
		if err != nil {
			t.Errorf("TestValidateDevfileSchema() unexpected error: '%v'", err)
		}
	})

	expectedErr := "invalid devfile schema. errors :\n*.*schemaVersion is required"
	t.Run("invalid 2.0.0 json schema", func(t *testing.T) {

		var (
			d = DevfileCtx{
				jsonSchema: v200.JsonSchema200,
				rawContent: []byte("{}"),
			}
		)

		err := d.ValidateDevfileSchema()
		if err == nil {
			t.Errorf("TestValidateDevfileSchema() expected error, didn't get one")
		} else {
			assert.Regexp(t, expectedErr, err.Error(), "TestValidateDevfileSchema(): Error message should match")
		}
	})
}

func validJsonRawContent200() []byte {
	return []byte(validJson200)
}
