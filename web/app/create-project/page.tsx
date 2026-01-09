'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Bell, User, LayoutGrid, Github, Database, FileText, Package, Code, HardDrive, Plus, ChevronRight } from 'lucide-react'
import { useProjectsStore } from '@/stores/projectsStore'
import GitHubRepoSelection from '@/components/CreateProject/GitHubRepoSelection'
import DatabaseSelection from '@/components/CreateProject/DatabaseSelection'
import VolumeSelection from '@/components/CreateProject/VolumeSelection'

export default function CreateProjectPage() {
  const router = useRouter()
  const [selectedOption, setSelectedOption] = useState<string | null>(null)
  const [projectId, setProjectId] = useState<string | null>(null)
  const { createProject } = useProjectsStore()

  // Listen for back events from child components
  useEffect(() => {
    const handleBack = () => setSelectedOption(null)
    window.addEventListener('back-to-selection', handleBack)
    return () => window.removeEventListener('back-to-selection', handleBack)
  }, [])

  const handleOptionSelect = async (option: string) => {
    setSelectedOption(option)
    
    // Create project if not exists
    if (!projectId) {
      const project = await createProject({
        name: 'New Project',
        description: 'Project created from service selection',
      })
      setProjectId(project.id)
    }
  }

  const handleBack = () => {
    setSelectedOption(null)
  }

  const handleServiceCreated = () => {
    if (projectId) {
      router.push(`/canvas/${projectId}`)
    }
  }

  const deploymentOptions = [
    { id: 'github', label: 'GitHub Repository', icon: Github, color: 'text-gray-900' },
    { id: 'database', label: 'Database', icon: Database, color: 'text-gray-900' },
    { id: 'template', label: 'Template', icon: FileText, color: 'text-gray-900' },
    { id: 'docker', label: 'Docker Image', icon: Package, color: 'text-gray-900' },
    { id: 'function', label: 'Function', icon: Code, color: 'text-gray-900' },
    { id: 'bucket', label: 'Bucket', icon: HardDrive, color: 'text-gray-900' },
    { id: 'empty', label: 'Empty Project', icon: Plus, color: 'text-gray-900' },
  ]

  return (
    <div className="min-h-screen bg-gray-50" style={{
      backgroundImage: 'radial-gradient(circle, #e5e7eb 1px, transparent 1px)',
      backgroundSize: '20px 20px'
    }}>
      {/* Header */}
      <header className="border-b bg-white">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-gray-900 flex items-center justify-center">
                <div className="w-4 h-4 bg-white rounded-sm"></div>
              </div>
              <h1 className="text-lg font-semibold text-gray-900">New Project</h1>
            </div>
            <div className="flex items-center gap-4">
              <a href="/" className="text-sm text-gray-600 hover:text-gray-900">Dashboard</a>
              <button className="relative">
                <Bell className="w-5 h-5 text-gray-600" />
                <span className="absolute -top-1 -right-1 w-5 h-5 bg-blue-600 text-white text-xs rounded-full flex items-center justify-center">24</span>
              </button>
              <div className="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center overflow-hidden">
                <User className="w-5 h-5 text-gray-600" />
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-6 py-12">
        <div className="max-w-4xl mx-auto">
          {/* Icon and Title Section */}
          <div className="text-center mb-12">
            <div className="inline-flex items-center justify-center mb-6">
              <div className="relative w-24 h-24">
                <div className="absolute inset-0 grid grid-cols-2 gap-1">
                  <div className="w-full h-full bg-purple-600 rounded-tl-lg"></div>
                  <div className="w-full h-full bg-purple-600 rounded-tr-lg"></div>
                  <div className="w-full h-full bg-purple-600 rounded-bl-lg"></div>
                  <div className="w-full h-full bg-purple-600 rounded-br-lg relative">
                    <Plus className="absolute bottom-0 right-0 w-4 h-4 text-white" />
                  </div>
                </div>
              </div>
            </div>
            <h1 className="text-4xl font-bold text-gray-900 mb-3">New project</h1>
            <p className="text-lg text-gray-600">Let&apos;s deploy your first service to production</p>
          </div>

          {/* Deployment Options Card */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200">
            <div className="p-6 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">What would you like to deploy today?</h2>
            </div>
            
            <div className="max-h-[500px] overflow-y-auto">
              {selectedOption === null ? (
                <div className="divide-y divide-gray-200">
                  {deploymentOptions.map((option) => {
                    const Icon = option.icon
                    return (
                      <button
                        key={option.id}
                        onClick={() => handleOptionSelect(option.id)}
                        className="w-full flex items-center justify-between p-4 hover:bg-gray-50 transition-colors group"
                      >
                        <div className="flex items-center gap-3">
                          <Icon className={`w-5 h-5 ${option.color}`} />
                          <span className="text-sm font-medium text-gray-900">{option.label}</span>
                        </div>
                        <ChevronRight className="w-5 h-5 text-gray-400 group-hover:text-gray-600 transition-colors" />
                      </button>
                    )
                  })}
                </div>
              ) : selectedOption === 'github' && projectId ? (
                <GitHubRepoSelection
                  projectId={projectId}
                  onServiceCreated={handleServiceCreated}
                />
              ) : selectedOption === 'database' && projectId ? (
                <DatabaseSelection
                  projectId={projectId}
                  onServiceCreated={handleServiceCreated}
                />
              ) : selectedOption === 'bucket' && projectId ? (
                <VolumeSelection
                  projectId={projectId}
                  onServiceCreated={handleServiceCreated}
                />
              ) : null}
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
