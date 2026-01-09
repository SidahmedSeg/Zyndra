'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useProjectsStore } from '@/stores/projectsStore'
import { apiClient } from '@/lib/api/client'
import Header from '@/components/Header/Header'
import CreateProjectDialog from '@/components/Header/CreateProjectDialog'
import ProjectCard from '@/components/Projects/ProjectCard'
import { Plus } from 'lucide-react'

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

  return (
    <div className="min-h-screen bg-gray-50">
      <Header
        onCreateProject={() => setCreateDialogOpen(true)}
        onInvite={handleInvite}
        onSettings={handleSettings}
      />
      
      <main className="container mx-auto px-6 py-8">
        {loading ? (
          <div className="flex items-center justify-center py-24">
            <p className="text-gray-500">Loading projects...</p>
          </div>
        ) : projectsList.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-24">
            <div className="w-24 h-24 rounded-full bg-gray-100 flex items-center justify-center mb-6">
              <Plus className="w-12 h-12 text-gray-400" />
            </div>
            <h2 className="text-2xl font-semibold text-gray-900 mb-2">No projects yet</h2>
            <p className="text-gray-500 mb-6">Create your first project to get started</p>
            <button
              onClick={() => setCreateDialogOpen(true)}
              className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
            >
              Create Project
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {projectsList.map((project) => (
              <ProjectCard
                key={project.id}
                project={project}
                onClick={() => handleProjectClick(project.id)}
              />
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

