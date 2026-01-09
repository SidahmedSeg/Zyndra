'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Bell, ChevronDown, ChevronRight, Github, Database, Container } from 'lucide-react'
import { useProjectsStore } from '@/stores/projectsStore'
import GitHubRepoSelection from '@/components/CreateProject/GitHubRepoSelection'
import DatabaseSelection from '@/components/CreateProject/DatabaseSelection'

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

  const handleServiceCreated = () => {
    if (projectId) {
      router.push(`/canvas/${projectId}`)
    }
  }

  return (
    <div className="min-h-screen bg-[#f8f8f8]" style={{
      backgroundImage: 'radial-gradient(circle, #d4d4d4 1px, transparent 1px)',
      backgroundSize: '24px 24px'
    }}>
      {/* Header */}
      <header className="bg-white border-b border-gray-200">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            {/* Left side - Logo and navigation */}
            <div className="flex items-center gap-4">
              {/* Logo */}
              <a href="/" className="flex-shrink-0">
                <img 
                  src="/logo-zyndra.svg" 
                  alt="Zyndra" 
                  className="h-[18px] w-auto"
                />
              </a>
              
              {/* Project name dropdown */}
              <button className="flex items-center gap-1.5 text-sm text-gray-700 hover:text-gray-900 transition-colors">
                <span>Project name</span>
                <ChevronDown className="w-4 h-4 text-gray-400" />
              </button>
              
              {/* Separator */}
              <span className="text-gray-300">/</span>
              
              {/* Environment dropdown */}
              <button className="flex items-center gap-1.5 text-sm text-gray-700 hover:text-gray-900 transition-colors">
                <span>Production</span>
                <ChevronDown className="w-4 h-4 text-gray-400" />
              </button>
            </div>

            {/* Right side - Navigation and user */}
            <div className="flex items-center gap-6">
              {/* Navigation links */}
              <nav className="flex items-center gap-6">
                <a href="#" className="text-sm text-gray-600 hover:text-gray-900 transition-colors">
                  Architecture
                </a>
                <a href="#" className="text-sm text-gray-600 hover:text-gray-900 transition-colors">
                  Logs
                </a>
                <a href="#" className="text-sm text-gray-600 hover:text-gray-900 transition-colors">
                  Settings
                </a>
              </nav>

              {/* Notification bell */}
              <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-full border border-gray-200 hover:bg-gray-50 transition-colors">
                <Bell className="w-4 h-4 text-gray-500" />
                <span className="text-sm text-gray-600">32</span>
              </button>

              {/* User avatar */}
              <div className="w-8 h-8 rounded-full bg-emerald-500 flex items-center justify-center overflow-hidden border-2 border-emerald-600">
                <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
                </svg>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex items-center justify-center min-h-[calc(100vh-73px)]">
        <div className="w-full max-w-md px-6">
          {/* Icon */}
          <div className="flex justify-center mb-6">
            <div className="relative">
              {/* 2x2 grid of squares */}
              <div className="grid grid-cols-2 gap-1">
                <div className="w-6 h-6 bg-indigo-600 rounded-sm"></div>
                <div className="w-6 h-6 bg-indigo-600 rounded-sm"></div>
                <div className="w-6 h-6 bg-indigo-600 rounded-sm"></div>
                <div className="w-6 h-6 bg-indigo-600 rounded-sm"></div>
              </div>
              {/* Small plus sign positioned at bottom-right */}
              <svg 
                className="absolute -bottom-1 -right-1 w-3.5 h-3.5 text-indigo-600" 
                fill="currentColor" 
                viewBox="0 0 24 24"
              >
                <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
              </svg>
            </div>
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
            {selectedOption === null ? (
              <div className="divide-y divide-gray-100">
                {/* GitHub repository option */}
                <button
                  onClick={() => handleOptionSelect('github')}
                  className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group"
                >
                  <div className="flex items-center gap-3">
                    <Github className="w-5 h-5 text-gray-700" />
                    <span className="text-sm font-medium text-gray-800">Github repository</span>
                  </div>
                  <ChevronRight className="w-4 h-4 text-gray-400 group-hover:text-gray-600 transition-colors" />
                </button>

                {/* Database option */}
                <button
                  onClick={() => handleOptionSelect('database')}
                  className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group"
                >
                  <div className="flex items-center gap-3">
                    <Database className="w-5 h-5 text-gray-500" />
                    <span className="text-sm text-gray-600">Database</span>
                  </div>
                  <ChevronRight className="w-4 h-4 text-gray-400 group-hover:text-gray-600 transition-colors" />
                </button>

                {/* Docker image option */}
                <button
                  onClick={() => handleOptionSelect('docker')}
                  className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group"
                >
                  <div className="flex items-center gap-3">
                    <Container className="w-5 h-5 text-gray-500" />
                    <span className="text-sm text-gray-600">Docker image</span>
                  </div>
                  <ChevronRight className="w-4 h-4 text-gray-400 group-hover:text-gray-600 transition-colors" />
                </button>
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
            ) : null}
          </div>
        </div>
      </main>
    </div>
  )
}
