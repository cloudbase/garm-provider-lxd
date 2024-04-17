// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2023 Cloudbase Solutions SRL
//
// Licensed under the AGPLv3, see LICENCE file for details

package provider

import (
	"context"
	"net/http"
	"testing"

	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
	commonParams "github.com/cloudbase/garm-provider-common/params"
	"github.com/cloudbase/garm-provider-lxd/config"
	"github.com/stretchr/testify/assert"
)

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "not found error",
			err:  errInstanceIsStopped,
			want: false,
		},
		{
			name: "no error",
			err:  httpResponseErrors[http.StatusNotFound][0],
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNotFoundError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLxdInstanceToAPIInstace(t *testing.T) {
	tests := []struct {
		name           string
		instance       *api.InstanceFull
		expectedOutput commonParams.ProviderInstance
		errString      string
	}{
		{
			name: "full specs",
			instance: &api.InstanceFull{
				Instance: api.Instance{
					ExpandedConfig: map[string]string{
						"image.os":      "ubuntu",
						"image.release": "20.04",
					},
					Name:         "test-instance",
					Architecture: "x86_64",
					Status:       "Running",
					Type:         "container",
					Project:      "default",
				},
				State: &api.InstanceState{
					Status:  "Running",
					Network: nil,
				},
			},
			expectedOutput: commonParams.ProviderInstance{
				ProviderID: "test-instance",
				Name:       "test-instance",
				OSType:     commonParams.Linux,
				OSArch:     "amd64",
				OSVersion:  "20.04",
				OSName:     "ubuntu",
				Addresses:  []commonParams.Address{},
				Status:     "running",
			},
		},
		{
			name: "missing os type",
			instance: &api.InstanceFull{
				Instance: api.Instance{
					ExpandedConfig: map[string]string{
						"image.os":      "",
						"image.release": "20.04",
						"user.os-type":  "linux",
					},
					Name:         "test-instance",
					Architecture: "x86_64",
					Status:       "Running",
					Type:         "container",
					Project:      "default",
				},
				State: &api.InstanceState{
					Status:  "Running",
					Network: nil,
				},
			},
			expectedOutput: commonParams.ProviderInstance{
				ProviderID: "test-instance",
				Name:       "test-instance",
				OSType:     commonParams.Unknown,
				OSArch:     "amd64",
				OSVersion:  "20.04",
				OSName:     "",
				Addresses:  []commonParams.Address{},
				Status:     "running",
			},
		},
		{
			name: "with addresses",
			instance: &api.InstanceFull{
				Instance: api.Instance{
					ExpandedConfig: map[string]string{
						"image.os":      "ubuntu",
						"image.release": "20.04",
					},
					Name:         "test-instance",
					Architecture: "x86_64",
					Status:       "Running",
					Type:         "container",
					Project:      "default",
				},
				State: &api.InstanceState{
					Status: "Stopped",
					Network: map[string]api.InstanceStateNetwork{
						"eth0": {
							Addresses: []api.InstanceStateNetworkAddress{
								{
									Address: "10.10.10.0",
									Scope:   "global",
								},
							},
						},
						"eth1": {
							Addresses: []api.InstanceStateNetworkAddress{
								{
									Address: "10.10.0.0",
									Scope:   "link",
								},
							},
						},
					},
				},
			},
			expectedOutput: commonParams.ProviderInstance{
				ProviderID: "test-instance",
				Name:       "test-instance",
				OSType:     commonParams.Linux,
				OSArch:     "amd64",
				OSVersion:  "20.04",
				OSName:     "ubuntu",
				Addresses: []commonParams.Address{
					{
						Address: "10.10.10.0",
						Type:    commonParams.PublicAddress,
					},
				},
				Status: "stopped",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lxdInstanceToAPIInstance(tt.instance)
			assert.Equal(t, tt.expectedOutput, got)
		})
	}
}

func TestGetClientFromConfig(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name      string
		cfg       *config.LXD
		expected  lxd.InstanceServer
		errString string
	}{
		{
			name:      "nil config",
			cfg:       nil,
			expected:  nil,
			errString: "no LXD configuration found",
		},
		{
			name:      "empty config",
			cfg:       &config.LXD{},
			expected:  lxd.InstanceServer(nil),
			errString: "",
		},
		{
			name: "invalid TLSServerCert",
			cfg: &config.LXD{
				ProjectName:   "test-project",
				TLSServerCert: "bad-cert-path",
			},
			expected:  lxd.InstanceServer(nil),
			errString: "reading TLSServerCert",
		},
		{
			name: "invalid TLSCA",
			cfg: &config.LXD{
				ProjectName: "test-project",
				TLSCA:       "bad-TLSA-path",
			},
			expected:  lxd.InstanceServer(nil),
			errString: "reading TLSCA",
		},
		{
			name: "invalid ClientCertificate",
			cfg: &config.LXD{
				ProjectName:       "test-project",
				ClientCertificate: "bad-cert-path",
			},
			expected:  lxd.InstanceServer(nil),
			errString: "reading ClientCertificate",
		},
		{
			name: "invalid ClientKey",
			cfg: &config.LXD{
				ProjectName: "test-project",
				ClientKey:   "bad-key-path",
			},
			expected:  lxd.InstanceServer(nil),
			errString: "reading ClientKey",
		},
		{
			name: "UnixSocket set",
			cfg: &config.LXD{
				UnixSocket: "/var/snap/lxd/common/lxd/unix.socket",
			},
			expected:  lxd.InstanceServer(nil),
			errString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getClientFromConfig(ctx, tt.cfg)
			assert.Equal(t, got, tt.expected)
			assert.ErrorContains(t, err, tt.errString)
		})
	}
}
