'use client'

import { memo } from 'react'
import { Handle, Position, NodeProps } from 'reactflow'
import { HardDrive, CheckCircle, XCircle, Clock, Link } from 'lucide-react'
import type { Volume } from '@/lib/api/volumes'
import { useVolumesStore } from '@/stores/volumesStore'

interface VolumeNodeData {
  label: string
  volume: Volume
}

function VolumeNode({ data, selected }: NodeProps<VolumeNodeData>) {
  const { volume } = data
  const { setSelectedVolume } = useVolumesStore()
  const statusColor = {
    available: 'bg-green-500',
    attached: 'bg-blue-500',
    error: 'bg-red-500',
    pending: 'bg-yellow-500',
  }[volume.status] || 'bg-gray-400'

  const StatusIcon = {
    available: CheckCircle,
    attached: Link,
    error: XCircle,
    pending: Clock,
  }[volume.status] || Clock

  const sizeGB = (volume.size_mb / 1024).toFixed(1)

  return (
    <div
      className={`px-4 py-3 shadow-lg rounded-lg border-2 min-w-[200px] cursor-pointer ${
        selected ? 'border-blue-500' : 'border-gray-300'
      } bg-white`}
      onClick={() => setSelectedVolume(volume)}
    >
      <Handle type="target" position={Position.Top} />
      
      <div className="flex items-center gap-2 mb-2">
        <HardDrive className="w-5 h-5 text-purple-600" />
        <div className="flex-1">
          <div className="font-semibold text-sm">{data.label}</div>
          <div className="text-xs text-gray-500">{sizeGB} GB</div>
        </div>
        <div className={`w-3 h-3 rounded-full ${statusColor}`} />
      </div>

      <div className="flex items-center gap-2 text-xs text-gray-600">
        <StatusIcon className="w-3 h-3" />
        <span className="capitalize">{volume.status}</span>
      </div>

      {volume.mount_path && (
        <div className="mt-2 text-xs text-gray-500 truncate">
          {volume.mount_path}
        </div>
      )}

      <Handle type="source" position={Position.Bottom} />
    </div>
  )
}

export default memo(VolumeNode)

