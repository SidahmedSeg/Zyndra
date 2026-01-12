'use client'

import { memo, useEffect, useState } from 'react'
import { Handle, Position, NodeProps } from 'reactflow'
import type { Service } from '@/lib/api/services'
import { useServicesStore } from '@/stores/servicesStore'
import { deploymentsApi, type Deployment, getStatusDisplay } from '@/lib/api/deployments'
import { Loader2, CheckCircle, XCircle, Circle } from 'lucide-react'

interface ServiceNodeData {
  label: string
  service: Service
}

// Deployment stages for visual progress
const deploymentStages = [
  { id: 'queued', label: 'Initializing' },
  { id: 'building', label: 'Building image' },
  { id: 'pushing', label: 'Pushing' },
  { id: 'deploying', label: 'Deploying' },
  { id: 'success', label: 'Online' },
]

function getStageIndex(status: string): number {
  const index = deploymentStages.findIndex(s => s.id === status)
  return index >= 0 ? index : 0
}

function ServiceNode({ data, selected }: NodeProps<ServiceNodeData>) {
  const { setSelectedService } = useServicesStore()
  const { service } = data
  
  const [deployment, setDeployment] = useState<Deployment | null>(null)
  const [elapsedTime, setElapsedTime] = useState<string>('00:00')

  const handleClick = () => {
    setSelectedService(service)
  }

  // Poll for latest deployment status
  useEffect(() => {
    let intervalId: NodeJS.Timeout | null = null
    let timerIntervalId: NodeJS.Timeout | null = null

    const fetchLatestDeployment = async () => {
      try {
        const deployments = await deploymentsApi.listByService(service.id, 1)
        if (deployments && deployments.length > 0) {
          setDeployment(deployments[0])
        }
      } catch (error) {
        console.error('Failed to fetch deployment:', error)
      }
    }

    // Initial fetch
    fetchLatestDeployment()

    // Poll every 3 seconds
    intervalId = setInterval(fetchLatestDeployment, 3000)

    // Elapsed time counter
    timerIntervalId = setInterval(() => {
      if (deployment?.started_at && !['success', 'failed', 'cancelled'].includes(deployment.status)) {
        const started = new Date(deployment.started_at).getTime()
        const now = Date.now()
        const elapsed = Math.floor((now - started) / 1000)
        const minutes = Math.floor(elapsed / 60).toString().padStart(2, '0')
        const seconds = (elapsed % 60).toString().padStart(2, '0')
        setElapsedTime(`${minutes}:${seconds}`)
      }
    }, 1000)

    return () => {
      if (intervalId) clearInterval(intervalId)
      if (timerIntervalId) clearInterval(timerIntervalId)
    }
  }, [service.id, deployment?.started_at, deployment?.status])

  // Determine display status based on deployment
  const isDeploying = deployment && !['success', 'failed', 'cancelled'].includes(deployment.status)
  const isFailed = deployment?.status === 'failed'
  const isOnline = deployment?.status === 'success' || service.status === 'running'
  
  const currentStageIndex = deployment ? getStageIndex(deployment.status) : -1

  // Parse URL for display
  const displayUrl = service.generated_url || `${service.name.toLowerCase()}.up.zyndra.app`

  // Status colors
  const getStatusIndicator = () => {
    if (isFailed) {
      return { dot: 'bg-red-500', text: 'text-red-500', label: 'Failed' }
    }
    if (isDeploying) {
      return { dot: 'bg-blue-500 animate-pulse', text: 'text-blue-500', label: getStatusDisplay(deployment!.status).label }
    }
    if (isOnline) {
      return { dot: 'bg-emerald-500', text: 'text-emerald-500', label: 'Online' }
    }
    return { dot: 'bg-gray-400', text: 'text-gray-400', label: 'Idle' }
  }

  const status = getStatusIndicator()

  return (
    <div
      onClick={handleClick}
      className="rounded-2xl min-w-[280px] max-w-[320px] cursor-pointer bg-white shadow-md overflow-hidden border-2 transition-all duration-200"
      style={{ borderColor: selected ? '#4F46E5' : '#e5e7eb' }}
    >
      <Handle type="target" position={Position.Top} className="!bg-gray-400 !w-2 !h-2" />
      
      {/* Content */}
      <div className="px-5 py-4">
        {/* Header with GitHub icon and repo name */}
        <div className="flex items-start gap-3 mb-4">
          <div className="w-10 h-10 rounded-xl bg-gray-900 flex items-center justify-center flex-shrink-0">
            <img src="/github-icon.svg" alt="" className="w-5 h-5 invert" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="font-medium text-base text-gray-900">{data.label}</div>
            <div className="text-xs text-gray-400 truncate">{displayUrl}</div>
          </div>
        </div>

        {/* Deployment Progress (shown when deploying) */}
        {isDeploying && (
          <div className="mb-4">
            <div className="flex items-center justify-between mb-2">
              {deploymentStages.slice(0, 4).map((stage, index) => {
                const isComplete = index < currentStageIndex
                const isCurrent = index === currentStageIndex
                const isPending = index > currentStageIndex
                
                return (
                  <div key={stage.id} className="flex flex-col items-center flex-1">
                    <div className={`w-6 h-6 rounded-full flex items-center justify-center ${
                      isComplete ? 'bg-emerald-500' :
                      isCurrent ? 'bg-blue-500' :
                      'bg-gray-200'
                    }`}>
                      {isComplete ? (
                        <CheckCircle className="w-4 h-4 text-white" />
                      ) : isCurrent ? (
                        <Loader2 className="w-4 h-4 text-white animate-spin" />
                      ) : (
                        <Circle className="w-3 h-3 text-gray-400" />
                      )}
                    </div>
                    <span className={`text-[10px] mt-1 ${
                      isCurrent ? 'text-blue-600 font-medium' :
                      isComplete ? 'text-emerald-600' :
                      'text-gray-400'
                    }`}>
                      {stage.label}
                    </span>
                  </div>
                )
              })}
            </div>
            {/* Progress line */}
            <div className="relative h-0.5 bg-gray-200 rounded-full mt-1">
              <div 
                className="absolute top-0 left-0 h-full bg-blue-500 rounded-full transition-all duration-500"
                style={{ width: `${Math.min((currentStageIndex / 3) * 100, 100)}%` }}
              />
            </div>
          </div>
        )}

        {/* Status section */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className={`w-2 h-2 rounded-full ${status.dot}`} />
            <span className={`text-sm font-medium ${status.text}`}>{status.label}</span>
          </div>
          
          {/* Elapsed time */}
          {(isDeploying || (elapsedTime !== '00:00' && deployment?.status === 'success')) && (
            <span className="text-xs text-gray-400 font-mono">
              {elapsedTime}
            </span>
          )}
        </div>

        {/* Failed state */}
        {isFailed && (
          <div className="mt-3 p-2 bg-red-50 rounded-lg flex items-center gap-2">
            <XCircle className="w-4 h-4 text-red-500 flex-shrink-0" />
            <span className="text-xs text-red-600">Deployment failed</span>
          </div>
        )}
      </div>

      <Handle type="source" position={Position.Bottom} className="!bg-gray-400 !w-2 !h-2" />
    </div>
  )
}

export default memo(ServiceNode)
