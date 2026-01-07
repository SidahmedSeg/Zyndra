'use client'

import Drawer from './Drawer'
import Tabs from './Tabs'
import { Database } from '@/lib/api/databases'
import MetricsTab from '@/components/Metrics/MetricsTab'

interface DatabaseDrawerProps {
  database: Database | null
  isOpen: boolean
  onClose: () => void
}

export default function DatabaseDrawer({
  database,
  isOpen,
  onClose,
}: DatabaseDrawerProps) {
  if (!database) return null

  const tabs = [
    {
      id: 'config',
      label: 'Config',
      content: <ConfigTab database={database} />,
    },
    {
      id: 'credentials',
      label: 'Credentials',
      content: <CredentialsTab database={database} />,
    },
    {
      id: 'backups',
      label: 'Backups',
      content: <BackupsTab database={database} />,
    },
    {
      id: 'logs',
      label: 'Logs',
      content: <LogsTab database={database} />,
    },
    {
      id: 'metrics',
      label: 'Metrics',
      content: <MetricsTab resourceId={database.id} resourceType="database" />,
    },
  ]

  return (
    <Drawer
      isOpen={isOpen}
      onClose={onClose}
      title={`Database: ${database.engine}`}
      width="900px"
    >
      <Tabs tabs={tabs} />
    </Drawer>
  )
}

function ConfigTab({ database }: { database: Database }) {
  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Engine
        </label>
        <p className="text-sm text-gray-500 capitalize">{database.engine}</p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Version
        </label>
        <p className="text-sm text-gray-500">{database.version || 'Default'}</p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Size
        </label>
        <p className="text-sm text-gray-500 capitalize">{database.size}</p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Volume Size
        </label>
        <p className="text-sm text-gray-500">{database.volume_size_mb} MB</p>
      </div>
    </div>
  )
}

function CredentialsTab({ database }: { database: Database }) {
  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Hostname
        </label>
        <p className="text-sm text-gray-500 font-mono">
          {database.internal_hostname || 'Not available'}
        </p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Port
        </label>
        <p className="text-sm text-gray-500">{database.port || 'Not available'}</p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Connection URL
        </label>
        <p className="text-sm text-gray-500 font-mono break-all">
          {database.connection_url || 'Not available'}
        </p>
      </div>
    </div>
  )
}

function BackupsTab({ database }: { database: Database }) {
  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-500">Backup management will be available here</p>
    </div>
  )
}

function LogsTab({ database }: { database: Database }) {
  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-500">Database logs will be displayed here</p>
    </div>
  )
}

