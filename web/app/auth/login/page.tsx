'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { apiClient, API_BASE_URL } from '@/lib/api/client'
import { authApi } from '@/lib/api/auth'

export default function LoginPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [useCustomAuth, setUseCustomAuth] = useState(true) // Default to custom auth

  useEffect(() => {
    // Check if already logged in
    if (authApi.isAuthenticated()) {
      router.push('/')
      return
    }
  }, [router])

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      await authApi.login({ email, password })
      router.push('/')
    } catch (err: any) {
      console.error('Login error:', err)
      setError(err?.response?.data?.message || err?.message || 'Login failed. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const handleMockLogin = async () => {
    setLoading(true)
    setError('')

    try {
      // Call mock login endpoint
      const response = await fetch(`${API_BASE_URL}/auth/mock/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) {
        throw new Error('Login failed')
      }

      const data = await response.json()
      
      // Store token using the auth API
      authApi.setTokens(data.access_token, data.refresh_token || '')
      
      // Redirect to home
      router.push('/')
    } catch (err: any) {
      console.error('Login error:', err)
      setError('Mock login failed. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#0a0a0a]">
      <div className="w-full max-w-md space-y-8 rounded-xl bg-[#1a1a1a] p-8 border border-gray-800">
        <div className="text-center">
          <div className="flex justify-center mb-4">
            <img src="/logo-zyndra.svg" alt="Zyndra" className="h-10" />
          </div>
          <h1 className="text-2xl font-semibold text-white">Welcome back</h1>
          <p className="mt-2 text-sm text-gray-400">Sign in to your account</p>
        </div>

        {error && (
          <div className="rounded-lg bg-red-500/10 border border-red-500/20 p-4 text-sm text-red-400">
            {error}
          </div>
        )}

        {/* Toggle between custom auth and mock auth */}
        <div className="flex items-center justify-center gap-4 text-sm">
          <button
            type="button"
            onClick={() => setUseCustomAuth(true)}
            className={`px-3 py-1 rounded-md ${useCustomAuth ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:text-white'}`}
          >
            Email Login
          </button>
          <button
            type="button"
            onClick={() => setUseCustomAuth(false)}
            className={`px-3 py-1 rounded-md ${!useCustomAuth ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:text-white'}`}
          >
            Mock Login
          </button>
        </div>

        {useCustomAuth ? (
          <form onSubmit={handleLogin} className="mt-8 space-y-6">
            <div className="space-y-4">
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-300">
                  Email address
                </label>
                <input
                  id="email"
                  name="email"
                  type="email"
                  autoComplete="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="mt-1 block w-full rounded-lg border border-gray-700 bg-[#0a0a0a] px-4 py-3 text-white placeholder-gray-500 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                  placeholder="you@example.com"
                />
              </div>
              <div>
                <label htmlFor="password" className="block text-sm font-medium text-gray-300">
                  Password
                </label>
                <input
                  id="password"
                  name="password"
                  type="password"
                  autoComplete="current-password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="mt-1 block w-full rounded-lg border border-gray-700 bg-[#0a0a0a] px-4 py-3 text-white placeholder-gray-500 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                  placeholder="••••••••"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full rounded-lg bg-indigo-600 px-4 py-3 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {loading ? 'Signing in...' : 'Sign in'}
            </button>
          </form>
        ) : (
          <div className="mt-8">
            <button
              onClick={handleMockLogin}
              disabled={loading}
              className="w-full rounded-lg bg-gray-700 px-4 py-3 text-sm font-semibold text-white shadow-sm hover:bg-gray-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-gray-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {loading ? 'Signing in...' : 'Sign in with Mock Auth'}
            </button>
            <p className="mt-4 text-center text-xs text-gray-500">
              Using mock authentication for development/testing
            </p>
          </div>
        )}

        <div className="text-center text-sm">
          <span className="text-gray-400">Don't have an account?</span>{' '}
          <Link href="/auth/register" className="text-indigo-400 hover:text-indigo-300 font-medium">
            Sign up
          </Link>
        </div>
      </div>
    </div>
  )
}
