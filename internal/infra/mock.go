package infra

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockClient is a mock implementation of the OpenStack client
// It simulates OpenStack operations for development and testing
type MockClient struct {
	config      Config
	instances   map[string]*Instance
	floatingIPs map[string]*FloatingIP
	securityGroups map[string]*SecurityGroup
	dnsRecords  map[string]*DNSRecord
	containers  map[string]*Container
	volumes     map[string]*Volume
	mu          sync.RWMutex
}

// NewMockClient creates a new mock OpenStack client
func NewMockClient(cfg Config) *MockClient {
	return &MockClient{
		config:         cfg,
		instances:      make(map[string]*Instance),
		floatingIPs:    make(map[string]*FloatingIP),
		securityGroups: make(map[string]*SecurityGroup),
		dnsRecords:     make(map[string]*DNSRecord),
		containers:     make(map[string]*Container),
		volumes:        make(map[string]*Volume),
	}
}

// Instance operations

func (m *MockClient) CreateInstance(ctx context.Context, req CreateInstanceRequest) (*Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance := &Instance{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Status:    "building",
		IPAddress: generateMockIP(),
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	m.instances[instance.ID] = instance

	// Simulate async status change
	go func() {
		time.Sleep(2 * time.Second)
		m.mu.Lock()
		if inst, ok := m.instances[instance.ID]; ok {
			inst.Status = "active"
		}
		m.mu.Unlock()
	}()

	return instance, nil
}

func (m *MockClient) GetInstance(ctx context.Context, instanceID string) (*Instance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, ok := m.instances[instanceID]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}

	return instance, nil
}

func (m *MockClient) DeleteInstance(ctx context.Context, instanceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.instances[instanceID]; !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	delete(m.instances, instanceID)
	return nil
}

func (m *MockClient) WaitForInstanceStatus(ctx context.Context, instanceID string, targetStatus string) error {
	for i := 0; i < 30; i++ { // 30 attempts, 1 second each = 30 seconds max
		instance, err := m.GetInstance(ctx, instanceID)
		if err != nil {
			return err
		}

		if instance.Status == targetStatus {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for instance %s to reach status %s", instanceID, targetStatus)
}

// Network operations

func (m *MockClient) AllocateFloatingIP(ctx context.Context, req AllocateFloatingIPRequest) (*FloatingIP, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fip := &FloatingIP{
		ID:        uuid.New().String(),
		IPAddress: generateMockFloatingIP(),
		NetworkID: req.NetworkID,
		Status:    "active",
	}

	m.floatingIPs[fip.ID] = fip
	return fip, nil
}

func (m *MockClient) AttachFloatingIP(ctx context.Context, fipID string, instanceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fip, ok := m.floatingIPs[fipID]
	if !ok {
		return fmt.Errorf("floating IP not found: %s", fipID)
	}

	instance, ok := m.instances[instanceID]
	if !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	instance.FloatingIP = fip.IPAddress
	return nil
}

func (m *MockClient) CreateSecurityGroup(ctx context.Context, req CreateSecurityGroupRequest) (*SecurityGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sg := &SecurityGroup{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Rules:       req.Rules,
	}

	m.securityGroups[sg.ID] = sg
	return sg, nil
}

func (m *MockClient) CreateDNSRecord(ctx context.Context, req CreateDNSRecordRequest) (*DNSRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record := &DNSRecord{
		ID:      uuid.New().String(),
		Name:    req.Name,
		Type:    req.Type,
		Records: req.Records,
		TTL:     req.TTL,
		Status:  "active",
	}

	m.dnsRecords[record.ID] = record
	return record, nil
}

// Container operations

func (m *MockClient) CreateContainer(ctx context.Context, req CreateContainerRequest) (*Container, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	container := &Container{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Status:    "creating",
		Image:     req.Image,
		IPAddress: generateMockIP(),
		Ports:     req.Ports,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if req.FloatingIPID != "" {
		if fip, ok := m.floatingIPs[req.FloatingIPID]; ok {
			container.FloatingIP = fip.IPAddress
		}
	}

	m.containers[container.ID] = container

	// Simulate async status change
	go func() {
		time.Sleep(3 * time.Second)
		m.mu.Lock()
		if cont, ok := m.containers[container.ID]; ok {
			cont.Status = "running"
		}
		m.mu.Unlock()
	}()

	return container, nil
}

func (m *MockClient) GetContainerStatus(ctx context.Context, containerID string) (*Container, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	container, ok := m.containers[containerID]
	if !ok {
		return nil, fmt.Errorf("container not found: %s", containerID)
	}

	return container, nil
}

func (m *MockClient) StopContainer(ctx context.Context, containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	container, ok := m.containers[containerID]
	if !ok {
		return fmt.Errorf("container not found: %s", containerID)
	}

	container.Status = "stopped"
	return nil
}

func (m *MockClient) DeleteContainer(ctx context.Context, containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.containers[containerID]; !ok {
		return fmt.Errorf("container not found: %s", containerID)
	}

	delete(m.containers, containerID)
	return nil
}

func (m *MockClient) WaitForContainerStatus(ctx context.Context, containerID string, targetStatus string) error {
	for i := 0; i < 60; i++ { // 60 attempts, 1 second each = 60 seconds max
		container, err := m.GetContainerStatus(ctx, containerID)
		if err != nil {
			return err
		}

		if container.Status == targetStatus {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for container %s to reach status %s", containerID, targetStatus)
}

// Volume operations

func (m *MockClient) CreateVolume(ctx context.Context, req CreateVolumeRequest) (*Volume, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	volume := &Volume{
		ID:         uuid.New().String(),
		Name:       req.Name,
		SizeGB:     req.SizeGB,
		Status:     "available",
		VolumeType: req.VolumeType,
	}

	m.volumes[volume.ID] = volume
	return volume, nil
}

func (m *MockClient) AttachVolume(ctx context.Context, volumeID string, instanceID string, device string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	volume, ok := m.volumes[volumeID]
	if !ok {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	if _, ok := m.instances[instanceID]; !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	volume.Status = "in-use"
	volume.AttachedTo = instanceID
	return nil
}

func (m *MockClient) DetachVolume(ctx context.Context, volumeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	volume, ok := m.volumes[volumeID]
	if !ok {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	volume.Status = "available"
	volume.AttachedTo = ""
	return nil
}

func (m *MockClient) DeleteVolume(ctx context.Context, volumeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.volumes[volumeID]; !ok {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	delete(m.volumes, volumeID)
	return nil
}

// Helper functions

func generateMockIP() string {
	// Generate a mock private IP
	return fmt.Sprintf("10.0.%d.%d", 1+len(time.Now().String())%254, 1+len(time.Now().String())%254)
}

func generateMockFloatingIP() string {
	// Generate a mock public IP
	return fmt.Sprintf("41.100.%d.%d", 50+len(time.Now().String())%5, 10+len(time.Now().String())%245)
}

