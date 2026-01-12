'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useProjectsStore } from '@/stores/projectsStore'
import { apiClient } from '@/lib/api/client'
import AppHeader from '@/components/Header/AppHeader'
import CreateProjectDialog from '@/components/Header/CreateProjectDialog'
import { Plus, UserPlus, Settings } from 'lucide-react'

export default function Home() {
  const router = useRouter()
  const { projects, fetchProjects, createProject, loading } = useProjectsStore()
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  
  // Ensure projects is always an array (handle hydration from localStorage)
  const projectsList = Array.isArray(projects) ? projects : []

  useEffect(() => {
    // Check if user is authenticated
    const token = apiClient.getToken()
    if (!token) {
      router.push('/auth/login')
      return
    }
    
    // Only fetch projects if we have a token
    fetchProjects().catch((error) => {
      // If 401, redirect will happen in interceptor
      if (error.status !== 401) {
        console.error('Failed to fetch projects:', error)
      }
    })
  }, [fetchProjects, router])

  const handleCreateProject = async (name: string, description?: string) => {
    const project = await createProject({ name, description })
    router.push(`/canvas/${project.id}`)
  }

  const handleProjectClick = (projectId: string) => {
    router.push(`/canvas/${projectId}`)
  }

  const handleInvite = () => {
    // TODO: Implement invite functionality
    console.log('Invite clicked')
  }

  const handleSettings = () => {
    // TODO: Implement settings functionality
    console.log('Settings clicked')
  }

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
    
    if (diffDays === 0) return 'Created today'
    if (diffDays === 1) return 'Created yesterday'
    return `Created ${diffDays} days ago`
  }

  return (
    <div className="min-h-screen bg-[#f5f5f5]">
      {/* Unified Header */}
      <AppHeader variant="projects" />
      
      {/* Main content */}
      <main className="max-w-7xl mx-auto px-8 py-12">
        {/* Title section */}
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-2xl font-semibold text-gray-900">My Projects</h1>
          <div className="flex items-center gap-2">
            <button
              onClick={handleSettings}
              className="p-2 text-gray-400 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
              aria-label="Settings"
            >
              <Settings className="w-4 h-4" strokeWidth={1.5} />
            </button>
            <button
              onClick={handleInvite}
              className="flex items-center gap-1.5 px-3 py-2 text-indigo-600 bg-white border border-gray-200 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
            >
              <UserPlus className="w-4 h-4" strokeWidth={1.5} />
              <span>Invite</span>
            </button>
            <button
              onClick={() => router.push('/create-project')}
              className="flex items-center gap-1.5 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
            >
              <Plus className="w-4 h-4" strokeWidth={2} />
              <span>Create</span>
            </button>
          </div>
        </div>

        {/* Divider */}
        <div className="border-t border-gray-200 mb-6" />

        {/* Projects count */}
        <div className="mb-6">
          <span className="text-sm font-medium text-gray-900">
            {String(projectsList.length).padStart(2, '0')} Projects
          </span>
        </div>

        {/* Projects grid */}
        {loading ? (
          <div className="flex items-center justify-center py-24">
            <div className="flex items-center gap-3">
              <svg className="animate-spin h-5 w-5 text-indigo-600" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
              <span className="text-gray-500">Loading projects...</span>
            </div>
          </div>
        ) : projectsList.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-24">
            <div className="w-24 h-24 rounded-full bg-gray-100 flex items-center justify-center mb-6">
              <Plus className="w-12 h-12 text-gray-400" />
            </div>
            <h2 className="text-2xl font-semibold text-gray-900 mb-2">No projects yet</h2>
            <p className="text-gray-500 mb-6">Create your first project to get started</p>
            <button
              onClick={() => router.push('/create-project')}
              className="px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium"
            >
              Create Project
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            {projectsList.map((project) => (
              <div
                key={project.id}
                onClick={() => handleProjectClick(project.id)}
                className="bg-white rounded-lg border border-gray-200 p-5 cursor-pointer hover:border-gray-300 hover:shadow-sm transition-all min-h-[140px] flex flex-col"
              >
                <div className="flex-1">
                  <h3 className="text-sm font-semibold text-gray-900 mb-1">
                    {project.name}
                  </h3>
                  <p className="text-xs text-gray-400">
                    {formatDate(project.created_at)}
                  </p>
                </div>
                <div className="mt-auto pt-6">
                  <span className="text-xs text-indigo-600">
                    {String(project.service_count || 0).padStart(2, '0')} Services in use
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      <CreateProjectDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onCreateProject={handleCreateProject}
      />
    </div>
  )
}
