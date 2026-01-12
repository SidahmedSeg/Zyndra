import { apiClient } from './client'

export interface Service {
  id: string
  project_id: string
  git_source_id?: string
  name: string
  type: string
  status: string
  instance_size: string
  port?: number
  
  // Git source info
  repo_owner?: string
  repo_name?: string
  branch?: string
  root_dir?: string
  
  // Resource limits
  cpu_limit?: string
  memory_limit?: string
  
  // Build config
  start_command?: string
  build_command?: string
  
  // Infrastructure
  openstack_instance_id?: string
  openstack_fip_id?: string
  openstack_fip_address?: string
  security_group_id?: string
  subdomain?: string
  generated_url?: string
  current_image_tag?: string
  
  // Canvas position
  canvas_x?: number
  canvas_y?: number
  
  // Deployment status
  deployment_status?: 'idle' | 'initializing' | 'building' | 'pushing' | 'deploying' | 'post_deploy' | 'online' | 'failed'
  
  created_at: string
  updated_at: string
}

export interface GitSourceInfo {
  provider: string
  repo_owner: string
  repo_name: string
  branch: string
  root_dir?: string
}

export interface CreateServiceRequest {
  name: string
  type: string
  instance_size: string
  port?: number
  git_source?: GitSourceInfo
  canvas_x?: number
  canvas_y?: number
}

export interface UpdateServiceRequest {
  name?: string
  instance_size?: string
  port?: number
  branch?: string
  root_dir?: string
  cpu_limit?: string
  memory_limit?: string
  start_command?: string
  build_command?: string
}

export interface UpdateServicePositionRequest {
  canvas_x: number
  canvas_y: number
}

export const servicesApi = {
  listByProject: (projectId: string) =>
    apiClient.get<Service[]>(`/projects/${projectId}/services`),

  get: (id: string) => apiClient.get<Service>(`/services/${id}`),

  create: (projectId: string, data: CreateServiceRequest) =>
    apiClient.post<Service>(`/projects/${projectId}/services`, data),

  update: (id: string, data: UpdateServiceRequest) =>
    apiClient.patch<Service>(`/services/${id}`, data),

  updatePosition: (id: string, data: UpdateServicePositionRequest) =>
    apiClient.patch<Service>(`/services/${id}/position`, data),

  delete: (id: string) => apiClient.delete(`/services/${id}`),

  // Trigger deployment for a service
  triggerDeployment: (serviceId: string, data?: { commit_sha?: string; branch?: string }) =>
    apiClient.post<any>(`/services/${serviceId}/deploy`, data || {}),
}

