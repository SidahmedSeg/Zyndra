'use client'

import { Suspense, useEffect, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { CheckCircle, XCircle, Loader2 } from 'lucide-react'

function GitCallbackContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [message, setMessage] = useState('Completing GitHub connection...')

  useEffect(() => {
    const checkOAuthStatus = async () => {
      // Check if we have a pending OAuth from sessionStorage
      const oauthPending = sessionStorage.getItem('oauth_pending')
      const oauthProvider = sessionStorage.getItem('oauth_provider')

      if (oauthPending === 'true') {
        // Poll for connection status
        const checkConnection = async () => {
          try {
            const { gitApi } = await import('@/lib/api/git')
            const connections = await gitApi.listConnections()
            const hasConnection = connections.some(
              (conn) => conn.provider === oauthProvider
            )

            if (hasConnection) {
              setStatus('success')
              setMessage('GitHub connection successful!')
              sessionStorage.removeItem('oauth_pending')
              sessionStorage.removeItem('oauth_provider')

              // Check if we're in a popup window
              if (window.opener) {
                // Send message to parent window
                window.opener.postMessage({ type: 'oauth-success' }, window.location.origin)
                // Close popup
                window.close()
              } else {
                // Redirect after a short delay (not in popup)
                setTimeout(() => {
                  // Check if we're on create-project page, stay there, otherwise go to home
                  const fromPage = sessionStorage.getItem('oauth_from_page')
                  if (fromPage) {
                    sessionStorage.removeItem('oauth_from_page')
                    router.push(fromPage)
                  } else {
                    router.push('/')
                  }
                }, 2000)
              }
            } else {
              // Keep checking
              setTimeout(checkConnection, 1000)
            }
          } catch (error) {
            console.error('Error checking connection:', error)
            setStatus('error')
            setMessage('Failed to verify connection. Please try again.')
          }
        }

        // Start checking after a short delay to allow backend to process
        setTimeout(checkConnection, 2000)
      } else {
        // No pending OAuth, might be direct callback
        setStatus('success')
        setMessage('Connection completed!')
        
        // Check if we're in a popup window
        if (window.opener) {
          window.opener.postMessage({ type: 'oauth-success' }, window.location.origin)
          window.close()
        } else {
          setTimeout(() => router.push('/'), 2000)
        }
      }
    }

    checkOAuthStatus()
  }, [router])

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="text-center p-8 bg-white rounded-lg shadow-lg max-w-md">
        {status === 'loading' && (
          <>
            <Loader2 className="w-12 h-12 text-blue-600 animate-spin mx-auto mb-4" />
            <p className="text-lg text-gray-700">{message}</p>
          </>
        )}
        {status === 'success' && (
          <>
            <CheckCircle className="w-12 h-12 text-green-600 mx-auto mb-4" />
            <p className="text-lg text-gray-700">{message}</p>
            <p className="text-sm text-gray-500 mt-2">Redirecting...</p>
          </>
        )}
        {status === 'error' && (
          <>
            <XCircle className="w-12 h-12 text-red-600 mx-auto mb-4" />
            <p className="text-lg text-gray-700">{message}</p>
            <button
              onClick={() => router.push('/')}
              className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              Go to Home
            </button>
          </>
        )}
      </div>
    </div>
  )
}

export default function GitCallback() {
  return (
    <Suspense fallback={
      <div className="flex min-h-screen items-center justify-center">
        <Loader2 className="w-8 h-8 text-blue-600 animate-spin" />
      </div>
    }>
      <GitCallbackContent />
    </Suspense>
  )
}

