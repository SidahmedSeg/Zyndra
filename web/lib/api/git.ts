import { apiClient } from './client'

export interface GitRepository {
  id: number
  name: string
  full_name: string
  owner: string
  description?: string
  private: boolean
  default_branch: string
  url?: string
  clone_url?: string
}

export interface GitConnection {
  id: string
  provider: string
  account_name?: string
  created_at: string
}

// Helper to get base URL for OAuth redirects
const getBaseURL = (): string => {
  if (typeof window === 'undefined') return 'http://localhost:8080'
  
  // Get from environment variable first
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL
  }
  
  // Fallback: infer from current origin
  const hostname = window.location.hostname
  if (hostname === 'zyndra.armonika.cloud') {
    return 'https://api.zyndra.armonika.cloud'
  }
  
  // Default fallback
  return 'http://localhost:8080'
}

export const gitApi = {
  listConnections: () => apiClient.get<GitConnection[]>('/git/connections'),

  listRepositories: (provider: string = 'github') =>
    apiClient.get<GitRepository[]>(`/git/repos?provider=${provider}`),

  connectGitHub: async () => {
    // Get OAuth URL via authenticated API call, then redirect
    try {
      const response = await apiClient.get<{ auth_url: string }>('/git/connect/github/url')
      window.location.href = response.auth_url
    } catch (error) {
      console.error('Failed to get GitHub OAuth URL:', error)
      throw error
    }
  },

  connectGitLab: async () => {
    // Get OAuth URL via authenticated API call, then redirect
    try {
      const response = await apiClient.get<{ auth_url: string }>('/git/connect/gitlab/url')
      window.location.href = response.auth_url
    } catch (error) {
      console.error('Failed to get GitLab OAuth URL:', error)
      throw error
    }
  },

  deleteConnection: (id: string) => apiClient.delete(`/git/connections/${id}`),
}

