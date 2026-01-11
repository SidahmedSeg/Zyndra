package k8s

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DatabaseSpec defines the specification for a managed database
type DatabaseSpec struct {
	DatabaseID   string
	DatabaseName string
	ProjectID    string
	Engine       string // postgresql, mysql, redis, mongodb
	Version      string // e.g., "16", "8.0", "7"
	SizeMB       int64  // Storage size in MB
	CPURequest   string // e.g., "100m"
	CPULimit     string // e.g., "500m"
	MemoryRequest string // e.g., "256Mi"
	MemoryLimit  string // e.g., "1Gi"
}

// DatabaseCredentials holds the auto-generated credentials
type DatabaseCredentials struct {
	Username      string
	Password      string
	Database      string
	Host          string
	Port          int32
	ConnectionURL string
}

// CreateDatabase creates a managed database using StatefulSet
func (c *Client) CreateDatabase(ctx context.Context, spec DatabaseSpec) (*DatabaseCredentials, error) {
	namespace := c.ProjectNamespace(spec.ProjectID)
	
	// Generate credentials
	password, err := generateRandomPassword(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}
	
	creds := &DatabaseCredentials{
		Username: "admin",
		Password: password,
		Database: spec.DatabaseName,
		Port:     c.getDefaultPort(spec.Engine),
	}

	// Create the secret for credentials
	if err := c.createDatabaseSecret(ctx, namespace, spec, creds); err != nil {
		return nil, err
	}

	// Create PVC for database storage
	if err := c.createDatabasePVC(ctx, namespace, spec); err != nil {
		return nil, err
	}

	// Create StatefulSet
	if err := c.createDatabaseStatefulSet(ctx, namespace, spec); err != nil {
		return nil, err
	}

	// Create Service
	if err := c.createDatabaseService(ctx, namespace, spec); err != nil {
		return nil, err
	}

	// Set host to internal service DNS
	creds.Host = fmt.Sprintf("db-%s.%s.svc.cluster.local", spec.DatabaseID[:8], namespace)
	creds.ConnectionURL = c.buildConnectionURL(spec.Engine, creds)

	return creds, nil
}

func (c *Client) createDatabaseSecret(ctx context.Context, namespace string, spec DatabaseSpec, creds *DatabaseCredentials) error {
	secretName := c.dbSecretName(spec.DatabaseID)
	
	data := map[string][]byte{
		"username": []byte(creds.Username),
		"password": []byte(creds.Password),
		"database": []byte(creds.Database),
	}

	// Add engine-specific environment variables
	switch spec.Engine {
	case "postgresql":
		data["POSTGRES_USER"] = []byte(creds.Username)
		data["POSTGRES_PASSWORD"] = []byte(creds.Password)
		data["POSTGRES_DB"] = []byte(creds.Database)
	case "mysql":
		data["MYSQL_ROOT_PASSWORD"] = []byte(creds.Password)
		data["MYSQL_DATABASE"] = []byte(creds.Database)
		data["MYSQL_USER"] = []byte(creds.Username)
		data["MYSQL_PASSWORD"] = []byte(creds.Password)
	case "redis":
		data["REDIS_PASSWORD"] = []byte(creds.Password)
	case "mongodb":
		data["MONGO_INITDB_ROOT_USERNAME"] = []byte(creds.Username)
		data["MONGO_INITDB_ROOT_PASSWORD"] = []byte(creds.Password)
		data["MONGO_INITDB_DATABASE"] = []byte(creds.Database)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "zyndra",
				"zyndra.io/database-id":         spec.DatabaseID,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}

	_, err := c.clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create database secret: %w", err)
	}

	return nil
}

func (c *Client) createDatabasePVC(ctx context.Context, namespace string, spec DatabaseSpec) error {
	pvcName := c.dbPVCName(spec.DatabaseID)
	storageClass := "longhorn"
	
	sizeStr := fmt.Sprintf("%dMi", spec.SizeMB)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "zyndra",
				"zyndra.io/database-id":         spec.DatabaseID,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: &storageClass,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(sizeStr),
				},
			},
		},
	}

	_, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create database PVC: %w", err)
	}

	return nil
}

func (c *Client) createDatabaseStatefulSet(ctx context.Context, namespace string, spec DatabaseSpec) error {
	ssName := c.dbStatefulSetName(spec.DatabaseID)
	secretName := c.dbSecretName(spec.DatabaseID)
	pvcName := c.dbPVCName(spec.DatabaseID)

	image, dataPath := c.getDatabaseImage(spec.Engine, spec.Version)

	// Build container
	container := corev1.Container{
		Name:  spec.Engine,
		Image: image,
		Ports: []corev1.ContainerPort{
			{
				Name:          spec.Engine,
				ContainerPort: c.getDefaultPort(spec.Engine),
			},
		},
		EnvFrom: []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
				},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "data",
				MountPath: dataPath,
			},
		},
		Resources: c.buildDatabaseResources(spec),
	}

	// Add liveness probe
	container.LivenessProbe = c.getDatabaseProbe(spec.Engine)
	container.ReadinessProbe = c.getDatabaseProbe(spec.Engine)

	replicas := int32(1)

	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ssName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "zyndra",
				"zyndra.io/database-id":         spec.DatabaseID,
				"zyndra.io/database-engine":     spec.Engine,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: ssName,
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"zyndra.io/database-id": spec.DatabaseID,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"zyndra.io/database-id":     spec.DatabaseID,
						"zyndra.io/database-engine": spec.Engine,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientset.AppsV1().StatefulSets(namespace).Create(ctx, ss, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create database StatefulSet: %w", err)
	}

	return nil
}

func (c *Client) createDatabaseService(ctx context.Context, namespace string, spec DatabaseSpec) error {
	svcName := c.dbServiceName(spec.DatabaseID)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "zyndra",
				"zyndra.io/database-id":         spec.DatabaseID,
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"zyndra.io/database-id": spec.DatabaseID,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       spec.Engine,
					Port:       c.getDefaultPort(spec.Engine),
					TargetPort: intstr.FromInt32(c.getDefaultPort(spec.Engine)),
				},
			},
		},
	}

	_, err := c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create database service: %w", err)
	}

	return nil
}

// DeleteDatabase deletes a managed database
func (c *Client) DeleteDatabase(ctx context.Context, projectID, databaseID string) error {
	namespace := c.ProjectNamespace(projectID)

	// Delete StatefulSet
	ssName := c.dbStatefulSetName(databaseID)
	if err := c.clientset.AppsV1().StatefulSets(namespace).Delete(ctx, ssName, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete StatefulSet: %w", err)
	}

	// Delete Service
	svcName := c.dbServiceName(databaseID)
	if err := c.clientset.CoreV1().Services(namespace).Delete(ctx, svcName, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Service: %w", err)
	}

	// Delete PVC
	pvcName := c.dbPVCName(databaseID)
	if err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete PVC: %w", err)
	}

	// Delete Secret
	secretName := c.dbSecretName(databaseID)
	if err := c.clientset.CoreV1().Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Secret: %w", err)
	}

	return nil
}

// GetDatabaseCredentials retrieves the credentials for a database
func (c *Client) GetDatabaseCredentials(ctx context.Context, projectID, databaseID, engine string) (*DatabaseCredentials, error) {
	namespace := c.ProjectNamespace(projectID)
	secretName := c.dbSecretName(databaseID)

	secret, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get database credentials: %w", err)
	}

	creds := &DatabaseCredentials{
		Username: string(secret.Data["username"]),
		Password: string(secret.Data["password"]),
		Database: string(secret.Data["database"]),
		Host:     fmt.Sprintf("db-%s.%s.svc.cluster.local", databaseID[:8], namespace),
		Port:     c.getDefaultPort(engine),
	}
	creds.ConnectionURL = c.buildConnectionURL(engine, creds)

	return creds, nil
}

// GetDatabaseStatus returns the status of a database
func (c *Client) GetDatabaseStatus(ctx context.Context, projectID, databaseID string) (*DatabaseStatus, error) {
	namespace := c.ProjectNamespace(projectID)
	ssName := c.dbStatefulSetName(databaseID)

	ss, err := c.clientset.AppsV1().StatefulSets(namespace).Get(ctx, ssName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return &DatabaseStatus{Exists: false}, nil
		}
		return nil, err
	}

	return &DatabaseStatus{
		Exists:        true,
		Ready:         ss.Status.ReadyReplicas > 0,
		Replicas:      ss.Status.Replicas,
		ReadyReplicas: ss.Status.ReadyReplicas,
	}, nil
}

// DatabaseStatus represents the status of a database
type DatabaseStatus struct {
	Exists        bool
	Ready         bool
	Replicas      int32
	ReadyReplicas int32
}

// Helper functions

func (c *Client) dbSecretName(databaseID string) string {
	return "db-creds-" + databaseID[:8]
}

func (c *Client) dbPVCName(databaseID string) string {
	return "db-data-" + databaseID[:8]
}

func (c *Client) dbStatefulSetName(databaseID string) string {
	return "db-" + databaseID[:8]
}

func (c *Client) dbServiceName(databaseID string) string {
	return "db-" + databaseID[:8]
}

func (c *Client) getDefaultPort(engine string) int32 {
	switch engine {
	case "postgresql":
		return 5432
	case "mysql":
		return 3306
	case "redis":
		return 6379
	case "mongodb":
		return 27017
	default:
		return 5432
	}
}

func (c *Client) getDatabaseImage(engine, version string) (image string, dataPath string) {
	switch engine {
	case "postgresql":
		v := version
		if v == "" {
			v = "16"
		}
		return fmt.Sprintf("postgres:%s-alpine", v), "/var/lib/postgresql/data"
	case "mysql":
		v := version
		if v == "" {
			v = "8.0"
		}
		return fmt.Sprintf("mysql:%s", v), "/var/lib/mysql"
	case "redis":
		v := version
		if v == "" {
			v = "7"
		}
		return fmt.Sprintf("redis:%s-alpine", v), "/data"
	case "mongodb":
		v := version
		if v == "" {
			v = "7"
		}
		return fmt.Sprintf("mongo:%s", v), "/data/db"
	default:
		return "postgres:16-alpine", "/var/lib/postgresql/data"
	}
}

func (c *Client) getDatabaseProbe(engine string) *corev1.Probe {
	switch engine {
	case "postgresql":
		return &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"pg_isready", "-U", "admin"},
				},
			},
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
		}
	case "mysql":
		return &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"mysqladmin", "ping", "-h", "localhost"},
				},
			},
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
		}
	case "redis":
		return &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"redis-cli", "ping"},
				},
			},
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
		}
	case "mongodb":
		return &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"mongosh", "--eval", "db.adminCommand('ping')"},
				},
			},
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
		}
	default:
		return nil
	}
}

func (c *Client) buildDatabaseResources(spec DatabaseSpec) corev1.ResourceRequirements {
	cpuReq := spec.CPURequest
	if cpuReq == "" {
		cpuReq = "100m"
	}
	cpuLim := spec.CPULimit
	if cpuLim == "" {
		cpuLim = "500m"
	}
	memReq := spec.MemoryRequest
	if memReq == "" {
		memReq = "256Mi"
	}
	memLim := spec.MemoryLimit
	if memLim == "" {
		memLim = "1Gi"
	}

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuReq),
			corev1.ResourceMemory: resource.MustParse(memReq),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuLim),
			corev1.ResourceMemory: resource.MustParse(memLim),
		},
	}
}

func (c *Client) buildConnectionURL(engine string, creds *DatabaseCredentials) string {
	switch engine {
	case "postgresql":
		return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
			creds.Username, creds.Password, creds.Host, creds.Port, creds.Database)
	case "mysql":
		return fmt.Sprintf("mysql://%s:%s@%s:%d/%s",
			creds.Username, creds.Password, creds.Host, creds.Port, creds.Database)
	case "redis":
		return fmt.Sprintf("redis://:%s@%s:%d",
			creds.Password, creds.Host, creds.Port)
	case "mongodb":
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			creds.Username, creds.Password, creds.Host, creds.Port, creds.Database)
	default:
		return ""
	}
}

func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

