import { apiClient } from './client'

export interface CustomDomain {
  id: string
  service_id: string
  domain: string
  status: 'pending' | 'verifying' | 'active' | 'failed'
  ssl_status: 'pending' | 'provisioning' | 'active' | 'failed'
  verification_token?: string
  created_at: string
  updated_at: string
}

export interface CreateCustomDomainRequest {
  domain: string
}

export interface VerifyDomainResponse {
  verified: boolean
  cname_target: string
  message?: string
}

export const customDomainsApi = {
  // List custom domains for a service
  listByService: (serviceId: string) =>
    apiClient.get<CustomDomain[]>(`/services/${serviceId}/domains`),

  // Add a custom domain to a service
  create: (serviceId: string, data: CreateCustomDomainRequest) =>
    apiClient.post<CustomDomain>(`/services/${serviceId}/domains`, data),

  // Verify domain DNS configuration (backend uses /domains/{id}/verify)
  verify: (domainId: string) =>
    apiClient.post<VerifyDomainResponse>(`/domains/${domainId}/verify`, {}),

  // Delete a custom domain (backend uses /domains/{id})
  delete: (domainId: string) =>
    apiClient.delete(`/domains/${domainId}`),
}
