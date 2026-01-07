'use client'

import { memo } from 'react'
import { Handle, Position, NodeProps } from 'reactflow'
import { Server, Play, Square, AlertCircle } from 'lucide-react'
import type { Service } from '@/lib/api/services'
import { useServicesStore } from '@/stores/servicesStore'

interface ServiceNodeData {
  label: string
  service: Service
}

function ServiceNode({ data, selected }: NodeProps<ServiceNodeData>) {
  const { setSelectedService } = useServicesStore()

  const handleClick = () => {
    setSelectedService(data.service)
  }
  const { service } = data
  const statusColor = {
    running: 'bg-green-500',
    stopped: 'bg-gray-400',
    error: 'bg-red-500',
    pending: 'bg-yellow-500',
    deploying: 'bg-blue-500',
  }[service.status] || 'bg-gray-400'

  const StatusIcon = {
    running: Play,
    stopped: Square,
    error: AlertCircle,
    pending: AlertCircle,
    deploying: Play,
  }[service.status] || Square

  return (
    <div
      className={`px-4 py-3 shadow-lg rounded-lg border-2 min-w-[200px] cursor-pointer ${
        selected ? 'border-blue-500' : 'border-gray-300'
      } bg-white`}
      onClick={handleClick}
    >
      <Handle type="target" position={Position.Top} />
      
      <div className="flex items-center gap-2 mb-2">
        <Server className="w-5 h-5 text-blue-600" />
        <div className="flex-1">
          <div className="font-semibold text-sm">{data.label}</div>
          <div className="text-xs text-gray-500">{service.type}</div>
        </div>
        <div className={`w-3 h-3 rounded-full ${statusColor}`} />
      </div>

      <div className="flex items-center gap-2 text-xs text-gray-600">
        <StatusIcon className="w-3 h-3" />
        <span className="capitalize">{service.status}</span>
      </div>

      {service.generated_url && (
        <div className="mt-2 text-xs text-blue-600 truncate">
          {service.generated_url}
        </div>
      )}

      <Handle type="source" position={Position.Bottom} />
    </div>
  )
}

export default memo(ServiceNode)

