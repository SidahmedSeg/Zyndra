'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { apiClient } from '@/lib/api/client'

export default function LoginPage() {
  const router = useRouter()

  useEffect(() => {
    // Check if already logged in
    const token = apiClient.getToken()
    if (token) {
      router.push('/')
      return
    }
  }, [router])

  const handleLogin = () => {
    // Get API base URL
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 
      (typeof window !== 'undefined' && window.location.hostname === 'zyndra.armonika.cloud'
        ? 'https://api.zyndra.armonika.cloud'
        : 'http://localhost:8080')
    
    // Redirect to backend OAuth initiation endpoint
    const loginURL = `${apiUrl}/auth/casdoor/login?redirect_uri=${encodeURIComponent(
      typeof window !== 'undefined' ? window.location.origin + '/auth/callback' : ''
    )}`
    
    window.location.href = loginURL
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-md space-y-8 rounded-lg bg-white p-8 shadow-lg">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900">Click to Deploy</h1>
          <p className="mt-2 text-sm text-gray-600">No-code deployment platform</p>
        </div>
        
        <div className="mt-8">
          <button
            onClick={handleLogin}
            className="w-full rounded-md bg-blue-600 px-4 py-3 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          >
            Sign in with Casdoor
          </button>
        </div>
      </div>
    </div>
  )
}

