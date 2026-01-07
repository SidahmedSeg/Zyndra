import axios, { AxiosInstance, AxiosError } from 'axios'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export interface ApiError {
  code: string
  message: string
  details?: string
}

export class ApiClientError extends Error {
  code: string
  details?: string
  status?: number

  constructor(code: string, message: string, details?: string, status?: number) {
    super(message)
    this.name = 'ApiClientError'
    this.code = code
    this.details = details
    this.status = status
  }
}

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: `${API_BASE_URL}/v1/click-deploy`,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Request interceptor - add auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = this.getToken()
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    // Response interceptor - handle errors
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<ApiError>) => {
        if (error.response) {
          const apiError = error.response.data
          throw new ApiClientError(
            apiError?.code || 'UNKNOWN_ERROR',
            apiError?.message || error.message,
            apiError?.details,
            error.response.status
          )
        }
        throw new ApiClientError('NETWORK_ERROR', error.message || 'Network error')
      }
    )
  }

  private getToken(): string | null {
    if (typeof window === 'undefined') return null
    return localStorage.getItem('auth_token')
  }

  setToken(token: string) {
    if (typeof window === 'undefined') return
    localStorage.setItem('auth_token', token)
  }

  clearToken() {
    if (typeof window === 'undefined') return
    localStorage.removeItem('auth_token')
  }

  get<T>(url: string, config?: any) {
    return this.client.get<T>(url, config).then((res) => res.data)
  }

  post<T>(url: string, data?: any, config?: any) {
    return this.client.post<T>(url, data, config).then((res) => res.data)
  }

  patch<T>(url: string, data?: any, config?: any) {
    return this.client.patch<T>(url, data, config).then((res) => res.data)
  }

  delete<T>(url: string, config?: any) {
    return this.client.delete<T>(url, config).then((res) => res.data)
  }
}

export const apiClient = new ApiClient()

