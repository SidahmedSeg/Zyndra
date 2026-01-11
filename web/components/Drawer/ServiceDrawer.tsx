'use client'

import { useEffect, useState } from 'react'
import { X, Check, MoreVertical, ChevronDown, ChevronUp, Link2, Trash2, ExternalLink, Copy, Globe } from 'lucide-react'
import { Service } from '@/lib/api/services'
import { useDeploymentsStore } from '@/stores/deploymentsStore'
import type { Deployment } from '@/lib/api/deployments'

interface ServiceDrawerProps {
  service: Service | null
  isOpen: boolean
  onClose: () => void
}

type TabType = 'deployment' | 'variables' | 'metrics' | 'settings'
type SettingsSubTab = 'source' | 'network' | 'build' | 'deploy' | 'danger'

export default function ServiceDrawer({ service, isOpen, onClose }: ServiceDrawerProps) {
  const [activeTab, setActiveTab] = useState<TabType>('deployment')
  const [settingsSubTab, setSettingsSubTab] = useState<SettingsSubTab>('source')
  const { deployments, fetchDeployments } = useDeploymentsStore()
  
  const serviceId = service?.id || null
  const serviceDeployments = serviceId ? deployments[serviceId] || [] : []
  const latestDeployment = serviceDeployments[0] || null

  useEffect(() => {
    if (isOpen && serviceId) {
      fetchDeployments(serviceId)
    }
  }, [isOpen, serviceId, fetchDeployments])

  // Reset to deployment tab when drawer opens
  useEffect(() => {
    if (isOpen) {
      setActiveTab('deployment')
    }
  }, [isOpen])

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
            <DeploymentTab service={service} deployment={latestDeployment} />
          )}
          {activeTab === 'variables' && (
            <VariablesTab service={service} />
          )}
          {activeTab === 'metrics' && (
            <MetricsTab />
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
function DeploymentTab({ service, deployment }: { service: Service; deployment: Deployment | null }) {
  const [expanded, setExpanded] = useState(true)
  const displayUrl = service.generated_url || 'bcd.ila26.com'

  return (
    <div className="p-6 space-y-4">
      {/* URL with green indicator */}
      <div className="flex items-center gap-2">
        <div className="w-4 h-4 rounded-full bg-emerald-100 flex items-center justify-center">
          <div className="w-2 h-2 rounded-full bg-emerald-500" />
        </div>
        <span className="text-sm text-gray-500">{displayUrl}</span>
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
                    {deployment.commit_message || 'fix: Remove unused isWithinInterval import'}
                  </span>
                  <span className="px-2 py-0.5 text-xs font-medium bg-emerald-100 text-emerald-700 rounded-full">
                    Active
                  </span>
                </div>
                <div className="text-xs text-gray-400 mt-1">
                  02 hours ago via Github
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

      {/* Successfully deployed banner */}
      <div className="border border-emerald-200 bg-emerald-50 rounded-xl overflow-hidden">
        <button
          onClick={() => setExpanded(!expanded)}
          className="w-full p-4 flex items-center gap-3 text-left"
        >
          <div className="w-5 h-5 rounded-full bg-emerald-500 flex items-center justify-center flex-shrink-0">
            <Check className="w-3 h-3 text-white" />
          </div>
          <span className="text-sm font-medium text-emerald-700 flex-1">Successfully deployed</span>
          {expanded ? (
            <ChevronUp className="w-4 h-4 text-emerald-500" />
          ) : (
            <ChevronDown className="w-4 h-4 text-emerald-500" />
          )}
        </button>
      </div>
    </div>
  )
}

// Variables Tab
function VariablesTab({ service }: { service: Service }) {
  const [variables, setVariables] = useState([
    { key: 'VITE_API_URL', value: '******', isSecret: true }
  ])
  const [isAdding, setIsAdding] = useState(false)
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')

  const handleAddVariable = () => {
    if (newKey && newValue) {
      setVariables([...variables, { key: newKey, value: newValue, isSecret: false }])
      setNewKey('')
      setNewValue('')
      setIsAdding(false)
    }
  }

  return (
    <div className="p-6">
      {/* Header row */}
      <div className="flex items-center justify-between mb-4">
        <span className="text-sm text-gray-600">{variables.length} Service variable</span>
        <div className="flex items-center gap-2">
          <button className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-500 hover:bg-gray-100 rounded-lg transition-colors">
            <span className="font-mono text-xs text-gray-400">{'{}'}</span>
            <span>Editor</span>
          </button>
          <button 
            onClick={() => setIsAdding(true)}
            className="flex items-center gap-1 px-3 py-1.5 text-sm font-medium text-indigo-600 border border-indigo-200 rounded-lg hover:bg-indigo-50 transition-colors"
          >
            + New variable
          </button>
        </div>
      </div>

      {/* Add new variable form */}
      {isAdding && (
        <div className="flex items-center gap-2 mb-4">
          <input
            type="text"
            placeholder="VARIABLE_NAME"
            value={newKey}
            onChange={(e) => setNewKey(e.target.value)}
            className="flex-1 px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
          <input
            type="text"
            placeholder="VALUE or ${{REF}}"
            value={newValue}
            onChange={(e) => setNewValue(e.target.value)}
            className="flex-1 px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
          <button
            onClick={handleAddVariable}
            className="p-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
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
      <div className="space-y-0">
        {variables.map((variable, index) => (
          <div key={index} className="flex items-center gap-3 py-3 border-b border-gray-100 last:border-0">
            <span className="font-mono text-sm text-gray-400">{'{}'}</span>
            <span className="text-sm font-medium text-gray-900 flex-1">{variable.key}</span>
            <span className="text-sm text-gray-400">{variable.isSecret ? '*******' : variable.value}</span>
            <button className="p-1 hover:bg-gray-100 rounded transition-colors">
              <MoreVertical className="w-4 h-4 text-gray-400" />
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}

// Metrics Tab - Step line chart style
function MetricsTab() {
  // Generate mock step data for chart
  const cpuData = [0.05, 0.08, 0.06, 0.4, 0.55, 0.7, 0.65]
  const memoryData = [100, 120, 150, 350, 500, 700, 650]
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
    { id: 'build', label: 'Build' },
    { id: 'deploy', label: 'Deploy' },
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
        {subTab === 'build' && <BuildSettings />}
        {subTab === 'deploy' && <DeploySettings />}
        {subTab === 'danger' && <DangerSettings service={service} />}
      </div>
    </div>
  )
}

function SourceSettings({ service }: { service: Service }) {
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
            <span className="text-sm text-gray-900 flex-1">Sidahmedseg/{service.name}</span>
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
            Select the source we&apos;ll use to fetch your code. <a href="#" className="text-indigo-600 hover:underline">Docs</a>
          </p>
          <input
            type="text"
            defaultValue="/Backend"
            className="w-full px-3 py-2.5 text-sm border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
        </div>

        {/* Branch synced with production */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Branch synced with production</label>
          <p className="text-xs text-gray-400 mb-2">
            Updates to this GitHub branch will be automatically deployed to this environment.
          </p>
          <div className="flex items-center gap-3 p-3 border border-gray-200 rounded-xl bg-white">
            <span className="text-gray-400">↳</span>
            <span className="text-sm text-gray-900 flex-1">Main branch</span>
            <ChevronDown className="w-4 h-4 text-gray-400" />
            <button className="px-3 py-1.5 text-sm text-gray-600 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">
              Disconnect
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

function NetworkSettings({ service }: { service: Service }) {
  const domains = [
    { url: `Reponame-production.online.zyndra.app`, port: 8080 },
    { url: 'ila26bcd.com', port: 8080, status: 'Setup complete' },
  ]

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-2.5 mb-5">
          <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center">
            <Globe className="w-3.5 h-3.5 text-gray-500" />
          </div>
          <h3 className="font-medium text-gray-900">Network</h3>
        </div>

        <div>
          <h4 className="text-sm font-medium text-gray-900 mb-1">Public Networking</h4>
          <p className="text-xs text-gray-400 mb-4">Your application is available over HTTP at the following domains:</p>

          <div className="space-y-2">
            {domains.map((domain, index) => (
              <div key={index} className="flex items-center gap-3 p-3 border border-gray-200 rounded-xl bg-white">
                <div className="w-6 h-6 rounded-full bg-indigo-100 flex items-center justify-center flex-shrink-0">
                  <Globe className="w-3.5 h-3.5 text-indigo-600" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="text-sm text-gray-900 truncate">{domain.url}</div>
                  <div className="text-xs text-gray-400">Port : {domain.port}</div>
                </div>
                {domain.status && (
                  <span className="px-2 py-0.5 text-xs bg-emerald-100 text-emerald-700 rounded-full flex-shrink-0">
                    {domain.status}
                  </span>
                )}
                <div className="flex items-center gap-1 flex-shrink-0">
                  <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors">
                    <Copy className="w-4 h-4 text-gray-400" />
                  </button>
                  <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors">
                    <ExternalLink className="w-4 h-4 text-gray-400" />
                  </button>
                  <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors">
                    <Trash2 className="w-4 h-4 text-gray-400" />
                  </button>
                </div>
              </div>
            ))}
          </div>

          <button className="mt-4 flex items-center gap-1 px-4 py-2 text-sm font-medium text-indigo-600 border border-indigo-200 rounded-xl hover:bg-indigo-50 transition-colors">
            + Custom domain
          </button>
        </div>
      </div>
    </div>
  )
}

function BuildSettings() {
  return (
    <div className="space-y-5">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Build command</label>
        <input
          type="text"
          placeholder="npm run build"
          className="w-full px-3 py-2.5 text-sm border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Start command</label>
        <input
          type="text"
          placeholder="npm start"
          className="w-full px-3 py-2.5 text-sm border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>
    </div>
  )
}

function DeploySettings() {
  return (
    <div>
      <div className="flex items-center justify-between p-4 border border-gray-200 rounded-xl">
        <div>
          <div className="text-sm font-medium text-gray-900">Auto deploy</div>
          <div className="text-xs text-gray-400 mt-0.5">Automatically deploy when pushing to the main branch</div>
        </div>
        <button className="relative w-11 h-6 bg-indigo-600 rounded-full transition-colors">
          <span className="absolute right-1 top-1 w-4 h-4 bg-white rounded-full shadow transition-transform" />
        </button>
      </div>
    </div>
  )
}

function DangerSettings({ service }: { service: Service }) {
  return (
    <div className="p-5 border border-red-200 bg-red-50 rounded-xl">
      <h3 className="font-medium text-red-900 mb-2">Delete service</h3>
      <p className="text-sm text-red-700 mb-4">
        This will permanently delete the service and all associated data. This action cannot be undone.
      </p>
      <button className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors">
        Delete {service.name}
      </button>
    </div>
  )
}
