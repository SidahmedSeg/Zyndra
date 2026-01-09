'use client'

import { useState, useEffect } from 'react'
import { Github, Plus, Loader2 } from 'lucide-react'
import { gitApi, type GitRepository } from '@/lib/api/git'
import { useServicesStore } from '@/stores/servicesStore'
import OAuthConsentModal from './OAuthConsentModal'

interface GitHubRepoSelectionProps {
  projectId: string
  onServiceCreated: () => void
}

export default function GitHubRepoSelection({ projectId, onServiceCreated }: GitHubRepoSelectionProps) {
  const [repos, setRepos] = useState<GitRepository[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [oauthModalOpen, setOAuthModalOpen] = useState(false)
  const { createService } = useServicesStore()

  useEffect(() => {
    loadRepositories()
  }, [])

  const loadRepositories = async () => {
    setLoading(true)
    setError(null)
    try {
      const repositories = await gitApi.listRepositories('github')
      setRepos(Array.isArray(repositories) ? repositories : [])
    } catch (err: any) {
      if (err.status === 404 || err.message?.includes('No connection')) {
        setError('No GitHub connection found')
      } else {
        setError(err.message || 'Failed to load repositories')
      }
    } finally {
      setLoading(false)
    }
  }

  const handleConfigureRepo = () => {
    setOAuthModalOpen(true)
  }

  const handleOAuthSuccess = () => {
    setOAuthModalOpen(false)
    // Reload repositories after OAuth success (will be called from callback page)
    // For now, just reload
    setTimeout(() => {
      loadRepositories()
    }, 3000)
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
      <div className="flex-1 overflow-y-auto p-6">
        {loading ? (
          <div className="flex items-center justify-center h-full">
            <Loader2 className="w-6 h-6 text-gray-400 animate-spin" />
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center h-full">
            <p className="text-gray-500 mb-4">{error}</p>
            <button
              onClick={handleConfigureRepo}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
            >
              <Plus className="w-4 h-4" />
              <span>Configure new repo</span>
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {repos.map((repo) => (
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
                      {repo.private && <span>Private</span>}
                    </div>
                  </div>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>

      <div className="p-6 border-t border-gray-200">
        <button
          onClick={handleConfigureRepo}
          className="w-full px-4 py-2 text-blue-600 hover:bg-blue-50 rounded-lg transition-colors flex items-center justify-center gap-2 border border-blue-200"
        >
          <Plus className="w-4 h-4" />
          <span>Configure new repo</span>
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

