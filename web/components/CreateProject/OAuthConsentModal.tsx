'use client'

import * as Dialog from '@radix-ui/react-dialog'
import { X, Github, Gitlab, AlertCircle } from 'lucide-react'
import { gitApi } from '@/lib/api/git'
import { useState } from 'react'

interface OAuthConsentModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
  provider: 'github' | 'gitlab'
}

export default function OAuthConsentModal({
  open,
  onOpenChange,
  onSuccess,
  provider,
}: OAuthConsentModalProps) {
  const [isConnecting, setIsConnecting] = useState(false)

  const handleConnect = async () => {
    setIsConnecting(true)
    try {
      // Store callback to handle OAuth success
      if (typeof window !== 'undefined') {
        // Store a flag to detect OAuth completion
        sessionStorage.setItem('oauth_pending', 'true')
        sessionStorage.setItem('oauth_provider', provider)
        sessionStorage.setItem('oauth_from_page', window.location.pathname)
        
        // Open OAuth in popup window
        let popup: Window | null = null
        if (provider === 'github') {
          popup = await gitApi.connectGitHub()
        } else {
          popup = await gitApi.connectGitLab()
        }

        if (!popup) {
          throw new Error('Failed to open OAuth popup. Please check your popup blocker settings.')
        }

        // Poll for popup to close (OAuth complete)
        const checkPopup = setInterval(() => {
          if (popup?.closed) {
            clearInterval(checkPopup)
            setIsConnecting(false)
            
            // Check if OAuth was successful by checking sessionStorage
            const oauthPending = sessionStorage.getItem('oauth_pending')
            if (oauthPending === 'true') {
              // Still pending, might have been cancelled
              sessionStorage.removeItem('oauth_pending')
              sessionStorage.removeItem('oauth_provider')
              sessionStorage.removeItem('oauth_from_page')
            } else {
              // Success - callback page should have cleared the flag
              onSuccess()
              onOpenChange(false)
            }
          }
        }, 500)

        // Also listen for message from popup (if callback page sends it)
        const handleMessage = (event: MessageEvent) => {
          if (event.data === 'oauth-success' || event.data?.type === 'oauth-success') {
            clearInterval(checkPopup)
            setIsConnecting(false)
            onSuccess()
            onOpenChange(false)
            window.removeEventListener('message', handleMessage)
          }
        }
        window.addEventListener('message', handleMessage)

        // Cleanup after 10 minutes
        setTimeout(() => {
          clearInterval(checkPopup)
          window.removeEventListener('message', handleMessage)
          if (!popup?.closed) {
            popup?.close()
          }
          setIsConnecting(false)
        }, 10 * 60 * 1000)
      }
    } catch (error: any) {
      console.error('OAuth connection failed:', error)
      setIsConnecting(false)
      // Show error to user
      const errorMessage = error?.message || error?.details || 'Failed to connect. Please try again.'
      alert(errorMessage)
    }
  }

  const providerInfo = {
    github: {
      name: 'GitHub',
      icon: Github,
      description: 'Connect your GitHub account to access your repositories',
      permissions: [
        'Read repository contents',
        'Read repository metadata',
        'Read and write webhooks',
      ],
    },
    gitlab: {
      name: 'GitLab',
      icon: Gitlab,
      description: 'Connect your GitLab account to access your repositories',
      permissions: [
        'Read repository contents',
        'Read repository metadata',
        'Read and write webhooks',
      ],
    },
  }

  const info = providerInfo[provider]
  const Icon = info.icon

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 bg-black/50 z-50" />
        <Dialog.Content className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-lg shadow-xl w-full max-w-md z-50">
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <Dialog.Title className="text-xl font-semibold flex items-center gap-2">
                <Icon className="w-6 h-6" />
                <span>Connect {info.name}</span>
              </Dialog.Title>
              <Dialog.Close asChild>
                <button
                  className="text-gray-400 hover:text-gray-600 transition-colors"
                  aria-label="Close"
                >
                  <X className="w-5 h-5" />
                </button>
              </Dialog.Close>
            </div>

            <div className="mb-6">
              <p className="text-gray-600 mb-4">{info.description}</p>
              
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
                <div className="flex items-start gap-3">
                  <AlertCircle className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
                  <div className="flex-1">
                    <h4 className="font-medium text-blue-900 mb-2">Zyndra will request access to:</h4>
                    <ul className="space-y-1 text-sm text-blue-800">
                      {info.permissions.map((permission, idx) => (
                        <li key={idx} className="flex items-center gap-2">
                          <span className="w-1.5 h-1.5 rounded-full bg-blue-600"></span>
                          {permission}
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>

              <p className="text-sm text-gray-500">
                You&apos;ll be redirected to {info.name} to authorize Zyndra. After authorization, you&apos;ll be redirected back to continue.
              </p>
            </div>

            <div className="flex gap-3">
              <Dialog.Close asChild>
                <button className="flex-1 px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors">
                  Cancel
                </button>
              </Dialog.Close>
              <button
                onClick={handleConnect}
                disabled={isConnecting}
                className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
              >
                {isConnecting ? (
                  <>
                    <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                    <span>Connecting...</span>
                  </>
                ) : (
                  <span>Authorize {info.name}</span>
                )}
              </button>
            </div>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}

