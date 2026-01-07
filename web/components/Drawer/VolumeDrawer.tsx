'use client'

import Drawer from './Drawer'
import Tabs from './Tabs'
import { Volume } from '@/lib/api/volumes'
import MetricsTab from '@/components/Metrics/MetricsTab'

interface VolumeDrawerProps {
  volume: Volume | null
  isOpen: boolean
  onClose: () => void
}

export default function VolumeDrawer({
  volume,
  isOpen,
  onClose,
}: VolumeDrawerProps) {
  if (!volume) return null

  const tabs = [
    {
      id: 'config',
      label: 'Config',
      content: <ConfigTab volume={volume} />,
    },
    {
      id: 'attached',
      label: 'Attached To',
      content: <AttachedTab volume={volume} />,
    },
    {
      id: 'usage',
      label: 'Usage',
      content: <UsageTab volume={volume} />,
    },
    {
      id: 'metrics',
      label: 'Metrics',
      content: <MetricsTab resourceId={volume.id} resourceType="volume" />,
    },
  ]

  return (
    <Drawer
      isOpen={isOpen}
      onClose={onClose}
      title={`Volume: ${volume.name}`}
      width="900px"
    >
      <Tabs tabs={tabs} />
    </Drawer>
  )
}

function ConfigTab({ volume }: { volume: Volume }) {
  const sizeGB = (volume.size_mb / 1024).toFixed(1)

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Name
        </label>
        <p className="text-sm text-gray-500">{volume.name}</p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Size
        </label>
        <p className="text-sm text-gray-500">{sizeGB} GB ({volume.size_mb} MB)</p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Type
        </label>
        <p className="text-sm text-gray-500 capitalize">{volume.volume_type}</p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Status
        </label>
        <p className="text-sm text-gray-500 capitalize">{volume.status}</p>
      </div>
    </div>
  )
}

function AttachedTab({ volume }: { volume: Volume }) {
  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Attached To Service
        </label>
        <p className="text-sm text-gray-500">
          {volume.attached_to_service_id || 'Not attached'}
        </p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Attached To Database
        </label>
        <p className="text-sm text-gray-500">
          {volume.attached_to_database_id || 'Not attached'}
        </p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Mount Path
        </label>
        <p className="text-sm text-gray-500 font-mono">
          {volume.mount_path || 'Not set'}
        </p>
      </div>
    </div>
  )
}

function UsageTab({ volume }: { volume: Volume }) {
  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-500">Volume usage statistics will be displayed here</p>
    </div>
  )
}

