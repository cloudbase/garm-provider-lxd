package provider

import (
	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
	"github.com/stretchr/testify/mock"
)

type MockLXDServer struct {
	mock.Mock
}

func (m *MockLXDServer) GetImageAliasArchitectures(imageType string, imageName string) (map[string]*api.ImageAliasesEntry, error) {
	args := m.Called(imageType, imageName)
	return args.Get(0).(map[string]*api.ImageAliasesEntry), args.Error(1)
}

func (m *MockLXDServer) GetImage(fingerprint string) (*api.Image, string, error) {
	args := m.Called(fingerprint)
	return args.Get(0).(*api.Image), args.Get(1).(string), args.Error(2)
}

func (m *MockLXDServer) GetProject(projectName string) (*api.Project, string, error) {
	args := m.Called(projectName)
	return args.Get(0).(*api.Project), args.Get(1).(string), args.Error(2)
}

func (m *MockLXDServer) UseProject(projectName string) lxd.InstanceServer {
	args := m.Called(projectName)
	return args.Get(0).(lxd.InstanceServer)
}

func (m *MockLXDServer) GetProfileNames() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockLXDServer) CreateInstance(instance api.InstancesPost) (lxd.Operation, error) {
	args := m.Called(instance)
	return args.Get(0).(lxd.Operation), args.Error(1)
}

func (m *MockLXDServer) UpdateInstanceState(projectName string, state api.InstanceStatePut, instanceName string) (lxd.Operation, error) {
	args := m.Called(projectName, instanceName, state)
	return args.Get(0).(lxd.Operation), args.Error(1)
}

func (m *MockLXDServer) GetInstanceFull(projectName string) (*api.InstanceFull, string, error) {
	args := m.Called(projectName)
	return args.Get(0).(*api.InstanceFull), args.Get(1).(string), args.Error(2)
}

func (m *MockLXDServer) DeleteInstance(projectName string) (lxd.Operation, error) {
	args := m.Called(projectName)
	return args.Get(0).(lxd.Operation), args.Error(1)
}

func (m *MockLXDServer) GetInstancesFull(instanceType api.InstanceType) ([]api.InstanceFull, error) {
	args := m.Called(instanceType)
	return args.Get(0).([]api.InstanceFull), args.Error(1)
}
