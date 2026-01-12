package api

// GitSourceInfo represents git source information for service creation
type GitSourceInfo struct {
	Provider  string  `json:"provider" validate:"required,oneof=github gitlab"`
	RepoOwner string  `json:"repo_owner" validate:"required,min=1,max=255"`
	RepoName  string  `json:"repo_name" validate:"required,min=1,max=255"`
	Branch    string  `json:"branch" validate:"required,min=1,max=255"`
	RootDir   *string `json:"root_dir,omitempty" validate:"omitempty,max=500"`
}

// CreateServiceRequest represents the request body for creating a service
type CreateServiceRequest struct {
	Name         string         `json:"name" validate:"required,min=1,max=255"`
	Type         string         `json:"type" validate:"required,oneof=app database volume"`
	InstanceSize string         `json:"instance_size,omitempty" validate:"omitempty,oneof=small medium large xlarge"`
	Port         *int           `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	GitSourceID  *string        `json:"git_source_id,omitempty"`
	GitSource    *GitSourceInfo `json:"git_source,omitempty"`
	CanvasX      *int            `json:"canvas_x,omitempty"`
	CanvasY      *int            `json:"canvas_y,omitempty"`
}

// UpdateServiceRequest represents the request body for updating a service
type UpdateServiceRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Type         *string `json:"type,omitempty" validate:"omitempty,oneof=app database volume"`
	InstanceSize *string `json:"instance_size,omitempty" validate:"omitempty,oneof=small medium large xlarge"`
	Port         *int    `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	Status       *string `json:"status,omitempty" validate:"omitempty,oneof=pending provisioning building deploying live failed stopped"`
	
	// Git source updates
	Branch  *string `json:"branch,omitempty" validate:"omitempty,min=1,max=255"`
	RootDir *string `json:"root_dir,omitempty" validate:"omitempty,max=500"`
	
	// Resource limits
	CPULimit    *string `json:"cpu_limit,omitempty" validate:"omitempty"`
	MemoryLimit *string `json:"memory_limit,omitempty" validate:"omitempty"`
	
	// Build config
	StartCommand *string `json:"start_command,omitempty" validate:"omitempty,max=1000"`
	BuildCommand *string `json:"build_command,omitempty" validate:"omitempty,max=1000"`
}

// UpdateServicePositionRequest represents the request body for updating canvas position
type UpdateServicePositionRequest struct {
	X int `json:"x" validate:"required"`
	Y int `json:"y" validate:"required"`
}

