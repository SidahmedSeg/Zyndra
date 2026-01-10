'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import dynamic from 'next/dynamic'
import { useProjectsStore } from '@/stores/projectsStore'

// ReactFlow doesn't work with SSR, so we need to dynamically import it
const Canvas = dynamic(() => import('@/components/Canvas/Canvas'), {
  ssr: false,
  loading: () => (
    <div className="flex items-center justify-center h-screen bg-gray-50">
      <div className="text-center">
        <div className="animate-spin w-8 h-8 border-4 border-gray-300 border-t-indigo-600 rounded-full mx-auto mb-4" />
        <p className="text-gray-500">Loading canvas...</p>
      </div>
    </div>
  ),
})

export default function CanvasPage() {
  const params = useParams()
  const projectId = params.projectId as string
  const { selectedProject, fetchProject } = useProjectsStore()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

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

  if (!mounted) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-50">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-4 border-gray-300 border-t-indigo-600 rounded-full mx-auto mb-4" />
          <p className="text-gray-500">Loading...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="w-full h-screen flex flex-col">
      <Canvas projectId={projectId} />
    </div>
  )
}

