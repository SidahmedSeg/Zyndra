import { apiClient } from './client'

export interface Project {
  id: string
  name: string
  slug: string
  description?: string
  casdoor_org_id: string
  openstack_tenant_id?: string
  openstack_network_id?: string
  created_at: string
  updated_at: string
  service_count?: number
}

export interface CreateProjectRequest {
  name: string
  description?: string
}

export interface UpdateProjectRequest {
  name?: string
  description?: string
}

export const projectsApi = {
  list: () => apiClient.get<Project[]>('/projects'),

  get: (id: string) => apiClient.get<Project>(`/projects/${id}`),

  create: (data: CreateProjectRequest) => apiClient.post<Project>('/projects', data),

  update: (id: string, data: UpdateProjectRequest) =>
    apiClient.patch<Project>(`/projects/${id}`, data),

  delete: (id: string) => apiClient.delete(`/projects/${id}`),
}

