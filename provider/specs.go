// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2023 Cloudbase Solutions SRL
//
// Licensed under the AGPLv3, see LICENCE file for details

package provider

import (
	"encoding/json"
	"fmt"

	commonParams "github.com/cloudbase/garm-provider-common/params"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

const jsonSchema string = `
	{
		"$schema": "http://cloudbase.it/garm-provider-lxd/schemas/extra_specs#",
		"type": "object",
		"description": "Schema defining supported extra specs for the Garm LXD Provider",
		"properties": {
			"extra_packages": {
				"type": "array",
				"description": "A list of packages that cloud-init should install on the instance.",
				"items": {
					"type": "string"
				}
			},
			"disable_updates": {
				"type": "boolean",
				"description": "Whether to disable updates when cloud-init comes online."
			},
			"enable_boot_debug": {
				"type": "boolean",
				"description": "Allows providers to set the -x flag in the runner install script."
			},
			"additionalProperties": false
		}
	}
`

type extraSpecs struct {
	DisableUpdates  bool     `json:"disable_updates"`
	ExtraPackages   []string `json:"extra_packages"`
	EnableBootDebug bool     `json:"enable_boot_debug"`
}

func jsonSchemaValidation(schema json.RawMessage) error {
	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)
	extraSpecsLoader := gojsonschema.NewBytesLoader(schema)
	result, err := gojsonschema.Validate(schemaLoader, extraSpecsLoader)
	if err != nil {
		return fmt.Errorf("failed to validate schema: %w", err)
	}
	if !result.Valid() {
		return fmt.Errorf("schema validation failed: %s", result.Errors())
	}
	return nil
}

func parseExtraSpecsFromBootstrapParams(bootstrapParams commonParams.BootstrapInstance) (extraSpecs, error) {
	specs := extraSpecs{}
	if bootstrapParams.ExtraSpecs == nil {
		return specs, nil
	}

	if err := jsonSchemaValidation(bootstrapParams.ExtraSpecs); err != nil {
		return specs, fmt.Errorf("failed to validate extra specs: %w", err)
	}

	if err := json.Unmarshal(bootstrapParams.ExtraSpecs, &specs); err != nil {
		return specs, errors.Wrap(err, "unmarshaling extra specs")
	}
	return specs, nil
}
