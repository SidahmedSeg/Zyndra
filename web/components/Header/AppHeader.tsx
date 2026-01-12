'use client'

import { useState, useRef, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { 
  ChevronDown, 
  Bell, 
  User, 
  Settings, 
  BarChart3, 
  LayoutGrid, 
  HelpCircle, 
  Moon, 
  LogOut 
} from 'lucide-react'
import { useProjectsStore } from '@/stores/projectsStore'
import { authApi } from '@/lib/api/auth'

type TabType = 'architecture' | 'logs' | 'settings'

interface AppHeaderProps {
  variant: 'projects' | 'canvas'
  projectId?: string
  onEnvironmentChange?: (env: string) => void
  activeTab?: TabType
  onTabChange?: (tab: TabType) => void
}

export default function AppHeader({ 
  variant, 
  projectId, 
  onEnvironmentChange,
  activeTab = 'architecture',
  onTabChange
}: AppHeaderProps) {
  const router = useRouter()
  const { projects, selectedProject, fetchProjects, setSelectedProject } = useProjectsStore()
  const [projectDropdownOpen, setProjectDropdownOpen] = useState(false)
  const [envDropdownOpen, setEnvDropdownOpen] = useState(false)
  const [avatarMenuOpen, setAvatarMenuOpen] = useState(false)
  const [currentEnv, setCurrentEnv] = useState('Production')
  const [currentTab, setCurrentTab] = useState<TabType>(activeTab)
  const [darkTheme, setDarkTheme] = useState(false)
  const projectDropdownRef = useRef<HTMLDivElement>(null)
  const envDropdownRef = useRef<HTMLDivElement>(null)
  const avatarMenuRef = useRef<HTMLDivElement>(null)

  // Mock user data - replace with actual user data from auth context
  const user = {
    email: 'user@example.com',
    role: 'org owner',
    plan: 'Pro'
  }

  useEffect(() => {
    if (variant === 'canvas' && projectId) {
      if (!selectedProject || selectedProject.id !== projectId) {
        fetchProjects().then(() => {
          const project = projects.find(p => p.id === projectId)
          if (project) {
            setSelectedProject(project)
          }
        })
      }
    }
  }, [variant, projectId, selectedProject, projects, fetchProjects, setSelectedProject])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (projectDropdownRef.current && !projectDropdownRef.current.contains(event.target as Node)) {
        setProjectDropdownOpen(false)
      }
      if (envDropdownRef.current && !envDropdownRef.current.contains(event.target as Node)) {
        setEnvDropdownOpen(false)
      }
      if (avatarMenuRef.current && !avatarMenuRef.current.contains(event.target as Node)) {
        setAvatarMenuOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleProjectSelect = (id: string) => {
    router.push(`/canvas/${id}`)
    setProjectDropdownOpen(false)
  }

  const handleEnvironmentSelect = (env: string) => {
    setCurrentEnv(env)
    setEnvDropdownOpen(false)
    onEnvironmentChange?.(env)
  }

  const handleAddProject = () => {
    router.push('/create-project')
    setProjectDropdownOpen(false)
  }

  const handleTabClick = (tab: TabType) => {
    setCurrentTab(tab)
    onTabChange?.(tab)
  }

  const handleLogout = () => {
    authApi.logout()
    router.push('/auth/login')
  }

  const handleDarkThemeToggle = () => {
    setDarkTheme(!darkTheme)
    // TODO: Implement actual dark theme toggle
  }

  const environments = ['Production', 'Staging', 'Development']

  return (
    <header className="bg-white border-b border-gray-200">
      <div className="px-6">
        <div className="flex items-center justify-between h-14">
          {/* Left side - Logo and navigation */}
          <div className="flex items-center gap-4">
            <a href="/" className="flex-shrink-0">
              <img src="/logo-zyndra.svg" alt="Zyndra" className="h-[18px] w-auto" />
            </a>
            
            {variant === 'canvas' && (
              <>
                {/* Project dropdown */}
                <div className="relative" ref={projectDropdownRef}>
                  <button
                    onClick={() => setProjectDropdownOpen(!projectDropdownOpen)}
                    className="flex items-center gap-1.5 text-sm text-gray-700 hover:text-gray-900 transition-colors"
                  >
                    <span>{selectedProject?.name || 'Project name'}</span>
                    <ChevronDown className="w-4 h-4 text-gray-400" />
                  </button>
                  
                  {projectDropdownOpen && (
                    <div className="absolute top-full left-0 mt-1 bg-white border border-gray-200 rounded-md shadow-lg z-50 min-w-[200px]">
                      {projects.map((project) => (
                        <button
                          key={project.id}
                          onClick={() => handleProjectSelect(project.id)}
                          className={`w-full text-left px-3 py-2 text-sm hover:bg-gray-50 transition-colors ${
                            project.id === projectId ? 'bg-indigo-50 text-indigo-700' : 'text-gray-700'
                          }`}
                        >
                          {project.name}
                        </button>
                      ))}
                      <div className="border-t border-gray-200">
                        <button
                          onClick={handleAddProject}
                          className="w-full text-left px-3 py-2 text-sm text-indigo-600 hover:bg-indigo-50 transition-colors"
                        >
                          + Add new project
                        </button>
                      </div>
                    </div>
                  )}
                </div>
                
                <span className="text-gray-300">/</span>
                
                {/* Environment dropdown */}
                <div className="relative" ref={envDropdownRef}>
                  <button
                    onClick={() => setEnvDropdownOpen(!envDropdownOpen)}
                    className="flex items-center gap-1.5 text-sm text-gray-700 hover:text-gray-900 transition-colors"
                  >
                    <span>{currentEnv}</span>
                    <ChevronDown className="w-4 h-4 text-gray-400" />
                  </button>
                  
                  {envDropdownOpen && (
                    <div className="absolute top-full left-0 mt-1 bg-white border border-gray-200 rounded-md shadow-lg z-50 min-w-[150px]">
                      {environments.map((env) => (
                        <button
                          key={env}
                          onClick={() => handleEnvironmentSelect(env)}
                          className={`w-full text-left px-3 py-2 text-sm hover:bg-gray-50 transition-colors ${
                            env === currentEnv ? 'bg-indigo-50 text-indigo-700' : 'text-gray-700'
                          }`}
                        >
                          {env}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              </>
            )}
          </div>

          {/* Right side */}
          <div className="flex items-center gap-6 h-full">
            {variant === 'projects' ? (
              /* Documentation link for projects page */
              <a 
                href="#" 
                className="text-sm text-gray-600 hover:text-gray-900 transition-colors"
              >
                Documentation
              </a>
            ) : (
              /* Navigation tabs for canvas page */
              <nav className="flex items-center h-full">
                <button
                  onClick={() => handleTabClick('architecture')}
                  className={`h-full px-4 text-sm transition-colors relative flex items-center ${
                    currentTab === 'architecture' 
                      ? 'text-gray-900 font-medium' 
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                >
                  Architecture
                  {currentTab === 'architecture' && (
                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-[#4F46E5]" />
                  )}
                </button>
                <button
                  onClick={() => handleTabClick('logs')}
                  className={`h-full px-4 text-sm transition-colors relative flex items-center ${
                    currentTab === 'logs' 
                      ? 'text-gray-900 font-medium' 
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                >
                  Logs
                  {currentTab === 'logs' && (
                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-[#4F46E5]" />
                  )}
                </button>
                <button
                  onClick={() => handleTabClick('settings')}
                  className={`h-full px-4 text-sm transition-colors relative flex items-center ${
                    currentTab === 'settings' 
                      ? 'text-gray-900 font-medium' 
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                >
                  Settings
                  {currentTab === 'settings' && (
                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-[#4F46E5]" />
                  )}
                </button>
              </nav>
            )}

            {/* Separator */}
            <div className="w-px h-5 bg-gray-200" />

            {/* Notification bell */}
            <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-full border border-gray-200 hover:bg-gray-50 transition-colors">
              <Bell className="w-4 h-4 text-gray-500" />
              <span className="text-sm text-gray-600">32</span>
            </button>

            {/* User avatar with dropdown */}
            <div className="relative" ref={avatarMenuRef}>
              <button
                onClick={() => setAvatarMenuOpen(!avatarMenuOpen)}
                className="w-8 h-8 rounded-full bg-indigo-500 flex items-center justify-center overflow-hidden ring-2 ring-indigo-300 ring-offset-2 ring-offset-white hover:ring-indigo-400 transition-all cursor-pointer"
              >
                <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
                </svg>
              </button>

              {/* Avatar Menu Dropdown */}
              {avatarMenuOpen && (
                <div className="absolute top-full right-0 mt-2 w-64 bg-white border border-gray-200 rounded-xl shadow-lg z-50 overflow-hidden">
                  {/* Header section */}
                  <div className="p-4 pb-3">
                    {/* Plan badge */}
                    <span className="inline-block px-2.5 py-0.5 bg-cyan-500 text-white text-xs font-medium rounded">
                      {user.plan}
                    </span>
                    
                    {/* Avatar */}
                    <div className="flex justify-center mt-4">
                      <div className="w-16 h-16 rounded-full bg-indigo-500 flex items-center justify-center">
                        <svg className="w-10 h-10 text-white" fill="currentColor" viewBox="0 0 24 24">
                          <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
                        </svg>
                      </div>
                    </div>
                    
                    {/* User info */}
                    <div className="text-center mt-3">
                      <p className="text-sm font-medium text-gray-900">{user.email}</p>
                      <p className="text-xs text-gray-500">{user.role}</p>
                    </div>
                  </div>

                  {/* Menu items */}
                  <div className="border-t border-gray-100">
                    <button className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors">
                      <User className="w-4 h-4 text-gray-400" />
                      <span>Account settings</span>
                    </button>
                    <button className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors">
                      <Settings className="w-4 h-4 text-gray-400" />
                      <span>Org settings</span>
                    </button>
                    <button className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors">
                      <BarChart3 className="w-4 h-4 text-gray-400" />
                      <span>Account usage</span>
                    </button>
                  </div>

                  <div className="border-t border-gray-100">
                    <button className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors">
                      <LayoutGrid className="w-4 h-4 text-gray-400" />
                      <span>Documentation</span>
                    </button>
                    <button className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors">
                      <HelpCircle className="w-4 h-4 text-gray-400" />
                      <span>Support</span>
                    </button>
                  </div>

                  <div className="border-t border-gray-100">
                    <button 
                      onClick={handleDarkThemeToggle}
                      className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                    >
                      <Moon className="w-4 h-4 text-gray-400" />
                      <span>Dark theme</span>
                    </button>
                  </div>

                  <div className="border-t border-gray-100">
                    <button 
                      onClick={handleLogout}
                      className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 transition-colors"
                    >
                      <LogOut className="w-4 h-4" />
                      <span>Logout</span>
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </header>
  )
}
