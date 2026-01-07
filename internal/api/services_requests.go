package api

// CreateServiceRequest represents the request body for creating a service
type CreateServiceRequest struct {
	Name         string  `json:"name" validate:"required,min=1,max=255"`
	Type         string  `json:"type" validate:"required,oneof=app database volume"`
	InstanceSize string  `json:"instance_size,omitempty" validate:"omitempty,oneof=small medium large xlarge"`
	Port         *int    `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	GitSourceID  *string `json:"git_source_id,omitempty"`
	CanvasX      *int    `json:"canvas_x,omitempty"`
	CanvasY      *int    `json:"canvas_y,omitempty"`
}

// UpdateServiceRequest represents the request body for updating a service
type UpdateServiceRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Type         *string `json:"type,omitempty" validate:"omitempty,oneof=app database volume"`
	InstanceSize *string `json:"instance_size,omitempty" validate:"omitempty,oneof=small medium large xlarge"`
	Port         *int    `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	Status       *string `json:"status,omitempty" validate:"omitempty,oneof=pending provisioning building deploying live failed stopped"`
}

// UpdateServicePositionRequest represents the request body for updating canvas position
type UpdateServicePositionRequest struct {
	X int `json:"x" validate:"required"`
	Y int `json:"y" validate:"required"`
}

