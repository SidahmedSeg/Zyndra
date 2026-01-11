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
  const deploymentStatus = deployment ? getStatusDisplay(deployment.status) : null
  
  // Service status when not deploying
  const serviceStatus = service.status === 'running' 
    ? { label: 'Online', color: 'text-emerald-500' }
    : service.status === 'error'
    ? { label: 'Failed', color: 'text-red-500' }
    : { label: service.status, color: 'text-gray-500' }

  // Use deployment status if deploying, otherwise use service status
  const displayStatus = isDeploying && deploymentStatus
    ? deploymentStatus
    : deployment?.status === 'success' 
    ? { label: 'Online', color: 'text-emerald-500' }
    : deployment?.status === 'failed'
    ? { label: 'Failed', color: 'text-red-500' }
    : serviceStatus

  // Deployment phase text (shown next to Online)
  const deploymentPhase = isDeploying && deploymentStatus ? deploymentStatus.label : null

  // Parse URL for display
  const displayUrl = service.generated_url || 'App custom link here'

  return (
    <div
      onClick={handleClick}
      className={`rounded-2xl min-w-[280px] max-w-[320px] cursor-pointer bg-white shadow-md overflow-hidden border ${
        selected ? 'border-gray-400' : 'border-gray-200'
      }`}
    >
      <Handle type="target" position={Position.Top} className="!bg-gray-400 !w-2 !h-2" />
      
      {/* Content */}
      <div className="px-5 py-4">
        {/* Header with GitHub icon and repo name */}
        <div className="flex items-start gap-3 mb-6">
          <img src="/github-icon.svg" alt="" className="w-6 h-6 mt-0.5" />
          <div className="flex-1 min-w-0">
            <div className="font-medium text-base text-gray-900">{data.label}</div>
            <div className="text-sm text-gray-400">{displayUrl}</div>
          </div>
        </div>

        {/* Status section */}
        <div className="flex items-center gap-2">
          {/* Green dot for Online */}
          <span className="w-2 h-2 rounded-full bg-emerald-500" />
          <span className="text-sm text-emerald-500 font-medium">Online</span>
          
          {/* Deployment phase if deploying */}
          {deploymentPhase && (
            <span className="text-sm text-gray-400 ml-1">
              {deploymentPhase} ({elapsedTime})
            </span>
          )}
          
          {/* Show elapsed time even if not deploying but just finished */}
          {!deploymentPhase && elapsedTime !== '00:00' && (
            <span className="text-sm text-gray-400 ml-1">
              ({elapsedTime})
            </span>
          )}
        </div>
      </div>

      <Handle type="source" position={Position.Bottom} className="!bg-gray-400 !w-2 !h-2" />
    </div>
  )
}

export default memo(ServiceNode)
