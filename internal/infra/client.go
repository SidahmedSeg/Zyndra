package infra

import (
	"context"
)

// Client is the interface for OpenStack operations
// This allows us to swap between mock and real implementations
type Client interface {
	// Instance operations
	CreateInstance(ctx context.Context, req CreateInstanceRequest) (*Instance, error)
	GetInstance(ctx context.Context, instanceID string) (*Instance, error)
	DeleteInstance(ctx context.Context, instanceID string) error
	WaitForInstanceStatus(ctx context.Context, instanceID string, status string) error

	// Network operations
	AllocateFloatingIP(ctx context.Context, req AllocateFloatingIPRequest) (*FloatingIP, error)
	AttachFloatingIP(ctx context.Context, fipID string, instanceID string) error
	CreateSecurityGroup(ctx context.Context, req CreateSecurityGroupRequest) (*SecurityGroup, error)
	CreateDNSRecord(ctx context.Context, req CreateDNSRecordRequest) (*DNSRecord, error)

	// Container operations
	CreateContainer(ctx context.Context, req CreateContainerRequest) (*Container, error)
	GetContainerStatus(ctx context.Context, containerID string) (*Container, error)
	StopContainer(ctx context.Context, containerID string) error
	DeleteContainer(ctx context.Context, containerID string) error
	WaitForContainerStatus(ctx context.Context, containerID string, status string) error

	// Volume operations
	CreateVolume(ctx context.Context, req CreateVolumeRequest) (*Volume, error)
	AttachVolume(ctx context.Context, volumeID string, instanceID string, device string) error
	DetachVolume(ctx context.Context, volumeID string) error
	DeleteVolume(ctx context.Context, volumeID string) error
}

// Config holds configuration for the OpenStack client
type Config struct {
	BaseURL    string
	APIKey     string
	TenantID   string
	UseMock    bool // If true, use mock client instead of real HTTP client
}

// NewClient creates a new OpenStack client (mock or real based on config)
func NewClient(cfg Config) Client {
	if cfg.UseMock {
		return NewMockClient(cfg)
	}
	return NewHTTPClient(cfg)
}

// Request/Response types

type CreateInstanceRequest struct {
	Name        string
	FlavorID    string
	ImageID     string
	NetworkID   string
	SecurityGroups []string
	UserData    string
	Metadata    map[string]string
}

type Instance struct {
	ID          string
	Name        string
	Status      string // active, building, error, etc.
	IPAddress   string
	FloatingIP  string
	CreatedAt   string
}

type AllocateFloatingIPRequest struct {
	NetworkID string
}

type FloatingIP struct {
	ID        string
	IPAddress string
	NetworkID string
	Status    string
}

type CreateSecurityGroupRequest struct {
	Name        string
	Description string
	Rules       []SecurityGroupRule
}

type SecurityGroupRule struct {
	Direction string // ingress, egress
	Protocol  string // tcp, udp, icmp
	PortMin   int
	PortMax   int
	RemoteIP  string // CIDR notation
}

type SecurityGroup struct {
	ID          string
	Name        string
	Description string
	Rules       []SecurityGroupRule
}

type CreateDNSRecordRequest struct {
	ZoneID  string
	Name    string
	Type    string // A, AAAA, CNAME
	Records []string
	TTL     int
}

type DNSRecord struct {
	ID      string
	Name    string
	Type    string
	Records []string
	TTL     int
	Status  string
}

type CreateContainerRequest struct {
	Name            string
	Image           string
	Command         []string
	Environment     map[string]string
	Ports           []ContainerPort
	SecurityGroupID string
	NetworkID       string
	FloatingIPID    string
}

type ContainerPort struct {
	ContainerPort int
	HostPort      int
	Protocol      string
}

type Container struct {
	ID          string
	Name        string
	Status      string // running, stopped, error
	Image       string
	IPAddress   string
	FloatingIP  string
	Ports       []ContainerPort
	CreatedAt   string
}

type CreateVolumeRequest struct {
	Name     string
	SizeGB   int
	VolumeType string
}

type Volume struct {
	ID         string
	Name       string
	SizeGB     int
	Status     string // available, in-use, error
	AttachedTo string // instance ID if attached
	VolumeType string
}

