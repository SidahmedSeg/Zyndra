import { apiClient } from './client'

export interface Volume {
  id: string
  project_id: string
  name: string
  size_mb: number
  mount_path?: string
  attached_to_service_id?: string
  attached_to_database_id?: string
  openstack_volume_id?: string
  status: string
  volume_type: string
  created_at: string
}

export interface CreateVolumeRequest {
  name: string
  size_mb: number
  mount_path?: string
}

export interface AttachVolumeRequest {
  service_id: string
  mount_path: string
}

export const volumesApi = {
  listByProject: (projectId: string) =>
    apiClient.get<Volume[]>(`/projects/${projectId}/volumes`),

  get: (id: string) => apiClient.get<Volume>(`/volumes/${id}`),

  create: (projectId: string, data: CreateVolumeRequest) =>
    apiClient.post<Volume>(`/projects/${projectId}/volumes`, data),

  attach: (id: string, data: AttachVolumeRequest) =>
    apiClient.patch(`/volumes/${id}/attach`, data),

  detach: (id: string) => apiClient.patch(`/volumes/${id}/detach`),

  delete: (id: string) => apiClient.delete(`/volumes/${id}`),
}

