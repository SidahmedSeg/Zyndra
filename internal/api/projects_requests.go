package api

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name              string  `json:"name" validate:"required,min=1,max=255"`
	Description       *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	OpenStackTenantID string  `json:"openstack_tenant_id" validate:"required"`
	DefaultRegion     *string `json:"default_region,omitempty" validate:"omitempty,max=100"`
	AutoDeploy        *bool   `json:"auto_deploy,omitempty"`
}

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	Name          *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description   *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	DefaultRegion *string `json:"default_region,omitempty" validate:"omitempty,max=100"`
	AutoDeploy    *bool   `json:"auto_deploy,omitempty"`
}

