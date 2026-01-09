'use client'

import { Suspense, useEffect, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { CheckCircle, XCircle, Loader2 } from 'lucide-react'

function GitSetupContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [message, setMessage] = useState('Processing GitHub App installation...')

  useEffect(() => {
    const handleInstallation = async () => {
      const installationId = searchParams.get('installation_id')
      const setupAction = searchParams.get('setup_action')

      console.log('GitHub App Setup:', { installationId, setupAction })

      if (!installationId) {
        setStatus('error')
        setMessage('Missing installation ID. Please try installing the GitHub App again.')
        return
      }

      // Store the installation ID for later use
      if (typeof window !== 'undefined') {
        sessionStorage.setItem('github_app_installation_id', installationId)
        sessionStorage.setItem('github_app_setup_action', setupAction || 'install')
      }

      // If we're in a popup, send message to parent and close
      if (window.opener) {
        try {
          window.opener.postMessage({
            type: 'github-app-installed',
            installationId,
            setupAction: setupAction || 'install'
          }, window.location.origin)
          
          setStatus('success')
          setMessage('GitHub App installed successfully!')
          
          // Close popup after short delay
          setTimeout(() => window.close(), 1000)
        } catch (error) {
          console.error('Error sending message to opener:', error)
          // Continue with redirect if message fails
          handleRedirect()
        }
      } else {
        // Not in a popup, redirect to appropriate page
        handleRedirect()
      }
    }

    const handleRedirect = () => {
      setStatus('success')
      setMessage('GitHub App installed successfully! Redirecting...')
      
      // Check if there's a stored page to return to
      const fromPage = sessionStorage.getItem('oauth_from_page')
      
      setTimeout(() => {
        if (fromPage) {
          sessionStorage.removeItem('oauth_from_page')
          router.push(fromPage)
        } else {
          // Default to create-project page
          router.push('/create-project')
        }
      }, 2000)
    }

    handleInstallation()
  }, [router, searchParams])

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
            <p className="text-sm text-gray-500 mt-2">
              {window.opener ? 'Closing...' : 'Redirecting...'}
            </p>
          </>
        )}
        {status === 'error' && (
          <>
            <XCircle className="w-12 h-12 text-red-600 mx-auto mb-4" />
            <p className="text-lg text-gray-700">{message}</p>
            <button
              onClick={() => router.push('/create-project')}
              className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              Go to Create Project
            </button>
          </>
        )}
      </div>
    </div>
  )
}

export default function GitSetup() {
  return (
    <Suspense fallback={
      <div className="flex min-h-screen items-center justify-center">
        <Loader2 className="w-8 h-8 text-blue-600 animate-spin" />
      </div>
    }>
      <GitSetupContent />
    </Suspense>
  )
}

