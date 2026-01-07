import { apiClient } from './client'

export interface EnvVar {
  id: string
  service_id: string
  key: string
  value?: string
  is_secret: boolean
  linked_database_id?: string
  link_type?: string
  created_at: string
}

export interface CreateEnvVarRequest {
  key: string
  value?: string
  is_secret?: boolean
  linked_database_id?: string
  link_type?: string
}

export const envVarsApi = {
  listByService: (serviceId: string) =>
    apiClient.get<EnvVar[]>(`/services/${serviceId}/env`),

  create: (serviceId: string, data: CreateEnvVarRequest) =>
    apiClient.post<EnvVar>(`/services/${serviceId}/env`, data),

  update: (serviceId: string, key: string, data: Partial<CreateEnvVarRequest>) =>
    apiClient.patch<EnvVar>(`/services/${serviceId}/env/${key}`, data),

  delete: (serviceId: string, key: string) =>
    apiClient.delete(`/services/${serviceId}/env/${key}`),
}

