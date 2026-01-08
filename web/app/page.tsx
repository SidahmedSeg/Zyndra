'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useProjectsStore } from '@/stores/projectsStore'
import { apiClient } from '@/lib/api/client'

export default function Home() {
  const router = useRouter()
  const { projects, fetchProjects, selectedProject } = useProjectsStore()

  useEffect(() => {
    // Check if user is authenticated
    const token = apiClient.getToken()
    if (!token) {
      router.push('/auth/login')
      return
    }
    
    fetchProjects()
  }, [fetchProjects, router])

  useEffect(() => {
    if (selectedProject) {
      router.push(`/canvas/${selectedProject.id}`)
    } else if (projects.length > 0) {
      router.push(`/canvas/${projects[0].id}`)
    }
  }, [selectedProject, projects, router])

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <div className="z-10 max-w-5xl w-full items-center justify-between font-mono text-sm">
        <h1 className="text-4xl font-bold mb-4">Click to Deploy</h1>
        <p className="text-lg mb-4">No-code deployment platform</p>
        {projects.length === 0 && (
          <p className="text-gray-500">Loading projects...</p>
        )}
      </div>
    </main>
  )
}

