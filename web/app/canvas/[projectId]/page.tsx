'use client'

import { useEffect } from 'react'
import { useParams } from 'next/navigation'
import Canvas from '@/components/Canvas/Canvas'
import { useProjectsStore } from '@/stores/projectsStore'

export default function CanvasPage() {
  const params = useParams()
  const projectId = params.projectId as string
  const { selectedProject, fetchProject } = useProjectsStore()

  useEffect(() => {
    if (projectId && !selectedProject) {
      fetchProject(projectId)
    }
  }, [projectId, selectedProject, fetchProject])

  if (!projectId) {
    return (
      <div className="flex items-center justify-center h-screen">
        <p className="text-gray-500">No project selected</p>
      </div>
    )
  }

  return (
    <div className="w-full h-screen">
      <Canvas projectId={projectId} />
    </div>
  )
}

