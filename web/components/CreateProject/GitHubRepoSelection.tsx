'use client'

import { useState, useEffect } from 'react'
import { Github, Plus, Loader2, RefreshCw } from 'lucide-react'
import { gitApi, type GitRepository, type GitHubAppInstallation } from '@/lib/api/git'
import { useServicesStore } from '@/stores/servicesStore'
import OAuthConsentModal from './OAuthConsentModal'

interface GitHubRepoSelectionProps {
  projectId: string
  onServiceCreated: () => void
}

export default function GitHubRepoSelection({ projectId, onServiceCreated }: GitHubRepoSelectionProps) {
  const [repos, setRepos] = useState<GitRepository[]>([])
  const [installations, setInstallations] = useState<GitHubAppInstallation[]>([])
  const [selectedInstallation, setSelectedInstallation] = useState<GitHubAppInstallation | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [oauthModalOpen, setOAuthModalOpen] = useState(false)
  const { createService } = useServicesStore()

  useEffect(() => {
    loadInstallations()
  }, [])

  const loadInstallations = async () => {
    setLoading(true)
    setError(null)
    try {
      const installs = await gitApi.listGitHubAppInstallations()
      setInstallations(Array.isArray(installs) ? installs : [])
      
      // If there's only one installation, auto-select it and load repos
      if (installs.length === 1) {
        setSelectedInstallation(installs[0])
        await loadRepositoriesForInstallation(installs[0].id)
      } else if (installs.length === 0) {
        setError('No GitHub App installation found')
      }
    } catch (err: any) {
      console.error('Failed to load installations:', err)
      if (err.message?.includes('not configured')) {
        setError('GitHub App is not configured')
      } else {
        setError('No GitHub App installation found. Click below to install.')
      }
    } finally {
      setLoading(false)
    }
  }

  const loadRepositoriesForInstallation = async (installationId: number) => {
    setLoading(true)
    try {
      const repositories = await gitApi.listGitHubAppInstallationRepos(installationId)
      setRepos(Array.isArray(repositories) ? repositories : [])
    } catch (err: any) {
      console.error('Failed to load repos:', err)
      setError(err.message || 'Failed to load repositories')
    } finally {
      setLoading(false)
    }
  }

  const handleSelectInstallation = async (installation: GitHubAppInstallation) => {
    setSelectedInstallation(installation)
    await loadRepositoriesForInstallation(installation.id)
  }

  const handleConfigureRepo = async () => {
    try {
      const popup = await gitApi.installGitHubApp()
      if (popup) {
        // Poll for popup to close
        const checkPopup = setInterval(() => {
          if (popup.closed) {
            clearInterval(checkPopup)
            // Reload installations after popup closes
            setTimeout(() => loadInstallations(), 2000)
          }
        }, 500)
      }
    } catch (err) {
      console.error('Failed to open GitHub App installation:', err)
      setOAuthModalOpen(true) // Fallback to OAuth modal
    }
  }

  const handleOAuthSuccess = () => {
    setOAuthModalOpen(false)
    setTimeout(() => loadInstallations(), 3000)
  }

  const handleSelectRepo = async (repo: GitRepository) => {
    try {
      const owner = repo.owner || repo.full_name.split('/')[0]
      const repoName = repo.name

      await createService(projectId, {
        name: repoName,
        type: 'app',
        instance_size: 'small',
        port: 8080,
        git_source: {
          provider: 'github',
          repo_owner: owner,
          repo_name: repoName,
          branch: repo.default_branch || 'main',
        },
      })

      onServiceCreated()
    } catch (error) {
      console.error('Failed to create service from repo:', error)
    }
  }

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-gray-200 flex items-center gap-3">
        <button
          onClick={() => {
            const event = new CustomEvent('back-to-selection')
            window.dispatchEvent(event)
          }}
          className="p-1 hover:bg-gray-100 rounded transition-colors"
        >
          <svg className="w-5 h-5 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h3 className="text-sm font-medium text-gray-900">GitHub Repository</h3>
      </div>
      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <div className="flex items-center justify-center h-full">
            <Loader2 className="w-6 h-6 text-gray-400 animate-spin" />
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <Github className="w-12 h-12 text-gray-300 mb-4" />
            <p className="text-gray-500 mb-4">{error}</p>
            <button
              onClick={handleConfigureRepo}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
            >
              <Plus className="w-4 h-4" />
              <span>Install GitHub App</span>
            </button>
            <p className="text-xs text-gray-400 mt-3 max-w-xs">
              You can choose which specific repositories to grant access to
            </p>
          </div>
        ) : installations.length > 1 && !selectedInstallation ? (
          // Multiple installations - let user choose
          <div className="space-y-3">
            <p className="text-sm text-gray-600 mb-3">Select an account:</p>
            {installations.map((inst) => (
              <button
                key={inst.id}
                onClick={() => handleSelectInstallation(inst)}
                className="w-full text-left p-4 border border-gray-200 rounded-lg hover:bg-gray-50 hover:border-blue-300 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <Github className="w-5 h-5 text-gray-600" />
                  <div className="flex-1">
                    <h3 className="font-medium text-gray-900">{inst.account_login}</h3>
                    <p className="text-xs text-gray-400">{inst.account_type}</p>
                  </div>
                </div>
              </button>
            ))}
          </div>
        ) : (
          <div className="space-y-3">
            {selectedInstallation && installations.length > 1 && (
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm text-gray-600">
                  Repos from: <strong>{selectedInstallation.account_login}</strong>
                </span>
                <button
                  onClick={() => {
                    setSelectedInstallation(null)
                    setRepos([])
                  }}
                  className="text-xs text-blue-600 hover:underline"
                >
                  Change
                </button>
              </div>
            )}
            {repos.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-gray-500 mb-4">No repositories found in this installation</p>
                <button
                  onClick={handleConfigureRepo}
                  className="text-blue-600 hover:underline text-sm"
                >
                  Modify GitHub App permissions
                </button>
              </div>
            ) : (
              repos.map((repo) => (
                <button
                  key={repo.id}
                  onClick={() => handleSelectRepo(repo)}
                  className="w-full text-left p-4 border border-gray-200 rounded-lg hover:bg-gray-50 hover:border-blue-300 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <Github className="w-5 h-5 text-gray-600" />
                    <div className="flex-1">
                      <h3 className="font-medium text-gray-900">{repo.full_name}</h3>
                      {repo.description && (
                        <p className="text-sm text-gray-500 mt-1 line-clamp-1">
                          {repo.description}
                        </p>
                      )}
                      <div className="flex items-center gap-4 mt-2 text-xs text-gray-400">
                        <span>Branch: {repo.default_branch || 'main'}</span>
                        {repo.private && <span className="text-orange-500">Private</span>}
                      </div>
                    </div>
                  </div>
                </button>
              ))
            )}
          </div>
        )}
      </div>

      <div className="p-4 border-t border-gray-200 flex gap-2">
        <button
          onClick={handleConfigureRepo}
          className="flex-1 px-4 py-2 text-blue-600 hover:bg-blue-50 rounded transition-colors flex items-center justify-center gap-2 text-sm"
        >
          <Plus className="w-4 h-4" />
          <span>Add more repos</span>
        </button>
        <button
          onClick={() => loadInstallations()}
          className="px-3 py-2 text-gray-600 hover:bg-gray-100 rounded transition-colors"
          title="Refresh"
        >
          <RefreshCw className="w-4 h-4" />
        </button>
      </div>

      <OAuthConsentModal
        open={oauthModalOpen}
        onOpenChange={setOAuthModalOpen}
        onSuccess={handleOAuthSuccess}
        provider="github"
      />
    </div>
  )
}

