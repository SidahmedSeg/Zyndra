package api

import "github.com/intelifox/click-deploy/internal/domain"

// ValidateAddCustomDomainRequest validates an AddCustomDomainRequest
func ValidateAddCustomDomainRequest(req *AddCustomDomainRequest) *domain.AppError {
	if req.Domain == "" {
		return domain.NewValidationError("Domain is required")
	}

	// Basic hostname validation
	if len(req.Domain) > 255 {
		return domain.NewValidationError("Domain name too long")
	}

	// More validation can be added here (e.g., regex for valid hostname)

	return nil
}

