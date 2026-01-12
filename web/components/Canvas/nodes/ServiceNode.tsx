'use client'

import { memo, useEffect, useState } from 'react'
import { Handle, Position, NodeProps } from 'reactflow'
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

  const handleClick = () => {
    setSelectedService(service)
  }

  // Poll for latest deployment status (reduced frequency to avoid rate limits)
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
        // Silently fail - don't spam console
      }
    }

    // Initial fetch
    fetchLatestDeployment()

    // Poll every 10 seconds (reduced from 3 to avoid rate limits)
    intervalId = setInterval(fetchLatestDeployment, 10000)

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
  
  // Get deployment phase label for display
  const getDeploymentPhase = (): string | null => {
    if (!deployment || ['success', 'failed', 'cancelled'].includes(deployment.status)) {
      return null
    }
    return getStatusDisplay(deployment.status).label
  }

  // Parse URL for display
  const displayUrl = service.generated_url || 'App custom link here'

  // Status indicator
  const getStatusInfo = () => {
    if (isFailed) {
      return { dot: 'bg-red-500', text: 'text-red-500', label: 'Failed' }
    }
    if (isOnline) {
      return { dot: 'bg-emerald-500', text: 'text-emerald-500', label: 'Online' }
    }
    if (isDeploying) {
      return { dot: 'bg-emerald-500', text: 'text-emerald-500', label: 'Online' }
    }
    return { dot: 'bg-gray-400', text: 'text-gray-400', label: 'Idle' }
  }

  const statusInfo = getStatusInfo()
  const deploymentPhase = getDeploymentPhase()

  return (
    <div
      onClick={handleClick}
      className="rounded-2xl min-w-[320px] max-w-[400px] cursor-pointer bg-white shadow-md overflow-hidden border-2 transition-all duration-200"
      style={{ borderColor: selected ? '#4F46E5' : '#e5e7eb' }}
    >
      <Handle type="target" position={Position.Top} className="!bg-gray-400 !w-2 !h-2" />
      
      {/* Content */}
      <div className="px-6 py-5">
        {/* Header with GitHub icon and repo name */}
        <div className="flex items-start gap-4 mb-8">
          <div className="w-12 h-12 rounded-xl bg-gray-900 flex items-center justify-center flex-shrink-0">
            <img src="/github-icon.svg" alt="" className="w-6 h-6 invert" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="font-semibold text-lg text-gray-900">{data.label}</div>
            <div className="text-sm text-gray-400">{displayUrl}</div>
          </div>
        </div>

        {/* Status section - simple line with Online + Phase + Time */}
        <div className="flex items-center gap-3">
          {/* Status dot */}
          <span className={`w-2.5 h-2.5 rounded-full ${statusInfo.dot}`} />
          
          {/* Status label */}
          <span className={`text-base font-medium ${statusInfo.text}`}>
            {statusInfo.label}
          </span>
          
          {/* Deployment phase with time (shown when deploying) */}
          {deploymentPhase && (
            <span className="text-base text-gray-400">
              {deploymentPhase} ({elapsedTime})
            </span>
          )}
        </div>
      </div>

      <Handle type="source" position={Position.Bottom} className="!bg-gray-400 !w-2 !h-2" />
    </div>
  )
}

export default memo(ServiceNode)
