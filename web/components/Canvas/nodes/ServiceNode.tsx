'use client'

import { memo, useEffect, useState } from 'react'
import { Handle, Position, NodeProps } from 'reactflow'
import { Loader2 } from 'lucide-react'
import type { Service } from '@/lib/api/services'
import { useServicesStore } from '@/stores/servicesStore'
import { deploymentsApi, type Deployment, getStatusDisplay } from '@/lib/api/deployments'

interface ServiceNodeData {
  label: string
  service: Service
}

function ServiceNode({ data, selected }: NodeProps<ServiceNodeData>) {
  const { setSelectedService } = useServicesStore()
  const { service } = data
  
  const [deployment, setDeployment] = useState<Deployment | null>(null)
  const [elapsedTime, setElapsedTime] = useState<string>('00:00')

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

  const handleClick = () => {
    setSelectedService(data.service)
  }

  // Determine display status based on deployment
  const isDeploying = deployment && !['success', 'failed', 'cancelled'].includes(deployment.status)
  const deploymentStatus = deployment ? getStatusDisplay(deployment.status) : null
  
  // Service status when not deploying
  const serviceStatus = service.status === 'running' 
    ? { label: 'Online', color: 'text-emerald-500', dot: 'bg-emerald-500' }
    : service.status === 'error'
    ? { label: 'Failed', color: 'text-red-500', dot: 'bg-red-500' }
    : { label: service.status, color: 'text-gray-500', dot: 'bg-gray-400' }

  // Use deployment status if deploying, otherwise use service status
  const displayStatus = isDeploying && deploymentStatus
    ? deploymentStatus
    : deployment?.status === 'success' 
    ? { label: 'Online', color: 'text-emerald-500' }
    : deployment?.status === 'failed'
    ? { label: 'Failed', color: 'text-red-500' }
    : serviceStatus

  // Parse URL for display
  const displayUrl = service.generated_url || 'App custom link here'

  return (
    <div
      className={`rounded-xl border-2 min-w-[200px] max-w-[220px] cursor-pointer bg-white shadow-sm overflow-hidden ${
        selected ? 'border-cyan-500' : 'border-cyan-400'
      }`}
      onClick={handleClick}
    >
      <Handle type="target" position={Position.Top} className="!bg-cyan-500" />
      
      {/* Header */}
      <div className="px-4 pt-3 pb-2">
        <div className="flex items-start gap-2">
          <img src="/github-icon.svg" alt="" className="w-5 h-5 opacity-70 mt-0.5" />
          <div className="flex-1 min-w-0">
            <div className="font-medium text-sm text-gray-900 truncate">{data.label}</div>
            <div className="text-xs text-gray-400 truncate">{displayUrl}</div>
          </div>
        </div>
      </div>

      {/* Status bar */}
      <div className="px-4 py-2 border-t border-gray-100">
        <div className="flex items-center gap-2">
          {isDeploying ? (
            <>
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              <Loader2 className={`w-3 h-3 ${displayStatus.color} animate-spin`} />
            </>
          ) : (
            <span className={`w-2 h-2 rounded-full ${
              displayStatus.label === 'Online' ? 'bg-emerald-500' : 
              displayStatus.label === 'Failed' ? 'bg-red-500' : 'bg-gray-400'
            }`} />
          )}
          <span className={`text-xs ${displayStatus.color}`}>
            {displayStatus.label}
          </span>
          {isDeploying && (
            <span className="text-xs text-gray-400 ml-auto">
              ({elapsedTime})
            </span>
          )}
        </div>
      </div>

      <Handle type="source" position={Position.Bottom} className="!bg-cyan-500" />
    </div>
  )
}

export default memo(ServiceNode)
