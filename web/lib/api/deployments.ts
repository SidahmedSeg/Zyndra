import { apiClient } from './client'

export interface Deployment {
  id: string
  service_id: string
  commit_sha?: string
  commit_message?: string
  commit_author?: string
  status: string
  image_tag?: string
  build_duration?: number
  deploy_duration?: number
  error_message?: string
  triggered_by: string
  started_at?: string
  finished_at?: string
  created_at: string
}

export interface DeploymentLog {
  id: number
  deployment_id: string
  timestamp: string
  phase: string
  level: string
  message: string
  metadata?: any
}

export interface TriggerDeploymentRequest {
  git_source_id?: string
  commit_sha?: string
  triggered_by: string
}

export const deploymentsApi = {
  listByService: (serviceId: string, limit?: number, offset?: number) => {
    const params = new URLSearchParams()
    if (limit) params.append('limit', limit.toString())
    if (offset) params.append('offset', offset.toString())
    return apiClient.get<Deployment[]>(
      `/services/${serviceId}/deployments?${params.toString()}`
    )
  },

  get: (id: string) => apiClient.get<Deployment>(`/deployments/${id}`),

  getLogs: (id: string, limit?: number, offset?: number) => {
    const params = new URLSearchParams()
    if (limit) params.append('limit', limit.toString())
    if (offset) params.append('offset', offset.toString())
    return apiClient.get<DeploymentLog[]>(
      `/deployments/${id}/logs?${params.toString()}`
    )
  },

  trigger: (serviceId: string, data: TriggerDeploymentRequest) =>
    apiClient.post<Deployment>(`/services/${serviceId}/deploy`, data),

  cancel: (id: string) => apiClient.post(`/deployments/${id}/cancel`),
}

