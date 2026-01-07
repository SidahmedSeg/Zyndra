'use client'

import { memo } from 'react'
import { Handle, Position, NodeProps } from 'reactflow'
import { Database, CheckCircle, XCircle, Clock } from 'lucide-react'
import type { Database as DatabaseType } from '@/lib/api/databases'
import { useDatabasesStore } from '@/stores/databasesStore'

interface DatabaseNodeData {
  label: string
  database: DatabaseType
}

function DatabaseNode({ data, selected }: NodeProps<DatabaseNodeData>) {
  const { database } = data
  const { setSelectedDatabase } = useDatabasesStore()
  const statusColor = {
    active: 'bg-green-500',
    error: 'bg-red-500',
    pending: 'bg-yellow-500',
    provisioning: 'bg-blue-500',
  }[database.status] || 'bg-gray-400'

  const StatusIcon = {
    active: CheckCircle,
    error: XCircle,
    pending: Clock,
    provisioning: Clock,
  }[database.status] || Clock

  const engineColors = {
    postgresql: 'text-blue-600',
    mysql: 'text-orange-600',
    redis: 'text-red-600',
  }

  return (
    <div
      className={`px-4 py-3 shadow-lg rounded-lg border-2 min-w-[200px] cursor-pointer ${
        selected ? 'border-blue-500' : 'border-gray-300'
      } bg-white`}
      onClick={() => setSelectedDatabase(database)}
    >
      <Handle type="target" position={Position.Top} />
      
      <div className="flex items-center gap-2 mb-2">
        <Database className={`w-5 h-5 ${engineColors[database.engine as keyof typeof engineColors] || 'text-gray-600'}`} />
        <div className="flex-1">
          <div className="font-semibold text-sm">{data.label}</div>
          <div className="text-xs text-gray-500 capitalize">{database.engine}</div>
        </div>
        <div className={`w-3 h-3 rounded-full ${statusColor}`} />
      </div>

      <div className="flex items-center gap-2 text-xs text-gray-600">
        <StatusIcon className="w-3 h-3" />
        <span className="capitalize">{database.status}</span>
      </div>

      {database.internal_hostname && (
        <div className="mt-2 text-xs text-gray-500 truncate">
          {database.internal_hostname}
        </div>
      )}

      <Handle type="source" position={Position.Bottom} />
    </div>
  )
}

export default memo(DatabaseNode)

