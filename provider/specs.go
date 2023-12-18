// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2023 Cloudbase Solutions SRL
//
// Licensed under the AGPLv3, see LICENCE file for details

package provider

import (
	"encoding/json"

	commonParams "github.com/cloudbase/garm-provider-common/params"
	"github.com/pkg/errors"
)

type extraSpecs struct {
	DisableUpdates  bool     `json:"disable_updates"`
	ExtraPackages   []string `json:"extra_packages"`
	EnableBootDebug bool     `json:"enable_boot_debug"`
}

func parseExtraSpecsFromBootstrapParams(bootstrapParams commonParams.BootstrapInstance) (extraSpecs, error) {
	specs := extraSpecs{}
	if bootstrapParams.ExtraSpecs == nil {
		return specs, nil
	}

	if err := json.Unmarshal(bootstrapParams.ExtraSpecs, &specs); err != nil {
		return specs, errors.Wrap(err, "unmarshaling extra specs")
	}
	return specs, nil
}
