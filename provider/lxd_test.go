// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2023 Cloudbase Solutions SRL
//
// Licensed under the AGPLv3, see LICENCE file for details

package provider

import (
	"context"
	"testing"

	"github.com/cloudbase/garm-provider-lxd/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/canonical/lxd/shared/api"
	commonParams "github.com/cloudbase/garm-provider-common/params"
)

func TestGetCLI(t *testing.T) {
	ctx := context.Background()

	cfg := &config.LXD{
		UnixSocket: "/var/snap/lxd/common/lxd/unix.socket",
	}
	l := &LXD{
		cfg: cfg,
		cli: &MockLXDServer{},
		imageManager: &image{
			remotes: map[string]config.LXDImageRemote{
				"remote": {
					Address: "mock-address",
				},
			},
		},
		controllerID: "controller",
	}

	_, err := l.getCLI(ctx)
	require.NoError(t, err)
}

func TestGetProfiles(t *testing.T) {
	ctx := context.Background()
	cli := new(MockLXDServer)

	cfg := &config.LXD{
		UnixSocket:            "/var/snap/lxd/common/lxd/unix.socket",
		IncludeDefaultProfile: true,
	}
	l := &LXD{
		cfg: cfg,
		cli: cli,
		imageManager: &image{
			remotes: map[string]config.LXDImageRemote{
				"remote": {
					Address: "mock-address",
				},
			},
		},
		controllerID: "controller",
	}
	expected := []string{"default", "project"}
	cli.On("GetProfileNames").Return(expected, nil)

	ret, err := l.getProfiles(ctx, "project")
	require.NoError(t, err)
	assert.Equal(t, expected, ret)
}

func TestGetCreateInstanceArgsContainer(t *testing.T) {
	ctx := context.Background()
	cli := new(MockLXDServer)

	cfg := &config.LXD{
		UnixSocket:            "/var/snap/lxd/common/lxd/unix.socket",
		InstanceType:          config.LXDImageType("container"),
		IncludeDefaultProfile: true,
	}
	l := &LXD{
		cfg: cfg,
		cli: cli,
		imageManager: &image{
			remotes: map[string]config.LXDImageRemote{
				"remote": {
					Address: "mock-address",
				},
			},
		},
		controllerID: "controller",
	}
	tools := []commonParams.RunnerApplicationDownload{
		{
			OS:           ptr("ubuntu"),
			Architecture: ptr("x86_64"),
			DownloadURL:  ptr("https://example.com"),
			Filename:     ptr("test-app"),
		},
	}
	aliases := map[string]*api.ImageAliasesEntry{
		"x86_64": {
			Name: "ubuntu",
			Type: "container",
		},
	}
	DefaultToolFetch = func(_ commonParams.OSType, _ commonParams.OSArch, tools []commonParams.RunnerApplicationDownload) (commonParams.RunnerApplicationDownload, error) {
		return tools[0], nil
	}
	DefaultGetCloudconfig = func(_ commonParams.BootstrapInstance, _ commonParams.RunnerApplicationDownload, _ string) (string, error) {
		return "#cloud-config", nil
	}
	cli.On("GetImageAliasArchitectures", config.LXDImageType("container").String(), "ubuntu").Return(aliases, nil)
	cli.On("GetImage", aliases["x86_64"].Target).Return(&api.Image{Fingerprint: "123abc"}, "", nil)
	cli.On("GetProfileNames").Return([]string{"default", "container"}, nil)
	specs := extraSpecs{}
	tests := []struct {
		name            string
		bootstrapParams commonParams.BootstrapInstance
		expected        api.InstancesPost
		errString       string
	}{
		{
			name:            "missing name",
			bootstrapParams: commonParams.BootstrapInstance{},
			expected:        api.InstancesPost{},
			errString:       "missing name",
		},
		{
			name: "looking for profile fails",
			bootstrapParams: commonParams.BootstrapInstance{
				Name:    "test-instance",
				Tools:   tools,
				Image:   "ubuntu",
				Flavor:  "bad-flavor",
				RepoURL: "mock-repo-url",
				PoolID:  "default",
				OSArch:  commonParams.Amd64,
				OSType:  commonParams.Linux,
			},
			expected:  api.InstancesPost{},
			errString: "looking for profile",
		},
		{
			name: "bad architecture fails",
			bootstrapParams: commonParams.BootstrapInstance{
				Name:    "test-instance",
				Tools:   tools,
				Image:   "ubuntu",
				Flavor:  "container",
				RepoURL: "mock-repo-url",
				PoolID:  "default",
				OSArch:  "bad-arch",
				OSType:  commonParams.Linux,
			},
			expected:  api.InstancesPost{},
			errString: "architecture bad-arch is not supported",
		},
		{
			name: "success container instance",
			bootstrapParams: commonParams.BootstrapInstance{
				Name:    "test-instance",
				Tools:   tools,
				Image:   "ubuntu",
				Flavor:  "container",
				RepoURL: "mock-repo-url",
				PoolID:  "default",
				OSArch:  commonParams.Amd64,
				OSType:  commonParams.Linux,
			},
			expected: api.InstancesPost{
				Name: "test-instance",
				InstancePut: api.InstancePut{
					Architecture: "x86_64",
					Profiles:     []string{"default", "container"},
					Description:  "Github runner provisioned by garm",
					Config: map[string]string{
						"user.user-data":    `#cloud-config`,
						osTypeKeyName:       "linux",
						osArchKeyNAme:       "amd64",
						controllerIDKeyName: "controller",
						poolIDKey:           "default",
					},
				},
				Source: api.InstanceSource{
					Type:        "image",
					Fingerprint: "123abc",
				},
				Type: "container",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, err := l.getCreateInstanceArgs(ctx, tt.bootstrapParams, specs)
			if tt.errString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, ret)
		})
	}
}

func TestGetCreateInstanceArgsVM(t *testing.T) {
	ctx := context.Background()
	cli := new(MockLXDServer)

	cfg := &config.LXD{
		UnixSocket:            "/var/snap/lxd/common/lxd/unix.socket",
		InstanceType:          config.LXDImageType("virtual-machine"),
		IncludeDefaultProfile: true,
	}
	l := &LXD{
		cfg: cfg,
		cli: cli,
		imageManager: &image{
			remotes: map[string]config.LXDImageRemote{
				"remote": {
					Address: "mock-address",
				},
			},
		},
		controllerID: "controller",
	}
	tools := []commonParams.RunnerApplicationDownload{
		{
			OS:           ptr("windows"),
			Architecture: ptr("x86_64"),
			DownloadURL:  ptr("https://example.com"),
			Filename:     ptr("test-app"),
		},
	}
	aliases := map[string]*api.ImageAliasesEntry{
		"x86_64": {
			Name: "windows",
			Type: "virtual-machine",
		},
	}
	DefaultToolFetch = func(_ commonParams.OSType, _ commonParams.OSArch, tools []commonParams.RunnerApplicationDownload) (commonParams.RunnerApplicationDownload, error) {
		return tools[0], nil
	}
	DefaultGetCloudconfig = func(_ commonParams.BootstrapInstance, _ commonParams.RunnerApplicationDownload, _ string) (string, error) {
		return "#cloud-config", nil
	}
	cli.On("GetImageAliasArchitectures", config.LXDImageType("virtual-machine").String(), "windows").Return(aliases, nil)
	cli.On("GetImage", aliases["x86_64"].Target).Return(&api.Image{Fingerprint: "123abc"}, "", nil)
	cli.On("GetProfileNames").Return([]string{"default", "virtual-machine"}, nil)
	specs := extraSpecs{}
	tests := []struct {
		name            string
		bootstrapParams commonParams.BootstrapInstance
		expected        api.InstancesPost
		errString       string
	}{
		{
			name:            "missing name",
			bootstrapParams: commonParams.BootstrapInstance{},
			expected:        api.InstancesPost{},
			errString:       "missing name",
		},
		{
			name: "success vm instance",
			bootstrapParams: commonParams.BootstrapInstance{
				Name:    "test-instance",
				Tools:   tools,
				Image:   "windows",
				Flavor:  "virtual-machine",
				RepoURL: "mock-repo-url",
				PoolID:  "default",
				OSArch:  commonParams.Amd64,
				OSType:  commonParams.Windows,
			},
			expected: api.InstancesPost{
				Name: "test-instance",
				InstancePut: api.InstancePut{
					Architecture: "x86_64",
					Profiles:     []string{"default", "virtual-machine"},
					Description:  "Github runner provisioned by garm",
					Config: map[string]string{
						"user.user-data":      "#ps1_sysnative\n" + "#cloud-config",
						osTypeKeyName:         "windows",
						osArchKeyNAme:         "amd64",
						controllerIDKeyName:   "controller",
						poolIDKey:             "default",
						"security.secureboot": "false",
					},
				},
				Source: api.InstanceSource{
					Type:        "image",
					Fingerprint: "123abc",
				},
				Type: "virtual-machine",
			},
			errString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, err := l.getCreateInstanceArgs(ctx, tt.bootstrapParams, specs)
			if tt.errString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, ret)
		})
	}

}

func TestLaunchInstance(t *testing.T) {
	ctx := context.Background()
	cli := new(MockLXDServer)

	cfg := &config.LXD{
		UnixSocket:            "/var/snap/lxd/common/lxd/unix.socket",
		InstanceType:          config.LXDImageType("container"),
		IncludeDefaultProfile: true,
	}
	l := &LXD{
		cfg: cfg,
		cli: cli,
		imageManager: &image{
			remotes: map[string]config.LXDImageRemote{
				"remote": {
					Address: "mock-address",
				},
			},
		},
		controllerID: "controller",
	}
	createArgs := api.InstancesPost{
		Name: "test-instance",
		InstancePut: api.InstancePut{
			Architecture: "x86_64",
			Profiles:     []string{"default", "container"},
			Description:  "Github runner provisioned by garm",
			Config: map[string]string{
				"user.user-data":    `#cloud-config`,
				osTypeKeyName:       "linux",
				osArchKeyNAme:       "amd64",
				controllerIDKeyName: "controller",
				poolIDKey:           "default",
			},
		},
		Source: api.InstanceSource{
			Type:        "image",
			Fingerprint: "123abc",
		},
		Type: "container",
	}
	DefaultToolFetch = func(_ commonParams.OSType, _ commonParams.OSArch, tools []commonParams.RunnerApplicationDownload) (commonParams.RunnerApplicationDownload, error) {
		return tools[0], nil
	}
	DefaultGetCloudconfig = func(_ commonParams.BootstrapInstance, _ commonParams.RunnerApplicationDownload, _ string) (string, error) {
		return "#cloud-config", nil
	}
	mockOp := new(MockOperation)
	mockOp.On("Wait").Return(nil)
	cli.On("CreateInstance", createArgs).Return(mockOp, nil)
	cli.On("UpdateInstanceState", "test-instance", "", api.InstanceStatePut{
		Action:  "start",
		Timeout: -1,
	}).Return(mockOp, nil)

	err := l.launchInstance(ctx, createArgs)
	require.NoError(t, err)
}

func TestCreateInstance(t *testing.T) {
	ctx := context.Background()
	cli := new(MockLXDServer)
	boostrapParams := commonParams.BootstrapInstance{
		Name: "test-instance",
		Tools: []commonParams.RunnerApplicationDownload{
			{
				OS:           ptr("windows"),
				Architecture: ptr("x86_64"),
				DownloadURL:  ptr("https://example.com"),
				Filename:     ptr("test-app"),
			},
		},
		Image:   "windows",
		Flavor:  "virtual-machine",
		RepoURL: "mock-repo-url",
		PoolID:  "default",
		OSArch:  commonParams.Amd64,
		OSType:  commonParams.Windows,
	}
	cfg := &config.LXD{
		UnixSocket:            "/var/snap/lxd/common/lxd/unix.socket",
		InstanceType:          config.LXDImageType("virtul-machine"),
		IncludeDefaultProfile: true,
	}
	l := &LXD{
		cfg: cfg,
		cli: cli,
		imageManager: &image{
			remotes: map[string]config.LXDImageRemote{
				"remote": {
					Address: "mock-address",
				},
			},
		},
		controllerID: "controller",
	}
	aliases := map[string]*api.ImageAliasesEntry{
		"x86_64": {
			Name: "windows",
			Type: "virtual-machine",
		},
	}
	expectedOutput := commonParams.ProviderInstance{
		OSArch:     commonParams.Amd64,
		ProviderID: "test-instance",
		Name:       "test-instance",
		OSType:     commonParams.Windows,
		OSName:     "windows",
		OSVersion:  "",
		Addresses: []commonParams.Address{
			{
				Address: "10.10.0.0",
				Type:    commonParams.PublicAddress,
			},
		},
		Status: commonParams.InstanceRunning,
	}
	DefaultToolFetch = func(_ commonParams.OSType, _ commonParams.OSArch, tools []commonParams.RunnerApplicationDownload) (commonParams.RunnerApplicationDownload, error) {
		return tools[0], nil
	}
	DefaultGetCloudconfig = func(_ commonParams.BootstrapInstance, _ commonParams.RunnerApplicationDownload, _ string) (string, error) {
		return "#cloud-config", nil
	}
	cli.On("GetImageAliasArchitectures", config.LXDImageType("virtual-machine").String(), "windows").Return(aliases, nil)
	cli.On("GetImage", aliases["x86_64"].Target).Return(&api.Image{Fingerprint: "123abc"}, "", nil)
	cli.On("GetProfileNames").Return([]string{"default", "virtual-machine"}, nil)
	mockOp := new(MockOperation)
	mockOp.On("Wait").Return(nil)
	cli.On("CreateInstance", mock.Anything).Return(mockOp, nil)
	cli.On("UpdateInstanceState", "test-instance", "", api.InstanceStatePut{
		Action:  "start",
		Timeout: -1,
	}).Return(mockOp, nil)
	cli.On("GetInstanceFull", "test-instance").Return(&api.InstanceFull{
		Instance: api.Instance{
			Name:         "test-instance",
			Architecture: "x86_64",
			ExpandedConfig: map[string]string{
				"image.os":      "windows",
				"image.release": "",
			},
			Type: "container",
		},
		State: &api.InstanceState{
			Status: "Running",
			Network: map[string]api.InstanceStateNetwork{
				"eth0": {
					Addresses: []api.InstanceStateNetworkAddress{
						{
							Address: "10.10.0.0",
							Scope:   "global",
						},
					},
				},
			},
		},
	}, "", nil)

	provider, err := l.CreateInstance(ctx, boostrapParams)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, provider)
}

func TestGetInstance(t *testing.T) {
	ctx := context.Background()
	cli := new(MockLXDServer)
	instanceName := "test-instance"

	cfg := &config.LXD{
		UnixSocket:            "/var/snap/lxd/common/lxd/unix.socket",
		InstanceType:          config.LXDImageType("container"),
		IncludeDefaultProfile: true,
	}
	l := &LXD{
		cfg: cfg,
		cli: cli,
		imageManager: &image{
			remotes: map[string]config.LXDImageRemote{
				"remote": {
					Address: "mock-address",
				},
			},
		},
		controllerID: "controller",
	}
	cli.On("GetInstanceFull", "test-instance").Return(&api.InstanceFull{
		Instance: api.Instance{
			Name:         "test-instance",
			Architecture: "x86_64",
			ExpandedConfig: map[string]string{
				"image.os":      "windows",
				"image.release": "",
			},
			Type: "container",
		},
		State: &api.InstanceState{
			Status: "Running",
			Network: map[string]api.InstanceStateNetwork{
				"eth0": {
					Addresses: []api.InstanceStateNetworkAddress{
						{
							Address: "10.10.0.0",
							Scope:   "global",
						},
					},
				},
			},
		},
	}, "", nil)
	expectedOutput := commonParams.ProviderInstance{
		OSArch:     commonParams.Amd64,
		ProviderID: "test-instance",
		Name:       "test-instance",
		OSType:     commonParams.Windows,
		OSName:     "windows",
		OSVersion:  "",
		Addresses: []commonParams.Address{
			{
				Address: "10.10.0.0",
				Type:    commonParams.PublicAddress,
			},
		},
		Status: commonParams.InstanceRunning,
	}

	provider, err := l.GetInstance(ctx, instanceName)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, provider)
}

func TestDeleteInstance(t *testing.T) {
	ctx := context.Background()
	cli := new(MockLXDServer)
	instanceName := "test-instance"
	cfg := &config.LXD{
		UnixSocket:            "/var/snap/lxd/common/lxd/unix.socket",
		InstanceType:          config.LXDImageType("container"),
		IncludeDefaultProfile: true,
	}
	l := &LXD{
		cfg: cfg,
		cli: cli,
		imageManager: &image{
			remotes: map[string]config.LXDImageRemote{
				"remote": {
					Address: "mock-address",
				},
			},
		},
		controllerID: "controller",
	}
	mockOp := new(MockOperation)
	mockOp.On("WaitContext", mock.Anything).Return(nil)
	cli.On("DeleteInstance", instanceName).Return(mockOp, nil)
	cli.On("UpdateInstanceState", "test-instance", "", api.InstanceStatePut{
		Action:  "stop",
		Timeout: -1,
		Force:   true,
	}).Return(mockOp, nil)
	err := l.DeleteInstance(ctx, instanceName)
	require.NoError(t, err)
}

func TestListInstances(t *testing.T) {
	ctx := context.Background()
	poolID := "test-pool-id"
	cli := new(MockLXDServer)
	cfg := &config.LXD{
		UnixSocket:            "/var/snap/lxd/common/lxd/unix.socket",
		InstanceType:          config.LXDImageType("container"),
		IncludeDefaultProfile: true,
	}
	l := &LXD{
		cfg: cfg,
		cli: cli,
		imageManager: &image{
			remotes: map[string]config.LXDImageRemote{
				"remote": {
					Address: "mock-address",
				},
			},
		},
		controllerID: "controller",
	}
	DefaultToolFetch = func(_ commonParams.OSType, _ commonParams.OSArch, tools []commonParams.RunnerApplicationDownload) (commonParams.RunnerApplicationDownload, error) {
		return tools[0], nil
	}
	DefaultGetCloudconfig = func(_ commonParams.BootstrapInstance, _ commonParams.RunnerApplicationDownload, _ string) (string, error) {
		return "#cloud-config", nil
	}
	cli.On("GetInstancesFull", api.InstanceTypeAny).Return([]api.InstanceFull{
		{
			Instance: api.Instance{
				Name:         "test-instance",
				Architecture: "x86_64",
				ExpandedConfig: map[string]string{
					"image.os":          "windows",
					"image.release":     "",
					poolIDKey:           poolID,
					controllerIDKeyName: "controller",
				},
				Type: "container",
			},
			State: &api.InstanceState{
				Status: "Running",
				Network: map[string]api.InstanceStateNetwork{
					"eth0": {
						Addresses: []api.InstanceStateNetworkAddress{
							{
								Address: "10.10.0.0",
								Scope:   "global",
							},
						},
					},
				},
			},
		},
	}, nil)
	expectedOutput := []commonParams.ProviderInstance{
		{
			OSArch:     commonParams.Amd64,
			ProviderID: "test-instance",
			Name:       "test-instance",
			OSType:     commonParams.Windows,
			OSName:     "windows",
			OSVersion:  "",
			Addresses: []commonParams.Address{
				{
					Address: "10.10.0.0",
					Type:    commonParams.PublicAddress,
				},
			},
			Status: commonParams.InstanceRunning,
		},
	}

	providers, err := l.ListInstances(ctx, poolID)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, providers)
}

func TestRemoveAllInstances(t *testing.T) {
	ctx := context.Background()
	poolID := "test-pool-id"
	instanceName := "test-instance"
	cli := new(MockLXDServer)
	l := &LXD{
		cfg:          &config.LXD{},
		cli:          cli,
		imageManager: &image{},
		controllerID: "controller",
	}
	cli.On("GetInstancesFull", api.InstanceTypeAny).Return([]api.InstanceFull{
		{
			Instance: api.Instance{
				Name:         instanceName,
				Architecture: "x86_64",
				ExpandedConfig: map[string]string{
					"image.os":          "windows",
					"image.release":     "",
					poolIDKey:           poolID,
					controllerIDKeyName: "controller",
				},
				Type: "container",
			},
			State: &api.InstanceState{
				Status: "Running",
			},
		},
	}, nil)
	mockOp := new(MockOperation)
	mockOp.On("WaitContext", mock.Anything).Return(nil)
	cli.On("DeleteInstance", instanceName).Return(mockOp, nil)
	cli.On("UpdateInstanceState", "test-instance", "", api.InstanceStatePut{
		Action:  "stop",
		Timeout: -1,
		Force:   true,
	}).Return(mockOp, nil)

	err := l.RemoveAllInstances(ctx)
	require.NoError(t, err)
}

func TestStop(t *testing.T) {
	ctx := context.Background()
	cli := new(MockLXDServer)
	force := true
	instanceName := "test-instance"
	l := &LXD{
		cfg:          &config.LXD{},
		cli:          cli,
		imageManager: &image{},
		controllerID: "controller",
	}
	mockOp := new(MockOperation)
	mockOp.On("WaitContext", mock.Anything).Return(nil)
	cli.On("UpdateInstanceState", instanceName, "", api.InstanceStatePut{
		Action:  "stop",
		Timeout: -1,
		Force:   force,
	}).Return(mockOp, nil)

	err := l.Stop(ctx, instanceName, force)
	require.NoError(t, err)
}

func TestStart(t *testing.T) {
	ctx := context.Background()
	cli := new(MockLXDServer)
	instanceName := "test-instance"
	l := &LXD{
		cfg:          &config.LXD{},
		cli:          cli,
		imageManager: &image{},
		controllerID: "controller",
	}
	mockOp := new(MockOperation)
	mockOp.On("WaitContext", mock.Anything).Return(nil)
	cli.On("UpdateInstanceState", instanceName, "", api.InstanceStatePut{
		Action:  "start",
		Timeout: -1,
		Force:   false,
	}).Return(mockOp, nil)

	err := l.Start(ctx, instanceName)
	require.NoError(t, err)
}
