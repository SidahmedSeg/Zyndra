'use client'

import { useState } from 'react'
import { Github, Database, HardDrive, ArrowRight, ArrowLeft } from 'lucide-react'
import { useProjectsStore } from '@/stores/projectsStore'
import GitHubRepoSelection from './GitHubRepoSelection'
import DatabaseSelection from './DatabaseSelection'
import VolumeSelection from './VolumeSelection'

interface ServiceSelectionCardProps {
  onProjectCreated: (projectId: string) => void
}

type View = 'main' | 'github' | 'database' | 'volume'

export default function ServiceSelectionCard({ onProjectCreated }: ServiceSelectionCardProps) {
  const [currentView, setCurrentView] = useState<View>('main')
  const [projectId, setProjectId] = useState<string | null>(null)
  const { createProject } = useProjectsStore()

  const handleServiceSelect = async (type: 'github' | 'database' | 'volume') => {
    // If no project exists yet, create one first
    if (!projectId) {
      const project = await createProject({
        name: 'New Project',
        description: 'Project created from service selection',
      })
      setProjectId(project.id)
    }

    setCurrentView(type)
  }

  const handleBack = () => {
    setCurrentView('main')
  }

  const handleServiceCreated = () => {
    if (projectId) {
      onProjectCreated(projectId)
    }
  }

  return (
    <div className="w-full max-w-2xl">
      <div className="bg-white rounded-2xl shadow-2xl overflow-hidden transition-all duration-300">
        <div className="relative h-[500px]">
          {/* Main view */}
          {currentView === 'main' && (
            <div className="p-8 h-full flex flex-col">
              <div className="flex-1 flex flex-col justify-center">
                <div className="space-y-4">
                  <button
                    onClick={() => handleServiceSelect('github')}
                    className="w-full flex items-center justify-between p-6 border-2 border-gray-200 rounded-xl hover:border-blue-500 hover:bg-blue-50 transition-all group"
                  >
                    <div className="flex items-center gap-4">
                      <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                        <Github className="w-6 h-6 text-white" />
                      </div>
                      <span className="text-lg font-semibold text-gray-900">Github Repo</span>
                    </div>
                    <ArrowRight className="w-5 h-5 text-gray-400 group-hover:text-blue-600 transition-colors" />
                  </button>

                  <button
                    onClick={() => handleServiceSelect('database')}
                    className="w-full flex items-center justify-between p-6 border-2 border-gray-200 rounded-xl hover:border-blue-500 hover:bg-blue-50 transition-all group"
                  >
                    <div className="flex items-center gap-4">
                      <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-green-500 to-emerald-600 flex items-center justify-center">
                        <Database className="w-6 h-6 text-white" />
                      </div>
                      <span className="text-lg font-semibold text-gray-900">Database</span>
                    </div>
                    <ArrowRight className="w-5 h-5 text-gray-400 group-hover:text-blue-600 transition-colors" />
                  </button>

                  <button
                    onClick={() => handleServiceSelect('volume')}
                    className="w-full flex items-center justify-between p-6 border-2 border-gray-200 rounded-xl hover:border-blue-500 hover:bg-blue-50 transition-all group"
                  >
                    <div className="flex items-center gap-4">
                      <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-orange-500 to-red-600 flex items-center justify-center">
                        <HardDrive className="w-6 h-6 text-white" />
                      </div>
                      <span className="text-lg font-semibold text-gray-900">Volume</span>
                    </div>
                    <ArrowRight className="w-5 h-5 text-gray-400 group-hover:text-blue-600 transition-colors" />
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* GitHub Repo view */}
          {currentView === 'github' && (
            <div className="h-full flex flex-col animate-slide-in">
              <div className="p-6 border-b border-gray-200 flex items-center gap-4">
                <button
                  onClick={handleBack}
                  className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  <ArrowLeft className="w-5 h-5 text-gray-600" />
                </button>
                <h2 className="text-xl font-semibold text-gray-900">Select GitHub Repository</h2>
              </div>
              <div className="flex-1 overflow-hidden">
                <GitHubRepoSelection
                  projectId={projectId!}
                  onServiceCreated={handleServiceCreated}
                />
              </div>
            </div>
          )}

          {/* Database view */}
          {currentView === 'database' && (
            <div className="h-full flex flex-col animate-slide-in">
              <div className="p-6 border-b border-gray-200 flex items-center gap-4">
                <button
                  onClick={handleBack}
                  className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  <ArrowLeft className="w-5 h-5 text-gray-600" />
                </button>
                <h2 className="text-xl font-semibold text-gray-900">Select Database</h2>
              </div>
              <div className="flex-1 overflow-hidden">
                <DatabaseSelection
                  projectId={projectId!}
                  onServiceCreated={handleServiceCreated}
                />
              </div>
            </div>
          )}

          {/* Volume view */}
          {currentView === 'volume' && (
            <div className="h-full flex flex-col animate-slide-in">
              <div className="p-6 border-b border-gray-200 flex items-center gap-4">
                <button
                  onClick={handleBack}
                  className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  <ArrowLeft className="w-5 h-5 text-gray-600" />
                </button>
                <h2 className="text-xl font-semibold text-gray-900">Create Volume</h2>
              </div>
              <div className="flex-1 overflow-hidden">
                <VolumeSelection
                  projectId={projectId!}
                  onServiceCreated={handleServiceCreated}
                />
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

