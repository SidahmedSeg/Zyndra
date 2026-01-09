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

  connectGitHub: async (): Promise<Window | null> => {
    // Get OAuth URL via authenticated API call, then open in popup
    try {
      const response = await apiClient.get<{ auth_url: string }>('/git/connect/github/url')
      
      // Open OAuth in a popup window
      const width = 600
      const height = 700
      const left = (window.screen.width - width) / 2
      const top = (window.screen.height - height) / 2
      
      const popup = window.open(
        response.auth_url,
        'github-oauth',
        `width=${width},height=${height},left=${left},top=${top},toolbar=no,menubar=no,scrollbars=yes,resizable=yes`
      )
      
      return popup
    } catch (error) {
      console.error('Failed to get GitHub OAuth URL:', error)
      throw error
    }
  },

  connectGitLab: async (): Promise<Window | null> => {
    // Get OAuth URL via authenticated API call, then open in popup
    try {
      const response = await apiClient.get<{ auth_url: string }>('/git/connect/gitlab/url')
      
      // Open OAuth in a popup window
      const width = 600
      const height = 700
      const left = (window.screen.width - width) / 2
      const top = (window.screen.height - height) / 2
      
      const popup = window.open(
        response.auth_url,
        'gitlab-oauth',
        `width=${width},height=${height},left=${left},top=${top},toolbar=no,menubar=no,scrollbars=yes,resizable=yes`
      )
      
      return popup
    } catch (error) {
      console.error('Failed to get GitLab OAuth URL:', error)
      throw error
    }
  },

  deleteConnection: (id: string) => apiClient.delete(`/git/connections/${id}`),
}

