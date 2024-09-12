// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2023 Cloudbase Solutions SRL
//
// Licensed under the AGPLv3, see LICENCE file for details

package provider

import (
	"encoding/json"
	"testing"

	"github.com/cloudbase/garm-provider-common/cloudconfig"
	commonParams "github.com/cloudbase/garm-provider-common/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonSchemaValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     json.RawMessage
		errString string
	}{
		{
			name: "Valid input",
			input: json.RawMessage(`{
				"disable_updates": true,
				"extra_packages": ["openssh-server", "jq"],
				"enable_boot_debug": false
			}`),
			errString: "",
		},
		{
			name: "Invalid input - wrong data type",
			input: json.RawMessage(`{
				"disable_updates": "true"
			}`),
			errString: "schema validation failed: [disable_updates: Invalid type. Expected: boolean, given: string]",
		},
		{
			name: "Invalid input - additional property",
			input: json.RawMessage(`{
				"additional_property": true
			}`),
			errString: "Additional property additional_property is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jsonSchemaValidation(tt.input)
			if tt.errString == "" {
				assert.NoError(t, err, "Expected no error, got %v", err)
			} else {
				assert.Error(t, err, "Expected an error")
				if err != nil {
					assert.Contains(t, err.Error(), tt.errString, "Error message does not match")
				}
			}
		})
	}
}

func TestParseExtraSpecsFromBootstrapParams(t *testing.T) {
	tests := []struct {
		name            string
		bootstrapParams commonParams.BootstrapInstance
		expectedOutput  extraSpecs
		errString       string
	}{
		{
			name: "full specs",
			bootstrapParams: commonParams.BootstrapInstance{
				ExtraSpecs: []byte(`{"disable_updates": true, "extra_packages": ["package1", "package2"], "enable_boot_debug": true, "runner_install_template": "IyEvYmluL2Jhc2gKZWNobyBJbnN0YWxsaW5nIHJ1bm5lci4uLg==", "pre_install_scripts": {"setup.sh": "IyEvYmluL2Jhc2gKZWNobyBTZXR1cCBzY3JpcHQuLi4="}, "extra_context": {"key": "value"}}`),
			},
			expectedOutput: extraSpecs{
				DisableUpdates:  true,
				ExtraPackages:   []string{"package1", "package2"},
				EnableBootDebug: true,
				CloudConfigSpec: cloudconfig.CloudConfigSpec{
					RunnerInstallTemplate: []byte("#!/bin/bash\necho Installing runner..."),
					PreInstallScripts: map[string][]byte{
						"setup.sh": []byte("#!/bin/bash\necho Setup script..."),
					},
					ExtraContext: map[string]string{"key": "value"},
				},
			},
			errString: "",
		},
		{
			name: "empty specs",
			bootstrapParams: commonParams.BootstrapInstance{
				ExtraSpecs: []byte(`{}`),
			},
			expectedOutput: extraSpecs{},
			errString:      "",
		},
		{
			name: "invalid json",
			bootstrapParams: commonParams.BootstrapInstance{
				ExtraSpecs: []byte(`{"disable_updates": true, "extra_packages": ["package1", "package2", "enable_boot_debug": true}`),
			},
			expectedOutput: extraSpecs{},
			errString:      "failed to validate extra specs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExtraSpecsFromBootstrapParams(tt.bootstrapParams)
			assert.Equal(t, tt.expectedOutput, got)
			if tt.errString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
