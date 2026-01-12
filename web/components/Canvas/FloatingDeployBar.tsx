'use client'

import { useState } from 'react'
import { Rocket, X, ChevronUp, ChevronDown, Loader2 } from 'lucide-react'
import { useChangesStore, type ChangeType } from '@/stores/changesStore'
import { servicesApi } from '@/lib/api/services'

interface FloatingDeployBarProps {
  serviceId: string
  serviceName: string
  onDeployStart?: () => void
  onDeployComplete?: () => void
}

const changeTypeLabels: Record<ChangeType, string> = {
  root_dir: 'Root Directory',
  branch: 'Branch',
  env_var_add: 'New Variable',
  env_var_update: 'Updated Variable',
  env_var_delete: 'Removed Variable',
  resource_cpu: 'CPU',
  resource_memory: 'Memory',
  custom_domain_add: 'New Domain',
  custom_domain_remove: 'Removed Domain',
  port: 'Port',
  start_command: 'Start Command',
}

export default function FloatingDeployBar({
  serviceId,
  serviceName,
  onDeployStart,
  onDeployComplete,
}: FloatingDeployBarProps) {
  const [expanded, setExpanded] = useState(false)
  const [deploying, setDeploying] = useState(false)
  
  const { 
    getChangesCount, 
    getChanges, 
    getChangeSummary, 
    clearChanges,
    discardChanges 
  } = useChangesStore()
  
  const changesCount = getChangesCount(serviceId)
  const changes = getChanges(serviceId)
  const summary = getChangeSummary(serviceId)

  if (changesCount === 0) {
    return null
  }

  const handleDeploy = async () => {
    setDeploying(true)
    onDeployStart?.()
    
    try {
      // Trigger deployment
      await servicesApi.triggerDeployment(serviceId)
      
      // Clear changes after successful deployment trigger
      clearChanges(serviceId)
      
      onDeployComplete?.()
    } catch (error) {
      console.error('Failed to trigger deployment:', error)
    } finally {
      setDeploying(false)
    }
  }

  const handleDiscard = () => {
    discardChanges(serviceId)
  }

  return (
    <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-slide-up">
      <div className="bg-white rounded-xl shadow-lg border border-gray-200 overflow-hidden min-w-[400px]">
        {/* Main bar */}
        <div className="flex items-center justify-between px-4 py-3">
          <div className="flex items-center gap-3">
            {/* Changes indicator */}
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-amber-500 animate-pulse" />
              <span className="text-sm font-medium text-gray-700">
                {changesCount} {changesCount === 1 ? 'change' : 'changes'} pending
              </span>
            </div>
            
            {/* Expand button */}
            <button
              onClick={() => setExpanded(!expanded)}
              className="p-1 rounded hover:bg-gray-100 transition-colors"
            >
              {expanded ? (
                <ChevronDown className="w-4 h-4 text-gray-500" />
              ) : (
                <ChevronUp className="w-4 h-4 text-gray-500" />
              )}
            </button>
          </div>

          <div className="flex items-center gap-2">
            {/* Discard button */}
            <button
              onClick={handleDiscard}
              disabled={deploying}
              className="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-lg transition-colors disabled:opacity-50"
            >
              Discard
            </button>
            
            {/* Deploy button */}
            <button
              onClick={handleDeploy}
              disabled={deploying}
              className="flex items-center gap-2 px-4 py-1.5 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
            >
              {deploying ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span>Deploying...</span>
                </>
              ) : (
                <>
                  <Rocket className="w-4 h-4" />
                  <span>Deploy now</span>
                </>
              )}
            </button>
          </div>
        </div>

        {/* Expanded details */}
        {expanded && (
          <div className="border-t border-gray-100 px-4 py-3 bg-gray-50">
            <div className="text-xs text-gray-500 uppercase tracking-wider mb-2">
              Changes to deploy
            </div>
            
            {/* Summary badges */}
            <div className="flex flex-wrap gap-2 mb-3">
              {summary.map(({ type, count }) => (
                <span
                  key={type}
                  className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-indigo-100 text-indigo-700"
                >
                  {changeTypeLabels[type]} {count > 1 && `(${count})`}
                </span>
              ))}
            </div>
            
            {/* Detailed list */}
            <div className="space-y-1.5 max-h-32 overflow-y-auto">
              {changes.map((change) => (
                <div
                  key={change.id}
                  className="flex items-center justify-between text-sm"
                >
                  <span className="text-gray-600">{change.description}</span>
                  <button
                    onClick={() => useChangesStore.getState().removeChange(serviceId, change.id)}
                    className="p-0.5 rounded hover:bg-gray-200 transition-colors"
                  >
                    <X className="w-3 h-3 text-gray-400" />
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

