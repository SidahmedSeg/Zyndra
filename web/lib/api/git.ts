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

  connectGitHub: () => {
    // Redirect to GitHub OAuth
    const baseURL = getBaseURL()
    window.location.href = `${baseURL}/v1/click-deploy/git/connect/github`
  },

  connectGitLab: () => {
    // Redirect to GitLab OAuth
    const baseURL = getBaseURL()
    window.location.href = `${baseURL}/v1/click-deploy/git/connect/gitlab`
  },

  deleteConnection: (id: string) => apiClient.delete(`/git/connections/${id}`),
}

