'use client'

import { useEffect, useState } from 'react'
import { X, Check, Copy, Eye, EyeOff, Database as DatabaseIcon, Settings, Cpu, Globe, RefreshCw, Loader2, ChevronDown, ChevronUp, AlertCircle, ExternalLink, Play, Square } from 'lucide-react'
import { Database, databasesApi } from '@/lib/api/databases'

interface DatabaseDrawerProps {
  database: Database | null
  isOpen: boolean
  onClose: () => void
}

type TabType = 'deployment' | 'database' | 'settings'

export default function DatabaseDrawer({ database, isOpen, onClose }: DatabaseDrawerProps) {
  const [activeTab, setActiveTab] = useState<TabType>('deployment')

  // Reset tab when drawer opens
  useEffect(() => {
    if (isOpen) {
      setActiveTab('deployment')
    }
  }, [isOpen])

  if (!database) return null

  const getEngineIcon = () => {
    switch (database.engine) {
      case 'postgresql': return '🐘'
      case 'mongodb': return '🍃'
      case 'redis': return '⚡'
      case 'mysql': return '🐬'
      default: return '🗄️'
    }
  }

  const tabs: { id: TabType; label: string }[] = [
    { id: 'deployment', label: 'Deployment' },
    { id: 'database', label: 'Database' },
    { id: 'settings', label: 'Settings' },
  ]

  return (
    <div
      className={`fixed right-0 w-[700px] bg-white border-l border-t border-gray-200 rounded-tl-xl z-50 transform transition-transform duration-300 ease-out ${
        isOpen ? 'translate-x-0' : 'translate-x-full'
      }`}
      style={{ top: '96px', height: 'calc(100vh - 96px)' }}
    >
      {/* Header */}
      <div className="px-6 py-5 border-b border-gray-100">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-indigo-100 flex items-center justify-center text-xl">
            {getEngineIcon()}
          </div>
          <div>
            <h2 className="text-lg font-semibold text-gray-900 capitalize">{database.engine}</h2>
            <div className="flex items-center gap-2">
              <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                database.status === 'active' ? 'bg-emerald-100 text-emerald-700' :
                database.status === 'pending' ? 'bg-amber-100 text-amber-700' :
                database.status === 'failed' ? 'bg-red-100 text-red-700' :
                'bg-gray-100 text-gray-700'
              }`}>
                {database.status}
              </span>
              <span className="text-xs text-gray-400">{database.size}</span>
            </div>
          </div>
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
        {activeTab === 'deployment' && <DeploymentTab database={database} />}
        {activeTab === 'database' && <DatabaseTab database={database} />}
        {activeTab === 'settings' && <SettingsTab database={database} />}
      </div>
    </div>
  )
}

// Deployment Tab
function DeploymentTab({ database }: { database: Database }) {
  const [expanded, setExpanded] = useState(true)

  return (
    <div className="p-6 space-y-4">
      {/* Internal hostname */}
      <div className="flex items-center gap-2">
        <div className={`w-4 h-4 rounded-full ${database.status === 'active' ? 'bg-emerald-100' : 'bg-gray-100'} flex items-center justify-center`}>
          <div className={`w-2 h-2 rounded-full ${database.status === 'active' ? 'bg-emerald-500' : 'bg-gray-400'}`} />
        </div>
        <span className="text-sm text-gray-500 font-mono">{database.internal_hostname || 'Provisioning...'}</span>
        {database.internal_hostname && (
          <button 
            onClick={() => navigator.clipboard.writeText(database.internal_hostname!)}
            className="p-1 hover:bg-gray-100 rounded transition-colors"
          >
            <Copy className="w-3.5 h-3.5 text-gray-400" />
          </button>
        )}
      </div>

      {/* Status Card */}
      <div className={`border rounded-xl overflow-hidden ${
        database.status === 'active' ? 'border-emerald-200 bg-emerald-50' :
        database.status === 'pending' ? 'border-amber-200 bg-amber-50' :
        database.status === 'failed' ? 'border-red-200 bg-red-50' :
        'border-gray-200 bg-gray-50'
      }`}>
        <button
          onClick={() => setExpanded(!expanded)}
          className="w-full p-4 flex items-center gap-3 text-left"
        >
          <div className={`w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 ${
            database.status === 'active' ? 'bg-emerald-500' :
            database.status === 'pending' ? 'bg-amber-500' :
            database.status === 'failed' ? 'bg-red-500' :
            'bg-gray-400'
          }`}>
            {database.status === 'active' ? (
              <Check className="w-3 h-3 text-white" />
            ) : database.status === 'pending' ? (
              <Loader2 className="w-3 h-3 text-white animate-spin" />
            ) : database.status === 'failed' ? (
              <X className="w-3 h-3 text-white" />
            ) : (
              <DatabaseIcon className="w-3 h-3 text-white" />
            )}
          </div>
          <span className={`text-sm font-medium flex-1 ${
            database.status === 'active' ? 'text-emerald-700' :
            database.status === 'pending' ? 'text-amber-700' :
            database.status === 'failed' ? 'text-red-700' :
            'text-gray-700'
          }`}>
            {database.status === 'active' ? 'Database running' :
             database.status === 'pending' ? 'Provisioning database...' :
             database.status === 'failed' ? 'Provisioning failed' :
             'Database status unknown'}
          </span>
          {expanded ? (
            <ChevronUp className="w-4 h-4 text-gray-400" />
          ) : (
            <ChevronDown className="w-4 h-4 text-gray-400" />
          )}
        </button>
      </div>

      {/* Quick Actions */}
      <div className="flex items-center gap-2">
        <button className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 border border-gray-200 rounded-xl hover:bg-gray-50 transition-colors">
          <RefreshCw className="w-4 h-4" />
          Restart
        </button>
        <button className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 border border-gray-200 rounded-xl hover:bg-gray-50 transition-colors">
          <Square className="w-4 h-4" />
          Stop
        </button>
      </div>
    </div>
  )
}

// Database Tab (Tables, Credentials)
function DatabaseTab({ database }: { database: Database }) {
  const [showPassword, setShowPassword] = useState(false)
  const [showConnectionUrl, setShowConnectionUrl] = useState(false)

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  return (
    <div className="p-6 space-y-6">
      {/* Credentials Section */}
      <div>
        <div className="flex items-center gap-2.5 mb-5">
          <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center">
            <DatabaseIcon className="w-3.5 h-3.5 text-gray-500" />
          </div>
          <h3 className="font-medium text-gray-900">Connection Credentials</h3>
        </div>

        <div className="space-y-4">
          {/* Host */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Host</label>
            <div className="flex items-center gap-2 p-3 bg-gray-50 rounded-xl">
              <code className="text-sm text-gray-900 flex-1 font-mono">{database.internal_hostname || 'N/A'}</code>
              {database.internal_hostname && (
                <button 
                  onClick={() => copyToClipboard(database.internal_hostname!)}
                  className="p-1.5 hover:bg-gray-200 rounded-lg transition-colors"
                >
                  <Copy className="w-4 h-4 text-gray-400" />
                </button>
              )}
            </div>
          </div>

          {/* Port */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Port</label>
            <div className="flex items-center gap-2 p-3 bg-gray-50 rounded-xl">
              <code className="text-sm text-gray-900 flex-1 font-mono">{database.port || 'N/A'}</code>
              {database.port && (
                <button 
                  onClick={() => copyToClipboard(String(database.port))}
                  className="p-1.5 hover:bg-gray-200 rounded-lg transition-colors"
                >
                  <Copy className="w-4 h-4 text-gray-400" />
                </button>
              )}
            </div>
          </div>

          {/* Username */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Username</label>
            <div className="flex items-center gap-2 p-3 bg-gray-50 rounded-xl">
              <code className="text-sm text-gray-900 flex-1 font-mono">{database.username || 'admin'}</code>
              <button 
                onClick={() => copyToClipboard(database.username || 'admin')}
                className="p-1.5 hover:bg-gray-200 rounded-lg transition-colors"
              >
                <Copy className="w-4 h-4 text-gray-400" />
              </button>
            </div>
          </div>

          {/* Password */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Password</label>
            <div className="flex items-center gap-2 p-3 bg-gray-50 rounded-xl">
              <code className="text-sm text-gray-900 flex-1 font-mono">
                {showPassword ? (database.password || 'N/A') : '••••••••••••'}
              </code>
              <button 
                onClick={() => setShowPassword(!showPassword)}
                className="p-1.5 hover:bg-gray-200 rounded-lg transition-colors"
              >
                {showPassword ? <EyeOff className="w-4 h-4 text-gray-400" /> : <Eye className="w-4 h-4 text-gray-400" />}
              </button>
              {database.password && (
                <button 
                  onClick={() => copyToClipboard(database.password!)}
                  className="p-1.5 hover:bg-gray-200 rounded-lg transition-colors"
                >
                  <Copy className="w-4 h-4 text-gray-400" />
                </button>
              )}
            </div>
          </div>

          {/* Connection URL */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Connection URL</label>
            <div className="flex items-center gap-2 p-3 bg-gray-50 rounded-xl">
              <code className="text-sm text-gray-900 flex-1 font-mono break-all">
                {showConnectionUrl 
                  ? (database.connection_url || 'N/A')
                  : `${database.engine}://****:****@${database.internal_hostname || '****'}:${database.port || '****'}/****`
                }
              </code>
              <button 
                onClick={() => setShowConnectionUrl(!showConnectionUrl)}
                className="p-1.5 hover:bg-gray-200 rounded-lg transition-colors flex-shrink-0"
              >
                {showConnectionUrl ? <EyeOff className="w-4 h-4 text-gray-400" /> : <Eye className="w-4 h-4 text-gray-400" />}
              </button>
              {database.connection_url && (
                <button 
                  onClick={() => copyToClipboard(database.connection_url!)}
                  className="p-1.5 hover:bg-gray-200 rounded-lg transition-colors flex-shrink-0"
                >
                  <Copy className="w-4 h-4 text-gray-400" />
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Tables/Collections Section (placeholder) */}
      <div>
        <div className="flex items-center gap-2.5 mb-5">
          <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center">
            <Settings className="w-3.5 h-3.5 text-gray-500" />
          </div>
          <h3 className="font-medium text-gray-900">
            {database.engine === 'mongodb' ? 'Collections' : 'Tables'}
          </h3>
        </div>

        <div className="p-6 border border-dashed border-gray-200 rounded-xl text-center">
          <DatabaseIcon className="w-8 h-8 text-gray-300 mx-auto mb-2" />
          <p className="text-sm text-gray-400">
            {database.engine === 'mongodb' ? 'Collection' : 'Table'} browser coming soon
          </p>
        </div>
      </div>
    </div>
  )
}

// Settings Tab
function SettingsTab({ database }: { database: Database }) {
  const [publicUrl, setPublicUrl] = useState(false)

  return (
    <div className="p-6 space-y-6">
      {/* Database Source */}
      <div>
        <div className="flex items-center gap-2.5 mb-5">
          <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center">
            <DatabaseIcon className="w-3.5 h-3.5 text-gray-500" />
          </div>
          <h3 className="font-medium text-gray-900">Database Source</h3>
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Engine</label>
            <div className="p-3 bg-gray-50 rounded-xl text-sm text-gray-900 capitalize">{database.engine}</div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Version</label>
            <div className="p-3 bg-gray-50 rounded-xl text-sm text-gray-900">{database.version || 'Latest'}</div>
          </div>
        </div>
      </div>

      {/* Public URL Toggle */}
      <div>
        <div className="flex items-center gap-2.5 mb-5">
          <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center">
            <Globe className="w-3.5 h-3.5 text-gray-500" />
          </div>
          <h3 className="font-medium text-gray-900">Network</h3>
        </div>

        <div className="flex items-center justify-between p-4 border border-gray-200 rounded-xl">
          <div>
            <div className="text-sm font-medium text-gray-900">Public Database URL</div>
            <div className="text-xs text-gray-400 mt-0.5">Allow connections from outside the project</div>
          </div>
          <button 
            onClick={() => setPublicUrl(!publicUrl)}
            className={`relative w-11 h-6 rounded-full transition-colors ${publicUrl ? 'bg-indigo-600' : 'bg-gray-200'}`}
          >
            <span className={`absolute top-1 w-4 h-4 bg-white rounded-full shadow transition-transform ${publicUrl ? 'right-1' : 'left-1'}`} />
          </button>
        </div>

        {publicUrl && (
          <div className="mt-3 p-3 bg-amber-50 border border-amber-200 rounded-xl">
            <div className="flex items-start gap-2">
              <AlertCircle className="w-4 h-4 text-amber-500 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-amber-700">
                <strong>Warning:</strong> Enabling public access exposes your database to the internet. Make sure to use strong credentials.
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Resource Limits */}
      <div>
        <div className="flex items-center gap-2.5 mb-5">
          <div className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center">
            <Cpu className="w-3.5 h-3.5 text-gray-500" />
          </div>
          <h3 className="font-medium text-gray-900">Resource Limits</h3>
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Instance Size</label>
            <div className="p-3 bg-gray-50 rounded-xl text-sm text-gray-900 capitalize">{database.size}</div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Storage</label>
            <div className="p-3 bg-gray-50 rounded-xl text-sm text-gray-900">{database.volume_size_mb} MB</div>
          </div>
        </div>
      </div>

      {/* Danger Zone */}
      <div className="p-5 border border-red-200 bg-red-50 rounded-xl">
        <h3 className="font-medium text-red-900 mb-2">Delete database</h3>
        <p className="text-sm text-red-700 mb-4">
          This will permanently delete the database and all its data. This action cannot be undone.
        </p>
        <button className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors">
          Delete database
        </button>
      </div>
    </div>
  )
}
