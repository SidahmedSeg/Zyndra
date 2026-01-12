'use client'

import { useEffect, useState, useCallback } from 'react'
import { X, Check, MoreVertical, ChevronDown, ChevronUp, Link2, Trash2, ExternalLink, Copy, Globe, GitBranch, Folder, Cpu, HardDrive, AlertCircle, Loader2, Eye, EyeOff, Plus, RefreshCw } from 'lucide-react'
import { Service, servicesApi } from '@/lib/api/services'
import { useDeploymentsStore } from '@/stores/deploymentsStore'
import { useChangesStore, type ServiceConfig } from '@/stores/changesStore'
import { envVarsApi, type EnvVar } from '@/lib/api/env-vars'
import { customDomainsApi, type CustomDomain } from '@/lib/api/custom-domains'
import type { Deployment } from '@/lib/api/deployments'

interface ServiceDrawerProps {
  service: Service | null
  isOpen: boolean
  onClose: () => void
  initialTab?: TabType
}

type TabType = 'deployment' | 'variables' | 'metrics' | 'settings'
type SettingsSubTab = 'source' | 'network' | 'resources' | 'danger'

export default function ServiceDrawer({ service, isOpen, onClose, initialTab = 'deployment' }: ServiceDrawerProps) {
  const [activeTab, setActiveTab] = useState<TabType>(initialTab)
  const [settingsSubTab, setSettingsSubTab] = useState<SettingsSubTab>('source')
  const { deployments, fetchDeployments } = useDeploymentsStore()
  const { initializeService, serviceChanges, hasChanges } = useChangesStore()
  
  const serviceId = service?.id || null
  const serviceDeployments = serviceId ? deployments[serviceId] || [] : []
  const latestDeployment = serviceDeployments[0] || null

  // Initialize change tracking when drawer opens
  useEffect(() => {
    if (isOpen && service) {
      const config: ServiceConfig = {
        rootDir: service.root_dir || '/',
        branch: service.branch || 'main',
        port: service.port || 8080,
        cpu: service.cpu_limit || '0.5',
        memory: service.memory_limit || '512Mi',
        startCommand: service.start_command || '',
        envVars: {},
        customDomains: [],
      }
      initializeService(service.id, service.name, config)
    }
  }, [isOpen, service, initializeService])

  useEffect(() => {
    if (isOpen && serviceId) {
      fetchDeployments(serviceId)
    }
  }, [isOpen, serviceId, fetchDeployments])

  // Set initial tab when drawer opens
  useEffect(() => {
    if (isOpen) {
      setActiveTab(initialTab)
    }
  }, [isOpen, initialTab])

  if (!service) return null

  const tabs: { id: TabType; label: string }[] = [
    { id: 'deployment', label: 'Deployment' },
    { id: 'variables', label: 'Variables' },
    { id: 'metrics', label: 'Metrics' },
    { id: 'settings', label: 'Settings' },
  ]

  return (
    <>
      {/* Drawer - no backdrop, no shadow, top-left rounded corner */}
      <div
        className={`fixed right-0 w-[700px] bg-white border-l border-t border-gray-200 rounded-tl-xl z-50 transform transition-transform duration-300 ease-out ${
          isOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
        style={{ top: '96px', height: 'calc(100vh - 96px)' }}
      >
        {/* Header */}
        <div className="px-6 py-5 border-b border-gray-100">
          <div className="flex items-center gap-3">
            <img src="/github-icon.svg" alt="" className="w-7 h-7" />
            <h2 className="text-lg font-semibold text-gray-900">{service.name}</h2>
            {hasChanges(service.id) && (
              <span className="px-2 py-0.5 text-xs font-medium bg-amber-100 text-amber-700 rounded-full">
                Unsaved changes
              </span>
            )}
            <button
              onClick={onClose}
              className="ml-auto p-1.5 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <X className="w-5 h-5 text-gray-400" />
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="px-6 border-b border-gray-100">
          <nav className="flex gap-6">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-3 text-sm font-medium transition-colors relative ${
                  activeTab === tab.id
                    ? 'text-gray-900'
                    : 'text-gray-400 hover:text-gray-600'
                }`}
              >
                {tab.label}
                {activeTab === tab.id && (
                  <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-gray-900" />
                )}
              </button>
            ))}
          </nav>
        </div>

        {/* Content */}
        <div className="overflow-y-auto h-[calc(100%-110px)]">
          {activeTab === 'deployment' && (
            <DeploymentTab service={service} deployment={latestDeployment} deployments={serviceDeployments} />
          )}
          {activeTab === 'variables' && (
            <VariablesTab service={service} />
          )}
          {activeTab === 'metrics' && (
            <MetricsTab service={service} />
          )}
          {activeTab === 'settings' && (
            <SettingsTab 
              service={service} 
              subTab={settingsSubTab} 
              onSubTabChange={setSettingsSubTab} 
            />
          )}
        </div>
      </div>
    </>
  )
}

// Deployment Tab
function DeploymentTab({ service, deployment, deployments }: { service: Service; deployment: Deployment | null; deployments: Deployment[] }) {
  const [expanded, setExpanded] = useState(true)
  const [showHistory, setShowHistory] = useState(false)
  const displayUrl = service.generated_url || `${service.name.toLowerCase()}-production.up.zyndra.app`

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success': return 'bg-emerald-500'
      case 'failed': return 'bg-red-500'
      case 'building': return 'bg-blue-500 animate-pulse'
      case 'deploying': return 'bg-amber-500 animate-pulse'
      default: return 'bg-gray-400'
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'success': return { bg: 'bg-emerald-100', text: 'text-emerald-700', label: 'Active' }
      case 'failed': return { bg: 'bg-red-100', text: 'text-red-700', label: 'Failed' }
      case 'building': return { bg: 'bg-blue-100', text: 'text-blue-700', label: 'Building' }
      case 'deploying': return { bg: 'bg-amber-100', text: 'text-amber-700', label: 'Deploying' }
      default: return { bg: 'bg-gray-100', text: 'text-gray-700', label: status }
    }
  }

  return (
    <div className="p-6 space-y-4">
      {/* URL with green indicator */}
      <div className="flex items-center gap-2">
        <div className={`w-4 h-4 rounded-full ${deployment?.status === 'success' ? 'bg-emerald-100' : 'bg-gray-100'} flex items-center justify-center`}>
          <div className={`w-2 h-2 rounded-full ${deployment?.status === 'success' ? 'bg-emerald-500' : 'bg-gray-400'}`} />
        </div>
        <a href={`https://${displayUrl}`} target="_blank" rel="noopener noreferrer" className="text-sm text-gray-500 hover:text-indigo-600 transition-colors">
          {displayUrl}
        </a>
        <button className="p-1 hover:bg-gray-100 rounded transition-colors">
          <Copy className="w-3.5 h-3.5 text-gray-400" />
        </button>
      </div>

      {/* Latest Deployment Card */}
      {deployment ? (
        <div className="border border-gray-200 rounded-xl overflow-hidden">
          <div className="p-4 bg-white">
            <div className="flex items-start gap-3">
              <img src="/github-icon.svg" alt="" className="w-5 h-5 mt-0.5 opacity-60" />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-sm font-medium text-gray-900">
                    {deployment.commit_message || 'Manual deployment'}
                  </span>
                  <span className={`px-2 py-0.5 text-xs font-medium ${getStatusBadge(deployment.status).bg} ${getStatusBadge(deployment.status).text} rounded-full`}>
                    {getStatusBadge(deployment.status).label}
                  </span>
                </div>
                <div className="text-xs text-gray-400 mt-1">
                  {deployment.created_at ? formatTimeAgo(deployment.created_at) : 'Just now'} via {deployment.triggered_by || 'manual'}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button className="px-3 py-1.5 text-xs font-medium text-gray-600 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">
                  View logs
                </button>
                <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors">
                  <MoreVertical className="w-4 h-4 text-gray-400" />
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className="border border-gray-200 rounded-xl p-4 text-center text-gray-400 text-sm">
          No deployments yet
        </div>
      )}

      {/* Status banner */}
      {deployment && (
        <div className={`border rounded-xl overflow-hidden ${
          deployment.status === 'success' 
            ? 'border-emerald-200 bg-emerald-50' 
            : deployment.status === 'failed'
            ? 'border-red-200 bg-red-50'
            : 'border-blue-200 bg-blue-50'
        }`}>
          <button
            onClick={() => setExpanded(!expanded)}
            className="w-full p-4 flex items-center gap-3 text-left"
          >
            <div className={`w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 ${
              deployment.status === 'success' ? 'bg-emerald-500' : 
              deployment.status === 'failed' ? 'bg-red-500' : 'bg-blue-500'
            }`}>
              {deployment.status === 'success' ? (
                <Check className="w-3 h-3 text-white" />
              ) : deployment.status === 'failed' ? (
                <X className="w-3 h-3 text-white" />
              ) : (
                <Loader2 className="w-3 h-3 text-white animate-spin" />
              )}
            </div>
            <span className={`text-sm font-medium flex-1 ${
              deployment.status === 'success' ? 'text-emerald-700' :
              deployment.status === 'failed' ? 'text-red-700' : 'text-blue-700'
            }`}>
              {deployment.status === 'success' ? 'Successfully deployed' :
               deployment.status === 'failed' ? 'Deployment failed' : 'Deploying...'}
            </span>
            {expanded ? (
              <ChevronUp className={`w-4 h-4 ${
                deployment.status === 'success' ? 'text-emerald-500' :
                deployment.status === 'failed' ? 'text-red-500' : 'text-blue-500'
              }`} />
            ) : (
              <ChevronDown className={`w-4 h-4 ${
                deployment.status === 'success' ? 'text-emerald-500' :
                deployment.status === 'failed' ? 'text-red-500' : 'text-blue-500'
              }`} />
            )}
          </button>
        </div>
      )}

      {/* Deployment History */}
      {deployments.length > 1 && (
        <div className="border-t border-gray-100 pt-4">
          <button
            onClick={() => setShowHistory(!showHistory)}
            className="flex items-center gap-2 text-sm text-gray-500 hover:text-gray-700"
          >
            <RefreshCw className="w-4 h-4" />
            <span>Deployment history ({deployments.length})</span>
            {showHistory ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
          
          {showHistory && (
            <div className="mt-3 space-y-2">
              {deployments.slice(1, 6).map((dep) => (
                <div key={dep.id} className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  <div className={`w-2 h-2 rounded-full ${getStatusColor(dep.status)}`} />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm text-gray-700 truncate">{dep.commit_message || 'Manual deployment'}</div>
                    <div className="text-xs text-gray-400">{formatTimeAgo(dep.created_at)}</div>
                  </div>
                  <span className={`px-2 py-0.5 text-xs font-medium ${getStatusBadge(dep.status).bg} ${getStatusBadge(dep.status).text} rounded-full`}>
                    {getStatusBadge(dep.status).label}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// Variables Tab with Change Tracking
function VariablesTab({ service }: { service: Service }) {
  const [variables, setVariables] = useState<EnvVar[]>([])
  const [loading, setLoading] = useState(true)
  const [isAdding, setIsAdding] = useState(false)
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')
  const [revealedVars, setRevealedVars] = useState<Set<string>>(new Set())
  const [editingVar, setEditingVar] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  
  const { addChange } = useChangesStore()

  useEffect(() => {
    loadVariables()
  }, [service.id])

  const loadVariables = async () => {
    try {
      setLoading(true)
      const vars = await envVarsApi.listByService(service.id)
      setVariables(vars)
    } catch (error) {
      console.error('Failed to load variables:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleAddVariable = async () => {
    if (newKey && newValue) {
      try {
        await envVarsApi.create(service.id, { key: newKey, value: newValue })
        
        // Track change
        addChange(service.id, {
          type: 'env_var_add',
          field: newKey,
          newValue: newValue,
          description: `Added variable ${newKey}`,
        })
        
        await loadVariables()
        setNewKey('')
        setNewValue('')
        setIsAdding(false)
      } catch (error) {
        console.error('Failed to add variable:', error)
      }
    }
  }

  const handleUpdateVariable = async (key: string) => {
    if (editValue) {
      try {
        await envVarsApi.update(service.id, key, { value: editValue })
        
        // Track change
        addChange(service.id, {
          type: 'env_var_update',
          field: key,
          newValue: editValue,
          description: `Updated variable ${key}`,
        })
        
        await loadVariables()
        setEditingVar(null)
        setEditValue('')
      } catch (error) {
        console.error('Failed to update variable:', error)
      }
    }
  }

  const handleDeleteVariable = async (key: string) => {
    try {
      await envVarsApi.delete(service.id, key)
      
      // Track change
      addChange(service.id, {
        type: 'env_var_delete',
        field: key,
        description: `Deleted variable ${key}`,
      })
      
      await loadVariables()
    } catch (error) {
      console.error('Failed to delete variable:', error)
    }
  }

  const toggleReveal = (key: string) => {
    setRevealedVars(prev => {
      const newSet = new Set(prev)
      if (newSet.has(key)) {
        newSet.delete(key)
      } else {
        newSet.add(key)
      }
      return newSet
    })
  }

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center">
        <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
      </div>
    )
  }

  return (
    <div className="p-6">
      {/* Header row */}
      <div className="flex items-center justify-between mb-4">
        <span className="text-sm text-gray-600">{variables.length} Service variable{variables.length !== 1 ? 's' : ''}</span>
        <div className="flex items-center gap-2">
          <button className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-500 hover:bg-gray-100 rounded-lg transition-colors">
            <span className="font-mono text-xs text-gray-400">{'{}'}</span>
            <span>Editor</span>
          </button>
          <button 
            onClick={() => setIsAdding(true)}
            className="flex items-center gap-1 px-3 py-1.5 text-sm font-medium text-indigo-600 border border-indigo-200 rounded-lg hover:bg-indigo-50 transition-colors"
          >
            <Plus className="w-4 h-4" />
            New variable
          </button>
        </div>
      </div>

      {/* Add new variable form */}
      {isAdding && (
        <div className="flex items-center gap-2 mb-4 p-3 bg-gray-50 rounded-xl">
          <input
            type="text"
            placeholder="VARIABLE_NAME"
            value={newKey}
            onChange={(e) => setNewKey(e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, ''))}
            className="flex-1 px-3 py-2 text-sm font-mono border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
          <input
            type="text"
            placeholder="Value"
            value={newValue}
            onChange={(e) => setNewValue(e.target.value)}
            className="flex-1 px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
          <button
            onClick={handleAddVariable}
            disabled={!newKey || !newValue}
            className="p-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
          >
            <Check className="w-4 h-4" />
          </button>
          <button
            onClick={() => { setIsAdding(false); setNewKey(''); setNewValue('') }}
            className="p-2 bg-gray-100 text-gray-500 rounded-lg hover:bg-gray-200 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Variables list */}
      <div className="space-y-0 border border-gray-200 rounded-xl overflow-hidden">
        {variables.length === 0 ? (
          <div className="p-6 text-center text-gray-400 text-sm">
            No environment variables defined yet
          </div>
        ) : (
          variables.map((variable, index) => (
            <div 
              key={variable.key} 
              className={`flex items-center gap-3 p-3 bg-white ${index !== variables.length - 1 ? 'border-b border-gray-100' : ''}`}
            >
              <span className="font-mono text-xs text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">{'{}'}</span>
              <span className="text-sm font-medium text-gray-900 font-mono min-w-[150px]">{variable.key}</span>
              
              {editingVar === variable.key ? (
                <div className="flex-1 flex items-center gap-2">
                  <input
                    type="text"
                    value={editValue}
                    onChange={(e) => setEditValue(e.target.value)}
                    className="flex-1 px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    autoFocus
                  />
                  <button
                    onClick={() => handleUpdateVariable(variable.key)}
                    className="p-1 text-emerald-600 hover:bg-emerald-50 rounded"
                  >
                    <Check className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => { setEditingVar(null); setEditValue('') }}
                    className="p-1 text-gray-400 hover:bg-gray-100 rounded"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              ) : (
                <>
                  <span className="text-sm text-gray-500 flex-1 font-mono truncate">
                    {revealedVars.has(variable.key) ? variable.value : '••••••••'}
                  </span>
                  <button 
                    onClick={() => toggleReveal(variable.key)}
                    className="p-1 hover:bg-gray-100 rounded transition-colors"
                  >
                    {revealedVars.has(variable.key) ? (
                      <EyeOff className="w-4 h-4 text-gray-400" />
                    ) : (
                      <Eye className="w-4 h-4 text-gray-400" />
                    )}
                  </button>
                  <button 
                    onClick={() => { setEditingVar(variable.key); setEditValue(variable.value || '') }}
                    className="p-1 hover:bg-gray-100 rounded transition-colors"
                  >
                    <MoreVertical className="w-4 h-4 text-gray-400" />
                  </button>
                  <button 
                    onClick={() => handleDeleteVariable(variable.key)}
                    className="p-1 hover:bg-red-50 rounded transition-colors"
                  >
                    <Trash2 className="w-4 h-4 text-gray-400 hover:text-red-500" />
                  </button>
                </>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  )
}

// Metrics Tab - Step line chart style
function MetricsTab({ service }: { service: Service }) {
  const dates = ['JAN 4', 'JAN 5', 'JAN 6', 'JAN 7', 'JAN 8', 'JAN 9', 'JAN 10']

  return (
    <div className="p-6 space-y-6">
      {/* CPU Chart */}
      <div className="border border-gray-200 rounded-xl p-5">
        <div className="flex items-center justify-between mb-6">
          <h3 className="font-semibold text-gray-900">CPU</h3>
          <div className="flex items-center gap-1.5 text-xs text-gray-500">
            <span className="w-2 h-2 rounded-full bg-indigo-600" />
            <span>Sum</span>
          </div>
        </div>
        
        <div className="relative h-44">
          {/* Y-axis labels */}
          <div className="absolute left-0 top-0 bottom-6 w-14 flex flex-col justify-between text-xs text-gray-400 text-right pr-2">
            <span>1.2 vCPU</span>
            <span>1.0 vCPU</span>
            <span>0.8 vCPU</span>
            <span>0.6 vCPU</span>
            <span>0.4 vCPU</span>
            <span>0.2 vCPU</span>
            <span>0.0 vCPU</span>
          </div>
          
          {/* Chart area */}
          <div className="ml-16 h-full border-l border-b border-gray-200 relative">
            <svg className="w-full h-[calc(100%-24px)]" viewBox="0 0 400 150" preserveAspectRatio="none">
              <polyline
                fill="none"
                stroke="#4F46E5"
                strokeWidth="2"
                points="0,140 57,135 57,130 114,130 114,100 171,100 171,60 228,60 228,30 285,30 285,40 342,40 342,45 400,45"
              />
            </svg>
            
            {/* X-axis labels */}
            <div className="absolute bottom-0 left-0 right-0 flex justify-between text-xs text-gray-400 pt-2 translate-y-full">
              {dates.map((date, i) => (
                <span key={i}>{date}</span>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Memory Chart */}
      <div className="border border-gray-200 rounded-xl p-5">
        <div className="flex items-center justify-between mb-6">
          <h3 className="font-semibold text-gray-900">MEMORY</h3>
          <div className="flex items-center gap-1.5 text-xs text-gray-500">
            <span className="w-2 h-2 rounded-full bg-indigo-600" />
            <span>Sum</span>
          </div>
        </div>
        
        <div className="relative h-44">
          {/* Y-axis labels */}
          <div className="absolute left-0 top-0 bottom-6 w-14 flex flex-col justify-between text-xs text-gray-400 text-right pr-2">
            <span>1.2 GB</span>
            <span>1 GB</span>
            <span>800 MB</span>
            <span>600 MB</span>
            <span>400 MB</span>
            <span>200 MB</span>
            <span>100 MB</span>
          </div>
          
          {/* Chart area */}
          <div className="ml-16 h-full border-l border-b border-gray-200 relative">
            <svg className="w-full h-[calc(100%-24px)]" viewBox="0 0 400 150" preserveAspectRatio="none">
              <polyline
                fill="none"
                stroke="#4F46E5"
                strokeWidth="2"
                points="0,145 57,140 57,135 114,135 114,90 171,90 171,55 228,55 228,25 285,25 285,30 342,30 342,35 400,35"
              />
            </svg>
            
            {/* X-axis labels */}
            <div className="absolute bottom-0 left-0 right-0 flex justify-between text-xs text-gray-400 pt-2 translate-y-full">
              {dates.map((date, i) => (
                <span key={i}>{date}</span>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Network Chart */}
      <div className="border border-gray-200 rounded-xl p-5">
        <div className="flex items-center justify-between mb-6">
          <h3 className="font-semibold text-gray-900">NETWORK</h3>
          <div className="flex items-center gap-4 text-xs text-gray-500">
            <div className="flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-emerald-500" />
              <span>In</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-blue-500" />
              <span>Out</span>
            </div>
          </div>
        </div>
        
        <div className="relative h-44">
          <div className="absolute left-0 top-0 bottom-6 w-14 flex flex-col justify-between text-xs text-gray-400 text-right pr-2">
            <span>100 MB</span>
            <span>80 MB</span>
            <span>60 MB</span>
            <span>40 MB</span>
            <span>20 MB</span>
            <span>0 MB</span>
          </div>
          
          <div className="ml-16 h-full border-l border-b border-gray-200 relative">
            <svg className="w-full h-[calc(100%-24px)]" viewBox="0 0 400 150" preserveAspectRatio="none">
              <polyline
                fill="none"
                stroke="#10B981"
                strokeWidth="2"
                points="0,140 80,120 160,100 240,80 320,90 400,85"
              />
              <polyline
                fill="none"
                stroke="#3B82F6"
                strokeWidth="2"
                points="0,145 80,130 160,115 240,100 320,110 400,105"
              />
            </svg>
            
            <div className="absolute bottom-0 left-0 right-0 flex justify-between text-xs text-gray-400 pt-2 translate-y-full">
              {dates.map((date, i) => (
                <span key={i}>{date}</span>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

// Settings Tab
function SettingsTab({ 
  service, 
  subTab, 
  onSubTabChange 
}: { 
  service: Service
  subTab: SettingsSubTab
  onSubTabChange: (tab: SettingsSubTab) => void 
}) {
  const subTabs: { id: SettingsSubTab; label: string; danger?: boolean }[] = [
    { id: 'source', label: 'Source' },
    { id: 'network', label: 'Network' },
    { id: 'resources', label: 'Resources' },
    { id: 'danger', label: 'Danger zone', danger: true },
  ]

  return (
    <div>
      {/* Sub tabs */}
      <div className="px-6 border-b border-gray-100">
        <div className="flex gap-4">
          {subTabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => onSubTabChange(tab.id)}
              className={`py-3 text-sm font-medium transition-colors relative ${
                subTab === tab.id
                  ? tab.danger ? 'text-red-600' : 'text-indigo-600'
                  : tab.danger ? 'text-red-400 hover:text-red-500' : 'text-gray-400 hover:text-gray-600'
              }`}
            >
              {tab.label}
              {subTab === tab.id && (
                <div className={`absolute bottom-0 left-0 right-0 h-0.5 ${tab.danger ? 'bg-red-600' : 'bg-indigo-600'}`} />
              )}
            </button>
          ))}
        </div>
      </div>

      <div className="p-6">
        {subTab === 'source' && <SourceSettings service={service} />}
        {subTab === 'network' && <NetworkSettings service={service} />}
        {subTab === 'resources' && <ResourceSettings service={service} />}
        {subTab === 'danger' && <DangerSettings service={service} />}
      </div>
    </div>
  )
}

function SourceSettings({ service }: { service: Service }) {
  const [rootDir, setRootDir] = useState(service.root_dir || '/')
  const [branch, setBranch] = useState(service.branch || 'main')
  const [branches, setBranches] = useState(['main', 'develop', 'staging'])
  const [branchDropdownOpen, setBranchDropdownOpen] = useState(false)
  
  const { addChange, serviceChanges } = useChangesStore()

  const handleRootDirChange = (value: string) => {
    setRootDir(value)
    const original = serviceChanges[service.id]?.originalConfig.rootDir || '/'
    if (value !== original) {
      addChange(service.id, {
        type: 'root_dir',
        field: 'rootDir',
        oldValue: original,
        newValue: value,
        description: `Changed root directory to ${value}`,
      })
    }
  }

  const handleBranchChange = (value: string) => {
    setBranch(value)
    setBranchDropdownOpen(false)
    const original = serviceChanges[service.id]?.originalConfig.branch || 'main'
    if (value !== original) {
      addChange(service.id, {
        type: 'branch',
        field: 'branch',
        oldValue: original,
        newValue: value,
        description: `Changed branch to ${value}`,
      })
    }
  }

  return (
    <div className="space-y-6">
      {/* Source code section */}
      <div>
        <div className="flex items-center gap-2.5 mb-5">
          <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center">
            <span className="text-xs text-gray-500">{'</>'}</span>
          </div>
          <h3 className="font-medium text-gray-900">Source code</h3>
        </div>

        {/* Source repository */}
        <div className="mb-5">
          <label className="block text-sm font-medium text-gray-700 mb-2">Source repository</label>
          <div className="flex items-center gap-3 p-3 border border-gray-200 rounded-xl bg-white">
            <img src="/github-icon.svg" alt="" className="w-5 h-5" />
            <span className="text-sm text-gray-900 flex-1">{service.repo_owner}/{service.repo_name || service.name}</span>
            <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors">
              <Link2 className="w-4 h-4 text-gray-400" />
            </button>
            <button className="px-3 py-1.5 text-sm text-gray-600 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">
              Disconnect
            </button>
          </div>
        </div>

        {/* Root directory */}
        <div className="mb-5">
          <label className="block text-sm font-medium text-gray-700 mb-1">Root directory</label>
          <p className="text-xs text-gray-400 mb-2">
            Select the directory containing your application code. <a href="#" className="text-indigo-600 hover:underline">Docs</a>
          </p>
          <div className="relative">
            <Folder className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              value={rootDir}
              onChange={(e) => handleRootDirChange(e.target.value)}
              className="w-full pl-10 pr-3 py-2.5 text-sm border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
        </div>

        {/* Branch */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Branch synced with production</label>
          <p className="text-xs text-gray-400 mb-2">
            Updates to this branch will trigger automatic deployments.
          </p>
          <div className="relative">
            <button
              onClick={() => setBranchDropdownOpen(!branchDropdownOpen)}
              className="w-full flex items-center gap-3 p-3 border border-gray-200 rounded-xl bg-white hover:bg-gray-50 transition-colors"
            >
              <GitBranch className="w-4 h-4 text-gray-400" />
              <span className="text-sm text-gray-900 flex-1 text-left">{branch}</span>
              <ChevronDown className="w-4 h-4 text-gray-400" />
            </button>
            
            {branchDropdownOpen && (
              <div className="absolute top-full left-0 right-0 mt-1 bg-white border border-gray-200 rounded-xl shadow-lg z-10 overflow-hidden">
                {branches.map((b) => (
                  <button
                    key={b}
                    onClick={() => handleBranchChange(b)}
                    className={`w-full flex items-center gap-3 px-3 py-2.5 text-sm hover:bg-gray-50 transition-colors ${
                      b === branch ? 'bg-indigo-50 text-indigo-700' : 'text-gray-700'
                    }`}
                  >
                    <GitBranch className="w-4 h-4" />
                    {b}
                    {b === branch && <Check className="w-4 h-4 ml-auto" />}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function NetworkSettings({ service }: { service: Service }) {
  const [domains, setDomains] = useState<CustomDomain[]>([])
  const [loading, setLoading] = useState(true)
  const [addingDomain, setAddingDomain] = useState(false)
  const [newDomain, setNewDomain] = useState('')
  const [domainError, setDomainError] = useState('')
  
  const { addChange } = useChangesStore()
  
  const generatedUrl = service.generated_url || `${service.name.toLowerCase()}-production.up.zyndra.app`

  useEffect(() => {
    loadDomains()
  }, [service.id])

  const loadDomains = async () => {
    try {
      setLoading(true)
      const data = await customDomainsApi.listByService(service.id)
      setDomains(data)
    } catch (error) {
      console.error('Failed to load domains:', error)
    } finally {
      setLoading(false)
    }
  }

  const validateDomain = (domain: string): boolean => {
    const domainRegex = /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/
    return domainRegex.test(domain)
  }

  const handleAddDomain = async () => {
    if (!newDomain) return
    
    if (!validateDomain(newDomain)) {
      setDomainError('Please enter a valid domain name')
      return
    }
    
    try {
      await customDomainsApi.create(service.id, { domain: newDomain })
      
      addChange(service.id, {
        type: 'custom_domain_add',
        field: 'customDomains',
        newValue: newDomain,
        description: `Added custom domain ${newDomain}`,
      })
      
      await loadDomains()
      setNewDomain('')
      setAddingDomain(false)
      setDomainError('')
    } catch (error) {
      console.error('Failed to add domain:', error)
      setDomainError('Failed to add domain')
    }
  }

  const handleRemoveDomain = async (domainId: string, domainName: string) => {
    try {
      await customDomainsApi.delete(service.id, domainId)
      
      addChange(service.id, {
        type: 'custom_domain_remove',
        field: 'customDomains',
        oldValue: domainName,
        description: `Removed custom domain ${domainName}`,
      })
      
      await loadDomains()
    } catch (error) {
      console.error('Failed to remove domain:', error)
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return { bg: 'bg-emerald-100', text: 'text-emerald-700', label: 'Verified' }
      case 'pending':
        return { bg: 'bg-amber-100', text: 'text-amber-700', label: 'Pending DNS' }
      case 'verifying':
        return { bg: 'bg-blue-100', text: 'text-blue-700', label: 'Verifying...' }
      case 'failed':
        return { bg: 'bg-red-100', text: 'text-red-700', label: 'Failed' }
      default:
        return { bg: 'bg-gray-100', text: 'text-gray-700', label: status }
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-2.5 mb-5">
          <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center">
            <Globe className="w-3.5 h-3.5 text-gray-500" />
          </div>
          <h3 className="font-medium text-gray-900">Network</h3>
        </div>

        {/* Generated URL */}
        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 mb-2">Generated URL</label>
          <div className="flex items-center gap-3 p-3 border border-gray-200 rounded-xl bg-gray-50">
            <div className="w-6 h-6 rounded-full bg-emerald-100 flex items-center justify-center flex-shrink-0">
              <Globe className="w-3.5 h-3.5 text-emerald-600" />
            </div>
            <span className="text-sm text-gray-900 flex-1 font-mono">{generatedUrl}</span>
            <button 
              onClick={() => navigator.clipboard.writeText(`https://${generatedUrl}`)}
              className="p-1.5 hover:bg-gray-200 rounded-lg transition-colors"
            >
              <Copy className="w-4 h-4 text-gray-400" />
            </button>
            <a 
              href={`https://${generatedUrl}`} 
              target="_blank" 
              rel="noopener noreferrer"
              className="p-1.5 hover:bg-gray-200 rounded-lg transition-colors"
            >
              <ExternalLink className="w-4 h-4 text-gray-400" />
            </a>
          </div>
        </div>

        {/* Custom Domains */}
        <div>
          <h4 className="text-sm font-medium text-gray-900 mb-1">Custom Domains</h4>
          <p className="text-xs text-gray-400 mb-4">Add your own domain and point its CNAME to the generated URL above.</p>

          {loading ? (
            <div className="flex justify-center py-4">
              <Loader2 className="w-5 h-5 animate-spin text-gray-400" />
            </div>
          ) : (
            <div className="space-y-2">
              {domains.map((domain) => (
                <div key={domain.id} className="flex items-center gap-3 p-3 border border-gray-200 rounded-xl bg-white">
                  <div className="w-6 h-6 rounded-full bg-indigo-100 flex items-center justify-center flex-shrink-0">
                    <Globe className="w-3.5 h-3.5 text-indigo-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-sm text-gray-900 truncate font-mono">{domain.domain}</div>
                  </div>
                  <span className={`px-2 py-0.5 text-xs font-medium ${getStatusBadge(domain.status).bg} ${getStatusBadge(domain.status).text} rounded-full flex-shrink-0`}>
                    {getStatusBadge(domain.status).label}
                  </span>
                  <div className="flex items-center gap-1 flex-shrink-0">
                    <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors">
                      <Copy className="w-4 h-4 text-gray-400" />
                    </button>
                    <button 
                      onClick={() => handleRemoveDomain(domain.id, domain.domain)}
                      className="p-1.5 hover:bg-red-50 rounded-lg transition-colors"
                    >
                      <Trash2 className="w-4 h-4 text-gray-400 hover:text-red-500" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Add domain form */}
          {addingDomain ? (
            <div className="mt-4 space-y-2">
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  placeholder="api.example.com"
                  value={newDomain}
                  onChange={(e) => { setNewDomain(e.target.value); setDomainError('') }}
                  className={`flex-1 px-3 py-2.5 text-sm border rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent ${
                    domainError ? 'border-red-300' : 'border-gray-200'
                  }`}
                />
                <button
                  onClick={handleAddDomain}
                  disabled={!newDomain}
                  className="p-2.5 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 transition-colors disabled:opacity-50"
                >
                  <Check className="w-4 h-4" />
                </button>
                <button
                  onClick={() => { setAddingDomain(false); setNewDomain(''); setDomainError('') }}
                  className="p-2.5 bg-gray-100 text-gray-500 rounded-xl hover:bg-gray-200 transition-colors"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
              {domainError && (
                <div className="flex items-center gap-2 text-sm text-red-600">
                  <AlertCircle className="w-4 h-4" />
                  {domainError}
                </div>
              )}
              <div className="p-3 bg-gray-50 rounded-xl text-xs text-gray-500">
                <strong>CNAME Setup:</strong> Point your domain&apos;s CNAME record to:
                <code className="block mt-1 font-mono bg-gray-100 px-2 py-1 rounded">{generatedUrl}</code>
              </div>
            </div>
          ) : (
            <button 
              onClick={() => setAddingDomain(true)}
              className="mt-4 flex items-center gap-1 px-4 py-2 text-sm font-medium text-indigo-600 border border-indigo-200 rounded-xl hover:bg-indigo-50 transition-colors"
            >
              <Plus className="w-4 h-4" />
              Add custom domain
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

function ResourceSettings({ service }: { service: Service }) {
  const [cpu, setCpu] = useState(service.cpu_limit || '0.5')
  const [memory, setMemory] = useState(service.memory_limit || '512Mi')
  
  const { addChange, serviceChanges } = useChangesStore()

  const cpuOptions = ['0.25', '0.5', '1', '2', '4']
  const memoryOptions = ['256Mi', '512Mi', '1Gi', '2Gi', '4Gi', '8Gi']

  const handleCpuChange = (value: string) => {
    setCpu(value)
    const original = serviceChanges[service.id]?.originalConfig.cpu || '0.5'
    if (value !== original) {
      addChange(service.id, {
        type: 'resource_cpu',
        field: 'cpu',
        oldValue: original,
        newValue: value,
        description: `Changed CPU to ${value} vCPU`,
      })
    }
  }

  const handleMemoryChange = (value: string) => {
    setMemory(value)
    const original = serviceChanges[service.id]?.originalConfig.memory || '512Mi'
    if (value !== original) {
      addChange(service.id, {
        type: 'resource_memory',
        field: 'memory',
        oldValue: original,
        newValue: value,
        description: `Changed memory to ${value}`,
      })
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-2.5 mb-5">
          <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center">
            <Cpu className="w-3.5 h-3.5 text-gray-500" />
          </div>
          <h3 className="font-medium text-gray-900">Resource Limits</h3>
        </div>

        {/* CPU */}
        <div className="mb-5">
          <label className="block text-sm font-medium text-gray-700 mb-2">CPU</label>
          <p className="text-xs text-gray-400 mb-3">Allocate vCPU cores for your service.</p>
          <div className="flex flex-wrap gap-2">
            {cpuOptions.map((option) => (
              <button
                key={option}
                onClick={() => handleCpuChange(option)}
                className={`px-4 py-2 text-sm font-medium rounded-lg border transition-colors ${
                  cpu === option
                    ? 'bg-indigo-50 border-indigo-300 text-indigo-700'
                    : 'bg-white border-gray-200 text-gray-700 hover:bg-gray-50'
                }`}
              >
                {option} vCPU
              </button>
            ))}
          </div>
        </div>

        {/* Memory */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Memory</label>
          <p className="text-xs text-gray-400 mb-3">Allocate RAM for your service.</p>
          <div className="flex flex-wrap gap-2">
            {memoryOptions.map((option) => (
              <button
                key={option}
                onClick={() => handleMemoryChange(option)}
                className={`px-4 py-2 text-sm font-medium rounded-lg border transition-colors ${
                  memory === option
                    ? 'bg-indigo-50 border-indigo-300 text-indigo-700'
                    : 'bg-white border-gray-200 text-gray-700 hover:bg-gray-50'
                }`}
              >
                {option}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Pricing hint */}
      <div className="p-4 bg-amber-50 border border-amber-200 rounded-xl">
        <div className="flex items-start gap-3">
          <AlertCircle className="w-5 h-5 text-amber-500 flex-shrink-0 mt-0.5" />
          <div>
            <h4 className="text-sm font-medium text-amber-800">Resource Usage</h4>
            <p className="text-xs text-amber-700 mt-1">
              Current selection: {cpu} vCPU, {memory} RAM. Resources are allocated per replica.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

function DangerSettings({ service }: { service: Service }) {
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [confirmText, setConfirmText] = useState('')

  return (
    <div className="p-5 border border-red-200 bg-red-50 rounded-xl">
      <h3 className="font-medium text-red-900 mb-2">Delete service</h3>
      <p className="text-sm text-red-700 mb-4">
        This will permanently delete the service and all associated data including deployments, environment variables, and custom domains. This action cannot be undone.
      </p>
      
      {confirmDelete ? (
        <div className="space-y-3">
          <p className="text-sm text-red-700">
            Type <strong>{service.name}</strong> to confirm deletion:
          </p>
          <input
            type="text"
            value={confirmText}
            onChange={(e) => setConfirmText(e.target.value)}
            placeholder={service.name}
            className="w-full px-3 py-2 text-sm border border-red-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500"
          />
          <div className="flex gap-2">
            <button 
              disabled={confirmText !== service.name}
              className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Delete permanently
            </button>
            <button 
              onClick={() => { setConfirmDelete(false); setConfirmText('') }}
              className="px-4 py-2 text-sm font-medium text-gray-600 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <button 
          onClick={() => setConfirmDelete(true)}
          className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
        >
          Delete {service.name}
        </button>
      )}
    </div>
  )
}

// Helper function
function formatTimeAgo(date: string): string {
  const now = new Date()
  const then = new Date(date)
  const diffMs = now.getTime() - then.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins} min${diffMins > 1 ? 's' : ''} ago`
  if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
  return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
}
