'use client'

import { memo } from 'react'
import { Handle, Position, NodeProps } from 'reactflow'
import { HardDrive } from 'lucide-react'
import type { Volume } from '@/lib/api/volumes'
import { useVolumesStore } from '@/stores/volumesStore'

interface VolumeNodeData {
  label: string
  volume: Volume
}

function VolumeNode({ data, selected }: NodeProps<VolumeNodeData>) {
  const { volume } = data
  const { setSelectedVolume } = useVolumesStore()

  // Calculate size for display
  const sizeDisplay = volume.size_mb >= 1024 
    ? `${(volume.size_mb / 1024).toFixed(1)} GB`
    : `${volume.size_mb} MB`

  return (
    <div
      onClick={() => setSelectedVolume(volume)}
      className="rounded-xl min-w-[280px] cursor-pointer bg-indigo-50 border border-indigo-100 overflow-hidden transition-all duration-200"
      style={{ borderColor: selected ? '#4F46E5' : 'rgb(224, 231, 255)' }}
    >
      <Handle type="target" position={Position.Top} className="!bg-gray-400 !w-2 !h-2" />
      
      {/* Compact content */}
      <div className="px-4 py-3 flex items-center gap-3">
        <HardDrive className="w-5 h-5 text-indigo-400" />
        <span className="text-sm font-medium text-indigo-700">{data.label}</span>
        <span className="text-xs text-indigo-400 ml-auto">{sizeDisplay}</span>
        {volume.status === 'attached' && (
          <span className="text-xs text-indigo-400">• Attached</span>
        )}
      </div>

      <Handle type="source" position={Position.Bottom} className="!bg-gray-400 !w-2 !h-2" />
    </div>
  )
}

export default memo(VolumeNode)

