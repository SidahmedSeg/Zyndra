import { apiClient } from './client'

export interface DataPoint {
  timestamp: string
  value: number
}

export interface MetricsResponse {
  cpu: DataPoint[]
  memory: DataPoint[]
  network_in: DataPoint[]
  network_out: DataPoint[]
  request_count?: DataPoint[]
  response_time?: DataPoint[]
  error_rate?: DataPoint[]
}

export const metricsApi = {
  getServiceMetrics: (serviceId: string, start?: string, end?: string, step?: string) => {
    const params = new URLSearchParams()
    if (start) params.append('start', start)
    if (end) params.append('end', end)
    if (step) params.append('step', step)
    return apiClient.get<MetricsResponse>(
      `/services/${serviceId}/metrics?${params.toString()}`
    )
  },

  getDatabaseMetrics: (databaseId: string, start?: string, end?: string, step?: string) => {
    const params = new URLSearchParams()
    if (start) params.append('start', start)
    if (end) params.append('end', end)
    if (step) params.append('step', step)
    return apiClient.get<MetricsResponse>(
      `/databases/${databaseId}/metrics?${params.toString()}`
    )
  },

  getVolumeMetrics: (volumeId: string, start?: string, end?: string, step?: string) => {
    const params = new URLSearchParams()
    if (start) params.append('start', start)
    if (end) params.append('end', end)
    if (step) params.append('step', step)
    return apiClient.get<MetricsResponse>(
      `/volumes/${volumeId}/metrics?${params.toString()}`
    )
  },
}

