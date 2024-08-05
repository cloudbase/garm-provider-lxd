// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2023 Cloudbase Solutions SRL
//
// Licensed under the AGPLv3, see LICENCE file for details

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudbase/garm-provider-common/execution"

	"github.com/cloudbase/garm-provider-lxd/provider"
)

var signals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
}

var (
	// Version is the version of the application
	Version = "v0.0.0-unknown"
)

func main() {
	// This is an unofficial command. It will be added into future versions of the
	// external provider interface. For now we manually hardcode it here. This is not
	// used by GARM itself. It is informative for the user to be able to check the version
	// of the provider.
	garmCommand := os.Getenv("GARM_COMMAND")
	if garmCommand == "GetVersion" {
		fmt.Println(Version)
		os.Exit(0)
	}

	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	executionEnv, err := execution.GetEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	prov, err := provider.NewLXDProvider(executionEnv.ProviderConfigFile, executionEnv.ControllerID)
	if err != nil {
		log.Fatal(err)
	}

	result, err := execution.Run(ctx, prov, executionEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run command: %s", err)
		os.Exit(execution.ResolveErrorToExitCode(err))
	}
	if len(result) > 0 {
		fmt.Fprint(os.Stdout, result)
	}
}
