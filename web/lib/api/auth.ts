import axios from 'axios'

// Get API URL from environment variable
const getApiBaseURL = (): string => {
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL
  }
  
  if (typeof window !== 'undefined') {
    const hostname = window.location.hostname
    if (hostname === 'zyndra.armonika.cloud') {
      return 'https://api.zyndra.armonika.cloud'
    }
  }
  
  return 'http://localhost:8080'
}

const API_BASE_URL = getApiBaseURL()

// Types
export interface User {
  id: string
  email: string
  name: string
  avatar_url?: string
  email_verified: boolean
  organization?: Organization
}

export interface Organization {
  id: string
  name: string
  slug: string
  role: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  expires_at: string
  token_type: string
  user: User
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface LoginRequest {
  email: string
  password: string
}

// Token storage keys
const ACCESS_TOKEN_KEY = 'auth_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const USER_KEY = 'auth_user'

// Auth API client (uses base URL, not /v1/click-deploy prefix)
const authClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Auth API functions
export const authApi = {
  // Register a new user
  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await authClient.post<AuthResponse>('/auth/register', data)
    const authData = response.data
    
    // Store tokens and user
    this.setTokens(authData.access_token, authData.refresh_token)
    this.setUser(authData.user)
    
    return authData
  },

  // Login user
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await authClient.post<AuthResponse>('/auth/login', data)
    const authData = response.data
    
    // Store tokens and user
    this.setTokens(authData.access_token, authData.refresh_token)
    this.setUser(authData.user)
    
    return authData
  },

  // Refresh tokens
  async refresh(): Promise<AuthResponse> {
    const refreshToken = this.getRefreshToken()
    if (!refreshToken) {
      throw new Error('No refresh token available')
    }

    const response = await authClient.post<AuthResponse>('/auth/refresh', {
      refresh_token: refreshToken,
    })
    const authData = response.data
    
    // Store new tokens and user
    this.setTokens(authData.access_token, authData.refresh_token)
    this.setUser(authData.user)
    
    return authData
  },

  // Logout user
  async logout(): Promise<void> {
    const refreshToken = this.getRefreshToken()
    if (refreshToken) {
      try {
        await authClient.post('/auth/logout', {
          refresh_token: refreshToken,
        })
      } catch (error) {
        // Ignore logout errors, still clear local tokens
        console.error('Logout error:', error)
      }
    }
    
    // Clear local tokens and user
    this.clearAuth()
  },

  // Get current user (requires auth)
  async getMe(): Promise<User> {
    const token = this.getAccessToken()
    if (!token) {
      throw new Error('Not authenticated')
    }

    const response = await authClient.get<User>('/auth/me', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    
    // Update stored user
    this.setUser(response.data)
    
    return response.data
  },

  // Token management
  setTokens(accessToken: string, refreshToken: string): void {
    if (typeof window === 'undefined') return
    localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  },

  getAccessToken(): string | null {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(ACCESS_TOKEN_KEY)
  },

  getRefreshToken(): string | null {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(REFRESH_TOKEN_KEY)
  },

  setUser(user: User): void {
    if (typeof window === 'undefined') return
    localStorage.setItem(USER_KEY, JSON.stringify(user))
  },

  getUser(): User | null {
    if (typeof window === 'undefined') return null
    const userJson = localStorage.getItem(USER_KEY)
    if (!userJson) return null
    try {
      return JSON.parse(userJson)
    } catch {
      return null
    }
  },

  clearAuth(): void {
    if (typeof window === 'undefined') return
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  },

  isAuthenticated(): boolean {
    return !!this.getAccessToken()
  },

  // Check if token is expired (simple check based on JWT structure)
  isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]))
      const exp = payload.exp * 1000 // Convert to milliseconds
      return Date.now() >= exp
    } catch {
      return true
    }
  },

  // Auto-refresh if token is about to expire
  async ensureValidToken(): Promise<string | null> {
    const accessToken = this.getAccessToken()
    if (!accessToken) return null

    // If token expires in less than 2 minutes, refresh it
    try {
      const payload = JSON.parse(atob(accessToken.split('.')[1]))
      const exp = payload.exp * 1000
      const twoMinutes = 2 * 60 * 1000

      if (Date.now() >= exp - twoMinutes) {
        const authData = await this.refresh()
        return authData.access_token
      }

      return accessToken
    } catch {
      return accessToken
    }
  },
}

