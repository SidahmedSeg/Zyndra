'use client'

import { useState, useRef, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { ChevronDown, User } from 'lucide-react'
import { useProjectsStore } from '@/stores/projectsStore'

interface CanvasHeaderProps {
  projectId: string
  onEnvironmentChange?: (env: string) => void
}

export default function CanvasHeader({ projectId, onEnvironmentChange }: CanvasHeaderProps) {
  const router = useRouter()
  const { projects, selectedProject, fetchProjects, setSelectedProject } = useProjectsStore()
  const [projectDropdownOpen, setProjectDropdownOpen] = useState(false)
  const [envDropdownOpen, setEnvDropdownOpen] = useState(false)
  const [currentEnv, setCurrentEnv] = useState('Production')
  const projectDropdownRef = useRef<HTMLDivElement>(null)
  const envDropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (projectId) {
      if (!selectedProject || selectedProject.id !== projectId) {
        fetchProjects().then(() => {
          const project = projects.find(p => p.id === projectId)
          if (project) {
            setSelectedProject(project)
          }
        })
      }
    }
  }, [projectId, selectedProject, projects, fetchProjects, setSelectedProject])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (projectDropdownRef.current && !projectDropdownRef.current.contains(event.target as Node)) {
        setProjectDropdownOpen(false)
      }
      if (envDropdownRef.current && !envDropdownRef.current.contains(event.target as Node)) {
        setEnvDropdownOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleProjectSelect = (projectId: string) => {
    router.push(`/canvas/${projectId}`)
    setProjectDropdownOpen(false)
  }

  const handleEnvironmentSelect = (env: string) => {
    setCurrentEnv(env)
    setEnvDropdownOpen(false)
    onEnvironmentChange?.(env)
  }

  const handleAddProject = () => {
    router.push('/')
    setProjectDropdownOpen(false)
  }

  const environments = ['Production', 'Staging', 'Development']

  return (
    <header className="w-full bg-white border-b border-gray-200 h-12 flex items-center justify-between" style={{ paddingLeft: '20px', paddingRight: '20px' }}>
      {/* Left side */}
      <div className="flex items-center gap-3">
        <h1 className="text-lg font-bold text-gray-900">Zyndra</h1>
        
        {/* Project dropdown */}
        <div className="relative" ref={projectDropdownRef}>
          <button
            onClick={() => setProjectDropdownOpen(!projectDropdownOpen)}
            className="flex items-center gap-1 px-2 py-1 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded transition-colors"
          >
            <span>{selectedProject?.name || 'Loading...'}</span>
            <ChevronDown className="w-4 h-4" />
          </button>
          
          {projectDropdownOpen && (
            <div className="absolute top-full left-0 mt-1 bg-white border border-gray-200 rounded-md shadow-lg z-50 min-w-[200px]">
              {projects.map((project) => (
                <button
                  key={project.id}
                  onClick={() => handleProjectSelect(project.id)}
                  className={`w-full text-left px-3 py-2 text-sm hover:bg-gray-50 transition-colors ${
                    project.id === projectId ? 'bg-blue-50 text-blue-700' : 'text-gray-700'
                  }`}
                >
                  {project.name}
                </button>
              ))}
              <div className="border-t border-gray-200">
                <button
                  onClick={handleAddProject}
                  className="w-full text-left px-3 py-2 text-sm text-blue-600 hover:bg-blue-50 transition-colors"
                >
                  + Add new project
                </button>
              </div>
            </div>
          )}
        </div>

        <span className="text-gray-400">/</span>

        {/* Environment dropdown */}
        <div className="relative" ref={envDropdownRef}>
          <button
            onClick={() => setEnvDropdownOpen(!envDropdownOpen)}
            className="flex items-center gap-1 px-2 py-1 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded transition-colors"
          >
            <span>{currentEnv}</span>
            <ChevronDown className="w-4 h-4" />
          </button>
          
          {envDropdownOpen && (
            <div className="absolute top-full left-0 mt-1 bg-white border border-gray-200 rounded-md shadow-lg z-50 min-w-[150px]">
              {environments.map((env) => (
                <button
                  key={env}
                  onClick={() => handleEnvironmentSelect(env)}
                  className={`w-full text-left px-3 py-2 text-sm hover:bg-gray-50 transition-colors ${
                    env === currentEnv ? 'bg-blue-50 text-blue-700' : 'text-gray-700'
                  }`}
                >
                  {env}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Right side */}
      <div className="flex items-center gap-3">
        <button className="px-2 py-1 text-sm text-gray-700 hover:bg-gray-100 rounded transition-colors">
          Settings
        </button>
        <button className="px-2 py-1 text-sm text-gray-700 hover:bg-gray-100 rounded transition-colors">
          Logs
        </button>
        <button className="px-2 py-1 text-sm text-gray-700 hover:bg-gray-100 rounded transition-colors">
          Architecture
        </button>
        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white">
          <User className="w-4 h-4" />
        </div>
      </div>
    </header>
  )
}

