import { apiClient } from './client'

export interface Database {
  id: string
  project_id?: string
  service_id?: string
  engine: string
  version?: string
  size: string
  volume_id?: string
  volume_size_mb: number
  internal_hostname?: string
  internal_ip?: string
  port?: number
  username?: string
  password?: string
  database_name?: string
  connection_url?: string
  status: string
  created_at: string
}

export interface DatabaseCredentials {
  id: string
  engine: string
  hostname: string
  port: number
  username: string
  password: string
  database: string
  connection_url: string
}

export interface CreateDatabaseRequest {
  service_id?: string
  engine: string
  version?: string
  size?: string
  volume_size_mb?: number
}

export const databasesApi = {
  listByProject: (projectId: string) =>
    apiClient.get<Database[]>(`/projects/${projectId}/databases`),

  get: (id: string) => apiClient.get<Database>(`/databases/${id}`),

  getCredentials: (id: string) =>
    apiClient.get<DatabaseCredentials>(`/databases/${id}/credentials`),

  create: (projectId: string, data: CreateDatabaseRequest) =>
    apiClient.post<Database>(`/projects/${projectId}/databases`, data),

  delete: (id: string) => apiClient.delete(`/databases/${id}`),
}

