'use client'

import { useState, useEffect } from 'react'
import * as Dialog from '@radix-ui/react-dialog'
import { X, Github, Plus } from 'lucide-react'
import { gitApi, type GitRepository } from '@/lib/api/git'

interface RepoSelectionModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSelectRepo: (repo: GitRepository) => void
  projectId: string
}

export default function RepoSelectionModal({
  open,
  onOpenChange,
  onSelectRepo,
  projectId,
}: RepoSelectionModalProps) {
  const [repos, setRepos] = useState<GitRepository[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (open) {
      loadRepositories()
    }
  }, [open])

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
    gitApi.connectGitHub()
  }

  const handleSelectRepo = (repo: GitRepository) => {
    onSelectRepo(repo)
    onOpenChange(false)
  }

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 bg-black/50 z-50" />
        <Dialog.Content className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[80vh] z-50 flex flex-col">
          <div className="p-6 border-b border-gray-200">
            <Dialog.Title className="text-xl font-semibold flex items-center gap-2">
              <Github className="w-5 h-5" />
              <span>Select GitHub Repository</span>
            </Dialog.Title>
          </div>

          <div className="flex-1 overflow-y-auto p-6">
            {loading ? (
              <div className="flex items-center justify-center py-12">
                <p className="text-gray-500">Loading repositories...</p>
              </div>
            ) : error ? (
              <div className="flex flex-col items-center justify-center py-12">
                <p className="text-gray-500 mb-4">{error}</p>
                <button
                  onClick={handleConfigureRepo}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
                >
                  <Plus className="w-4 h-4" />
                  <span>Configure repo</span>
                </button>
              </div>
            ) : repos.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12">
                <p className="text-gray-500 mb-4">No repositories found</p>
                <button
                  onClick={handleConfigureRepo}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
                >
                  <Plus className="w-4 h-4" />
                  <span>Configure repo</span>
                </button>
              </div>
            ) : (
              <div className="space-y-2">
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
                          <p className="text-sm text-gray-500 mt-1 line-clamp-2">
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

          <div className="p-6 border-t border-gray-200 flex justify-end">
            <button
              onClick={handleConfigureRepo}
              className="px-4 py-2 text-blue-600 hover:bg-blue-50 rounded-lg transition-colors flex items-center gap-2"
            >
              <Plus className="w-4 h-4" />
              <span>Configure repo</span>
            </button>
            <Dialog.Close asChild>
              <button className="ml-3 px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors">
                Cancel
              </button>
            </Dialog.Close>
          </div>

          <Dialog.Close asChild>
            <button
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-600 transition-colors"
              aria-label="Close"
            >
              <X className="w-5 h-5" />
            </button>
          </Dialog.Close>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}

