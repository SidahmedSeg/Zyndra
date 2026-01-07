package worker

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"

	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/infra"
	"github.com/intelifox/click-deploy/internal/metrics"
	"github.com/intelifox/click-deploy/internal/store"
)

// DatabaseWorker processes database provisioning jobs
type DatabaseWorker struct {
	store  *store.DB
	config *config.Config
	client infra.Client
}

// NewDatabaseWorker creates a new database worker
func NewDatabaseWorker(store *store.DB, cfg *config.Config, client infra.Client) *DatabaseWorker {
	return &DatabaseWorker{
		store:  store,
		config: cfg,
		client: client,
	}
}

// ProcessProvisionDatabaseJob processes a database provisioning job
func (w *DatabaseWorker) ProcessProvisionDatabaseJob(ctx context.Context, databaseID uuid.UUID) error {
	// Get database
	database, err := w.store.GetDatabase(ctx, databaseID)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}
	if database == nil {
		return fmt.Errorf("database not found: %s", databaseID)
	}

	// Update status
	w.store.UpdateDatabase(ctx, databaseID, &store.Database{
		Status: "provisioning",
	})

	// Get project for OpenStack tenant ID
	var projectID uuid.UUID
	if database.ServiceID.Valid {
		serviceID, _ := uuid.Parse(database.ServiceID.String)
		service, err := w.store.GetService(ctx, serviceID)
		if err != nil {
			return fmt.Errorf("failed to get service: %w", err)
		}
		projectID = service.ProjectID
	} else {
		// Database not linked to service - need project ID
		// For now, we'll need to get it from the request context or add project_id to database
		return fmt.Errorf("database must be linked to a service")
	}

	project, err := w.store.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Create infra client config
	infraConfig := infra.Config{
		BaseURL:  w.config.InfraServiceURL,
		APIKey:   w.config.InfraServiceAPIKey,
		TenantID: project.OpenStackTenantID,
		UseMock:  w.config.UseMockInfra,
	}

	baseClient := infra.NewClient(infraConfig)
	client := infra.NewRetryClient(baseClient)

	// Step 1: Create volume
	volumeSizeGB := (database.VolumeSizeMB + 1023) / 1024 // Round up to GB
	volumeReq := infra.CreateVolumeRequest{
		Name:       fmt.Sprintf("db-%s", databaseID.String()[:8]),
		SizeGB:     volumeSizeGB,
		VolumeType: "ssd",
	}

	volume, err := client.CreateVolume(ctx, volumeReq)
	if err != nil {
		w.store.UpdateDatabase(ctx, databaseID, &store.Database{
			Status: "error",
		})
		return fmt.Errorf("failed to create volume: %w", err)
	}

	// Update database with volume ID
	w.store.UpdateDatabase(ctx, databaseID, &store.Database{
		VolumeID: sql.NullString{String: volume.ID, Valid: true},
	})

	// Step 2: Create security group
	sgReq := infra.CreateSecurityGroupRequest{
		Name:        fmt.Sprintf("db-%s-sg", databaseID.String()[:8]),
		Description: fmt.Sprintf("Security group for database %s", databaseID.String()[:8]),
		Rules: []infra.SecurityGroupRule{
			{
				Direction: "ingress",
				Protocol:  "tcp",
				PortMin:   getDefaultPort(database.Engine),
				PortMax:   getDefaultPort(database.Engine),
				RemoteIP:  "10.0.0.0/8", // Internal network only
			},
		},
	}

	sg, err := client.CreateSecurityGroup(ctx, sgReq)
	if err != nil {
		w.store.UpdateDatabase(ctx, databaseID, &store.Database{
			Status: "error",
		})
		return fmt.Errorf("failed to create security group: %w", err)
	}

	// Step 3: Create database instance (using Nova for now, or Trove if available)
	// For mock, we'll create a container-like instance
	version := ""
	if database.Version.Valid {
		version = database.Version.String
	}
	
	networkID := ""
	if project.OpenStackNetworkID.Valid {
		networkID = project.OpenStackNetworkID.String
	}

	// Generate cloud-init script with Node Exporter
	userData := metrics.GenerateCloudInitScript()

	instanceReq := infra.CreateInstanceRequest{
		Name:          fmt.Sprintf("db-%s", databaseID.String()[:8]),
		FlavorID:      getFlavorID(database.Size),
		ImageID:       getDatabaseImage(database.Engine, version),
		NetworkID:     networkID,
		SecurityGroups: []string{sg.ID},
		UserData:      userData,
		Metadata: map[string]string{
			"database_id": databaseID.String(),
			"engine":      database.Engine,
		},
	}

	instance, err := client.CreateInstance(ctx, instanceReq)
	if err != nil {
		w.store.UpdateDatabase(ctx, databaseID, &store.Database{
			Status: "error",
		})
		return fmt.Errorf("failed to create instance: %w", err)
	}

	// Wait for instance to be active
	if err := client.WaitForInstanceStatus(ctx, instance.ID, "active"); err != nil {
		w.store.UpdateDatabase(ctx, databaseID, &store.Database{
			Status: "error",
		})
		return fmt.Errorf("instance failed to become active: %w", err)
	}

	// Step 4: Attach volume
	device := "/dev/vdb" // Standard device path
	if err := client.AttachVolume(ctx, volume.ID, instance.ID, device); err != nil {
		w.store.UpdateDatabase(ctx, databaseID, &store.Database{
			Status: "error",
		})
		return fmt.Errorf("failed to attach volume: %w", err)
	}

	// Step 5: Generate credentials
	username := generateUsername(database.Engine)
	password := generatePassword(16)
	databaseName := generateDatabaseName(databaseID)

	// Step 6: Generate internal hostname
	hostname := fmt.Sprintf("db%s.internal.armonika.cloud", databaseID.String()[:8])

	// Step 7: Create internal DNS record
	dnsReq := infra.CreateDNSRecordRequest{
		ZoneID:  w.config.DNSZoneID, // TODO: Add to config
		Name:    hostname,
		Type:    "A",
		Records: []string{instance.IPAddress},
		TTL:     300,
	}

	_, err = client.CreateDNSRecord(ctx, dnsReq)
	if err != nil {
		// DNS is optional, log but don't fail
		fmt.Printf("Warning: Failed to create DNS record: %v\n", err)
	}

	// Step 8: Generate connection URL
	port := getDefaultPort(database.Engine)
	connectionURL := generateConnectionURL(database.Engine, username, password, hostname, port, databaseName)

	// Step 9: Update database with all information
	w.store.UpdateDatabase(ctx, databaseID, &store.Database{
		InternalHostname:    sql.NullString{String: hostname, Valid: true},
		InternalIP:          sql.NullString{String: instance.IPAddress, Valid: true},
		Port:                sql.NullInt64{Int64: int64(port), Valid: true},
		Username:            sql.NullString{String: username, Valid: true},
		Password:            sql.NullString{String: password, Valid: true}, // TODO: Encrypt
		DatabaseName:       sql.NullString{String: databaseName, Valid: true},
		ConnectionURL:      sql.NullString{String: connectionURL, Valid: true},
		OpenStackInstanceID: sql.NullString{String: instance.ID, Valid: true},
		SecurityGroupID:     sql.NullString{String: sg.ID, Valid: true},
		Status:              "active",
	})

	return nil
}

// Helper functions

func getDefaultPort(engine string) int {
	switch engine {
	case "postgresql":
		return 5432
	case "mysql":
		return 3306
	case "redis":
		return 6379
	default:
		return 5432
	}
}

func getFlavorID(size string) string {
	switch size {
	case "small":
		return "small"
	case "medium":
		return "medium"
	case "large":
		return "large"
	default:
		return "small"
	}
}

func getDatabaseImage(engine, version string) string {
	// Return appropriate image ID based on engine and version
	// This would be configured per OpenStack environment
	switch engine {
	case "postgresql":
		if version == "" {
			return "postgresql-14"
		}
		return fmt.Sprintf("postgresql-%s", version)
	case "mysql":
		if version == "" {
			return "mysql-8.0"
		}
		return fmt.Sprintf("mysql-%s", version)
	case "redis":
		if version == "" {
			return "redis-7"
		}
		return fmt.Sprintf("redis-%s", version)
	default:
		return "postgresql-14"
	}
}

func generateUsername(engine string) string {
	prefix := "db"
	switch engine {
	case "postgresql":
		prefix = "pg"
	case "mysql":
		prefix = "mysql"
	case "redis":
		prefix = "redis"
	}

	random, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("%s%d", prefix, random.Int64())
}

func generatePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[random.Int64()]
	}
	return string(password)
}

func generateDatabaseName(databaseID uuid.UUID) string {
	return fmt.Sprintf("db_%s", databaseID.String()[:8])
}

func generateConnectionURL(engine, username, password, hostname string, port int, databaseName string) string {
	switch engine {
	case "postgresql":
		return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", username, password, hostname, port, databaseName)
	case "mysql":
		return fmt.Sprintf("mysql://%s:%s@%s:%d/%s", username, password, hostname, port, databaseName)
	case "redis":
		return fmt.Sprintf("redis://%s:%s@%s:%d", username, password, hostname, port)
	default:
		return fmt.Sprintf("%s://%s:%s@%s:%d/%s", engine, username, password, hostname, port, databaseName)
	}
}

