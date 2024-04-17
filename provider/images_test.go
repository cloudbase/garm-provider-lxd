// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2023 Cloudbase Solutions SRL
//
// Licensed under the AGPLv3, see LICENCE file for details

package provider

import (
	"fmt"
	"testing"

	"github.com/canonical/lxd/shared/api"
	"github.com/cloudbase/garm-provider-lxd/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseImageName(t *testing.T) {
	tests := []struct {
		name                   string
		image                  *image
		imageName              string
		expectedLXDImageRemote config.LXDImageRemote
		expectedImageName      string
		errString              string
	}{
		{
			name: "image with remote",
			image: &image{
				remotes: map[string]config.LXDImageRemote{
					"remote": {},
				},
			},
			imageName:              "remote:image",
			expectedImageName:      "image",
			expectedLXDImageRemote: config.LXDImageRemote{},
			errString:              "",
		},
		{
			name: "image without remote",
			image: &image{
				remotes: map[string]config.LXDImageRemote{
					"remote": {},
				},
			},
			imageName:              "image",
			expectedLXDImageRemote: config.LXDImageRemote{},
			expectedImageName:      "",
			errString:              "image does not include a remote",
		},
		{
			name: "image with invalid remote",
			image: &image{
				remotes: map[string]config.LXDImageRemote{
					"remote": {},
				},
			},
			imageName:              "invalid:image",
			expectedLXDImageRemote: config.LXDImageRemote{},
			expectedImageName:      "",
			errString:              "could not find invalid:image in",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lxdImageRemote, imageName, err := tt.image.parseImageName(tt.imageName)
			if tt.errString == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errString)
			}
			assert.Equal(t, tt.expectedLXDImageRemote, lxdImageRemote)
			assert.Equal(t, tt.expectedImageName, imageName)
		})
	}
}

func TestGetLocalImageByAlias_Success(t *testing.T) {
	cli := new(MockLXDServer)
	i := &image{
		remotes: map[string]config.LXDImageRemote{
			"remote": {},
		},
	}
	imageName := "ubuntu"
	imageType := config.LXDImageType("container")
	arch := "amd64"
	expectedImage := &api.Image{Fingerprint: "123abc"}

	aliases := map[string]*api.ImageAliasesEntry{
		"amd64": {},
	}

	cli.On("GetImageAliasArchitectures", imageType.String(), imageName).Return(aliases, nil)
	cli.On("GetImage", aliases[arch].Target).Return(expectedImage, "", nil)

	image, err := i.getLocalImageByAlias(imageName, imageType, arch, cli)

	assert.NoError(t, err)
	assert.Equal(t, expectedImage, image)
	cli.AssertExpectations(t)
}

func TestGetLocalImageByAlias_Error(t *testing.T) {
	cli := new(MockLXDServer)
	i := &image{
		remotes: map[string]config.LXDImageRemote{
			"remote": {},
		},
	}
	imageName := "ubuntu"
	imageType := config.LXDImageType("container")
	arch := "amd64"
	aliases := map[string]*api.ImageAliasesEntry{
		"amd64": {},
	}

	cli.On("GetImageAliasArchitectures", imageType.String(), imageName).Return(aliases, fmt.Errorf("error"))

	image, err := i.getLocalImageByAlias(imageName, imageType, arch, cli)

	assert.Error(t, err)
	assert.Nil(t, image)
	cli.AssertExpectations(t)
}

func TestGetInstanceSource_Success(t *testing.T) {
	cli := new(MockLXDServer)
	i := &image{
		remotes: map[string]config.LXDImageRemote{
			"remote": {},
		},
	}
	imageName := "ubuntu"
	imageType := config.LXDImageType("container")
	arch := "amd64"
	expectedImage := &api.Image{Fingerprint: "123abc"}

	aliases := map[string]*api.ImageAliasesEntry{
		"amd64": {},
	}

	cli.On("GetImageAliasArchitectures", imageType.String(), imageName).Return(aliases, nil)
	cli.On("GetImage", aliases[arch].Target).Return(expectedImage, "", nil)

	instanceSource, err := i.getInstanceSource(imageName, imageType, arch, cli)

	assert.NoError(t, err)
	assert.Equal(t, api.InstanceSource{Type: "image", Fingerprint: expectedImage.Fingerprint}, instanceSource)
	cli.AssertExpectations(t)
}

func TestGetInstanceSource_Error(t *testing.T) {
	cli := new(MockLXDServer)
	i := &image{
		remotes: map[string]config.LXDImageRemote{
			"remote": {},
		},
	}
	imageName := "ubuntu"
	imageType := config.LXDImageType("container")
	arch := "amd64"
	aliases := map[string]*api.ImageAliasesEntry{
		"amd64": {},
	}

	cli.On("GetImageAliasArchitectures", imageType.String(), imageName).Return(aliases, nil)
	cli.On("GetImage", aliases[arch].Target).Return(&api.Image{}, "", fmt.Errorf("error"))
	instanceSource, err := i.getInstanceSource(imageName, imageType, arch, cli)

	assert.Error(t, err)
	assert.Equal(t, api.InstanceSource{}, instanceSource)
	cli.AssertExpectations(t)
}
