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

export interface GitHubAppInstallation {
  id: number
  account_login: string
  account_id: number
  account_type: string
  html_url: string
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

  // GitHub App methods (Railway-style per-repo access)
  getGitHubAppInstallURL: async (): Promise<{ install_url: string; app_name: string }> => {
    const response = await apiClient.get<{ install_url: string; app_name: string }>('/git/app/github/install-url')
    return response
  },

  installGitHubApp: async (): Promise<Window | null> => {
    // Get GitHub App install URL and open in popup
    try {
      const { install_url } = await apiClient.get<{ install_url: string }>('/git/app/github/install-url')
      
      // Open installation in a popup window
      const width = 1000
      const height = 800
      const left = (window.screen.width - width) / 2
      const top = (window.screen.height - height) / 2
      
      const popup = window.open(
        install_url,
        'github-app-install',
        `width=${width},height=${height},left=${left},top=${top},toolbar=no,menubar=no,scrollbars=yes,resizable=yes`
      )
      
      return popup
    } catch (error) {
      console.error('Failed to get GitHub App install URL:', error)
      throw error
    }
  },

  listGitHubAppInstallations: () =>
    apiClient.get<GitHubAppInstallation[]>('/git/app/github/installations'),

  listGitHubAppInstallationRepos: (installationId: number) =>
    apiClient.get<GitRepository[]>(`/git/app/github/installations/${installationId}/repos`),
}

