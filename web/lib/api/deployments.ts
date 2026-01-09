import { apiClient } from './client'

export interface Deployment {
  id: string
  service_id: string
  commit_sha?: string
  commit_message?: string
  commit_author?: string
  status: 'queued' | 'building' | 'pushing' | 'deploying' | 'success' | 'failed' | 'cancelled'
  image_tag?: string
  build_duration?: number
  deploy_duration?: number
  error_message?: string
  triggered_by: 'webhook' | 'manual' | 'rollback'
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
  metadata?: Record<string, any>
}

export interface TriggerDeploymentRequest {
  commit_sha?: string
  branch?: string
  triggered_by?: string
}

// Map backend status to UI-friendly status
export const getStatusDisplay = (status: Deployment['status']): { label: string; color: string } => {
  switch (status) {
    case 'queued':
      return { label: 'Initializing', color: 'text-yellow-500' }
    case 'building':
      return { label: 'Building image', color: 'text-blue-500' }
    case 'pushing':
      return { label: 'Deploying', color: 'text-purple-500' }
    case 'deploying':
      return { label: 'Post deploy', color: 'text-indigo-500' }
    case 'success':
      return { label: 'Online', color: 'text-green-500' }
    case 'failed':
      return { label: 'Failed', color: 'text-red-500' }
    case 'cancelled':
      return { label: 'Cancelled', color: 'text-gray-500' }
    default:
      return { label: status, color: 'text-gray-400' }
  }
}

export const deploymentsApi = {
  // Trigger a new deployment for a service
  trigger: (serviceId: string, data?: TriggerDeploymentRequest) =>
    apiClient.post<Deployment>(`/services/${serviceId}/deploy`, data || {}),

  // Get a deployment by ID
  get: (deploymentId: string) =>
    apiClient.get<Deployment>(`/deployments/${deploymentId}`),

  // Get deployment logs
  getLogs: (deploymentId: string, limit?: number) =>
    apiClient.get<DeploymentLog[]>(`/deployments/${deploymentId}/logs${limit ? `?limit=${limit}` : ''}`),

  // Cancel a deployment
  cancel: (deploymentId: string) =>
    apiClient.post<void>(`/deployments/${deploymentId}/cancel`),

  // List deployments for a service
  listByService: (serviceId: string, limit?: number) =>
    apiClient.get<Deployment[]>(`/services/${serviceId}/deployments${limit ? `?limit=${limit}` : ''}`),
}
