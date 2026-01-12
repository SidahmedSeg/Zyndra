'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { ChevronRight, Database, Container, Settings, Loader2 } from 'lucide-react'
import { useProjectsStore } from '@/stores/projectsStore'
import { gitApi, type GitRepository, type GitHubAppInstallation } from '@/lib/api/git'
import { useServicesStore } from '@/stores/servicesStore'
import { servicesApi } from '@/lib/api/services'
import AppHeader from '@/components/Header/AppHeader'

export default function CreateProjectPage() {
  const router = useRouter()
  const [selectedOption, setSelectedOption] = useState<string | null>(null)
  const [projectId, setProjectId] = useState<string | null>(null)
  const { createProject } = useProjectsStore()

  // GitHub App state
  const [repos, setRepos] = useState<GitRepository[]>([])
  const [installations, setInstallations] = useState<GitHubAppInstallation[]>([])
  const [selectedInstallation, setSelectedInstallation] = useState<GitHubAppInstallation | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { createService } = useServicesStore()

  // Listen for back events
  useEffect(() => {
    const handleBack = () => setSelectedOption(null)
    window.addEventListener('back-to-selection', handleBack)
    return () => window.removeEventListener('back-to-selection', handleBack)
  }, [])

  // Listen for GitHub App installation messages
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.origin !== window.location.origin) return
      if (event.data?.type === 'github-app-installed') {
        console.log('GitHub App installed:', event.data)
        setTimeout(() => loadInstallations(), 1500)
      }
    }
    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [])

  // Load installations when GitHub option is selected
  useEffect(() => {
    if (selectedOption === 'github') {
      loadInstallations()
    }
  }, [selectedOption])

  const loadInstallations = async () => {
    setLoading(true)
    setError(null)
    try {
      const installs = await gitApi.listGitHubAppInstallations()
      setInstallations(Array.isArray(installs) ? installs : [])
      
      if (installs.length === 1) {
        setSelectedInstallation(installs[0])
        await loadRepositoriesForInstallation(installs[0].id)
      } else if (installs.length === 0) {
        setError('No GitHub App installation found')
      }
    } catch (err: any) {
      console.error('Failed to load installations:', err)
      setError('No GitHub App installation found. Click below to configure.')
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

  const handleConfigureGitHubApp = async () => {
    try {
      const popup = await gitApi.installGitHubApp()
      if (popup) {
        const checkPopup = setInterval(() => {
          if (popup.closed) {
            clearInterval(checkPopup)
            setTimeout(() => loadInstallations(), 2000)
          }
        }, 500)
      }
    } catch (err) {
      console.error('Failed to open GitHub App installation:', err)
    }
  }

  const handleOptionSelect = async (option: string) => {
    setSelectedOption(option)
    
    if (!projectId) {
      const project = await createProject({
        name: 'New Project',
        description: 'Project created from service selection',
      })
      setProjectId(project.id)
    }
  }

  const [deploying, setDeploying] = useState(false)
  const [deployingRepo, setDeployingRepo] = useState<string | null>(null)

  const handleSelectRepo = async (repo: GitRepository) => {
    if (!projectId) return
    
    setDeploying(true)
    setDeployingRepo(repo.id?.toString() || repo.name)
    
    try {
      const owner = repo.owner || repo.full_name.split('/')[0]
      const repoName = repo.name

      // Create the service with git source
      const service = await createService(projectId, {
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

      // Trigger deployment
      try {
        await servicesApi.triggerDeployment(service.id)
      } catch (deployErr) {
        console.warn('Failed to trigger deployment (build queue might not be running):', deployErr)
        // Don't block navigation, deployment can be triggered manually
      }

      // Navigate to canvas
      router.push(`/canvas/${projectId}`)
    } catch (error) {
      console.error('Failed to create service from repo:', error)
      setError('Failed to create service. Please try again.')
      setDeploying(false)
      setDeployingRepo(null)
    }
  }

  return (
    <div className="min-h-screen bg-[#f8f8f8]" style={{
      backgroundImage: 'radial-gradient(circle, #d4d4d4 1px, transparent 1px)',
      backgroundSize: '24px 24px'
    }}>
      {/* Header */}
      <AppHeader variant="canvas" projectId={projectId || undefined} />

      {/* Main Content */}
      <main className="flex items-center justify-center min-h-[calc(100vh-73px)]">
        <div className="w-full max-w-xl px-6">
          {selectedOption === null ? (
            <>
              {/* New Project Icon */}
              <div className="flex justify-center mb-6">
                <img src="/newproject-icon.svg" alt="" className="w-14 h-14" />
              </div>

              {/* Title */}
              <h1 className="text-3xl font-semibold text-gray-900 text-center mb-2">
                New project
              </h1>
              
              {/* Subtitle */}
              <p className="text-gray-500 text-center mb-8">
                Let&apos;s deploy your service to production
              </p>

              {/* Options Card */}
              <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="divide-y divide-gray-100">
                  {/* GitHub repository option */}
                  <button
                    onClick={() => handleOptionSelect('github')}
                    className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group"
                  >
                    <div className="flex items-center gap-3">
                      <img 
                        src="/github-icon.svg" 
                        alt="" 
                        className="w-5 h-5 opacity-60 group-hover:opacity-100 transition-opacity" 
                      />
                      <span className="text-sm text-gray-500 group-hover:text-gray-900 group-hover:font-medium transition-all">
                        Github repository
                      </span>
                    </div>
                    <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-gray-600 transition-colors" />
                  </button>

                  {/* Database option */}
                  <button
                    onClick={() => handleOptionSelect('database')}
                    className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group"
                  >
                    <div className="flex items-center gap-3">
                      <Database className="w-5 h-5 text-gray-400 group-hover:text-gray-900 transition-colors" />
                      <span className="text-sm text-gray-500 group-hover:text-gray-900 group-hover:font-medium transition-all">
                        Database
                      </span>
                    </div>
                    <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-gray-600 transition-colors" />
                  </button>

                  {/* Docker image option */}
                  <button
                    onClick={() => handleOptionSelect('docker')}
                    className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group"
                  >
                    <div className="flex items-center gap-3">
                      <Container className="w-5 h-5 text-gray-400 group-hover:text-gray-900 transition-colors" />
                      <span className="text-sm text-gray-500 group-hover:text-gray-900 group-hover:font-medium transition-all">
                        Docker image
                      </span>
                    </div>
                    <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-gray-600 transition-colors" />
                  </button>
                </div>
              </div>
            </>
          ) : selectedOption === 'github' ? (
            <>
              {/* GitHub Icon */}
              <div className="flex justify-center mb-6">
                <img src="/github-icon.svg" alt="" className="w-16 h-16" />
              </div>

              {/* Title */}
              <h1 className="text-3xl font-semibold text-gray-900 text-center mb-2">
                Deploy new repository
              </h1>
              
              {/* Subtitle */}
              <p className="text-gray-500 text-center mb-6">
                Select or configure a GitHub repository to deploy
              </p>

              {/* Badge */}
              <div className="flex justify-center mb-4">
                <span className="px-3 py-1 bg-indigo-100 text-indigo-700 text-sm font-medium rounded-full">
                  Github repositories
                </span>
              </div>

              {/* Repositories Card */}
              <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="divide-y divide-gray-100">
                  {/* Configure GitHub App option */}
                  <button
                    onClick={handleConfigureGitHubApp}
                    className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group"
                  >
                    <div className="flex items-center gap-3">
                      <Settings className="w-5 h-5 text-gray-400 group-hover:text-gray-900 transition-colors" />
                      <span className="text-sm text-gray-700 group-hover:text-gray-900 group-hover:font-medium transition-all">
                        Configure Github App
                      </span>
                    </div>
                    <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-gray-600 transition-colors" />
                  </button>

                  {/* Loading state */}
                  {loading && (
                    <div className="px-5 py-4 text-center">
                      <div className="animate-spin w-5 h-5 border-2 border-gray-300 border-t-gray-600 rounded-full mx-auto" />
                    </div>
                  )}

                  {/* Repositories list */}
                  {!loading && repos.map((repo) => {
                    const repoKey = repo.id?.toString() || repo.name
                    const isDeployingThis = deploying && deployingRepo === repoKey
                    
                    return (
                      <button
                        key={repo.id}
                        onClick={() => handleSelectRepo(repo)}
                        disabled={deploying}
                        className={`w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group ${
                          deploying ? 'cursor-not-allowed opacity-60' : ''
                        }`}
                      >
                        <div className="flex items-center gap-3">
                          {isDeployingThis ? (
                            <Loader2 className="w-5 h-5 text-indigo-600 animate-spin" />
                          ) : (
                            <img 
                              src="/github-icon.svg" 
                              alt="" 
                              className="w-5 h-5 opacity-50 group-hover:opacity-100 transition-opacity" 
                            />
                          )}
                          <span className="text-sm text-gray-500 group-hover:text-gray-900 transition-colors">
                            {repo.full_name || repo.name}
                          </span>
                        </div>
                        {isDeployingThis && (
                          <span className="text-xs text-indigo-600 font-medium">
                            Deploying...
                          </span>
                        )}
                      </button>
                    )
                  })}

                  {/* No repos message */}
                  {!loading && repos.length === 0 && !error && installations.length > 0 && (
                    <div className="px-5 py-4 text-center text-sm text-gray-500">
                      No repositories found. Configure GitHub App to add repositories.
                    </div>
                  )}

                  {/* Error/empty state */}
                  {!loading && error && repos.length === 0 && (
                    <div className="px-5 py-4 text-center text-sm text-gray-500">
                      {error}
                    </div>
                  )}
                </div>
              </div>

              {/* Back button */}
              <button
                onClick={() => {
                  setSelectedOption(null)
                  setRepos([])
                  setInstallations([])
                  setError(null)
                }}
                className="mt-6 w-full text-center text-sm text-gray-500 hover:text-gray-700 transition-colors"
              >
                ← Back to service selection
              </button>
            </>
          ) : selectedOption === 'database' ? (
            <>
              {/* Database Icon */}
              <div className="flex justify-center mb-6">
                <Database className="w-16 h-16 text-gray-400" />
              </div>

              {/* Title */}
              <h1 className="text-3xl font-semibold text-gray-900 text-center mb-2">
                Deploy database
              </h1>
              
              {/* Subtitle */}
              <p className="text-gray-500 text-center mb-8">
                Select a database type to deploy
              </p>

              {/* Database Options Card */}
              <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="divide-y divide-gray-100">
                  <button className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group">
                    <div className="flex items-center gap-3">
                      <span className="text-sm text-gray-500 group-hover:text-gray-900 group-hover:font-medium transition-all">
                        PostgreSQL
                      </span>
                    </div>
                    <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-gray-600 transition-colors" />
                  </button>
                  <button className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group">
                    <div className="flex items-center gap-3">
                      <span className="text-sm text-gray-500 group-hover:text-gray-900 group-hover:font-medium transition-all">
                        MySQL
                      </span>
                    </div>
                    <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-gray-600 transition-colors" />
                  </button>
                  <button className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group">
                    <div className="flex items-center gap-3">
                      <span className="text-sm text-gray-500 group-hover:text-gray-900 group-hover:font-medium transition-all">
                        MongoDB
                      </span>
                    </div>
                    <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-gray-600 transition-colors" />
                  </button>
                  <button className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group">
                    <div className="flex items-center gap-3">
                      <span className="text-sm text-gray-500 group-hover:text-gray-900 group-hover:font-medium transition-all">
                        Redis
                      </span>
                    </div>
                    <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-gray-600 transition-colors" />
                  </button>
                </div>
              </div>

              {/* Back button */}
              <button
                onClick={() => setSelectedOption(null)}
                className="mt-6 w-full text-center text-sm text-gray-500 hover:text-gray-700 transition-colors"
              >
                ← Back to service selection
              </button>
            </>
          ) : null}
        </div>
      </main>
    </div>
  )
}
