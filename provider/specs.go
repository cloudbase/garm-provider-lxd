// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2023 Cloudbase Solutions SRL
//
// Licensed under the AGPLv3, see LICENCE file for details

package provider

import (
	"encoding/json"
	"fmt"

	cloudconfig "github.com/cloudbase/garm-provider-common/cloudconfig"
	commonParams "github.com/cloudbase/garm-provider-common/params"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type extraSpecs struct {
	ExtraPackages   []string `json:"extra_packages,omitempty" jsonschema:"title=extra packages,description=A list of packages that cloud-init should install on the instance."`
	DisableUpdates  bool     `json:"disable_updates,omitempty" jsonschema:"title=disable updates,description=Whether to disable updates when cloud-init comes online."`
	EnableBootDebug bool     `json:"enable_boot_debug,omitempty" jsonschema:"title=enable boot debug,description=Allows providers to set the -x flag in the runner install script."`
	// The Cloudconfig struct from common package
	cloudconfig.CloudConfigSpec
}

func jsonSchemaValidation(schema json.RawMessage) error {
	jsonSchema := generateJSONSchema()
	schemaLoader := gojsonschema.NewGoLoader(jsonSchema)
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
