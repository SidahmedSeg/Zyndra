'use client'

import { memo } from 'react'
import { Handle, Position, NodeProps } from 'reactflow'
import { Database, CheckCircle, XCircle, Loader2 } from 'lucide-react'
import type { Database as DatabaseType } from '@/lib/api/databases'
import { useDatabasesStore } from '@/stores/databasesStore'

interface DatabaseNodeData {
  label: string
  database: DatabaseType
}

function DatabaseNode({ data, selected }: NodeProps<DatabaseNodeData>) {
  const { database } = data
  const { setSelectedDatabase } = useDatabasesStore()
  
  const getEngineIcon = () => {
    switch (database.engine) {
      case 'postgresql': return '🐘'
      case 'mongodb': return '🍃'
      case 'redis': return '⚡'
      case 'mysql': return '🐬'
      default: return '🗄️'
    }
  }

  const getEngineColor = () => {
    switch (database.engine) {
      case 'postgresql': return 'bg-blue-100'
      case 'mongodb': return 'bg-emerald-100'
      case 'redis': return 'bg-red-100'
      case 'mysql': return 'bg-orange-100'
      default: return 'bg-gray-100'
    }
  }

  const getStatusConfig = () => {
    switch (database.status) {
      case 'active':
        return { dot: 'bg-emerald-500', text: 'text-emerald-500', label: 'Running', Icon: CheckCircle }
      case 'error':
      case 'failed':
        return { dot: 'bg-red-500', text: 'text-red-500', label: 'Failed', Icon: XCircle }
      case 'pending':
      case 'provisioning':
        return { dot: 'bg-blue-500 animate-pulse', text: 'text-blue-500', label: 'Provisioning', Icon: Loader2 }
      default:
        return { dot: 'bg-gray-400', text: 'text-gray-400', label: database.status, Icon: Database }
    }
  }

  const status = getStatusConfig()
  const displayHost = database.internal_hostname || 'Provisioning...'

  return (
    <div
      onClick={() => setSelectedDatabase(database)}
      className="rounded-2xl min-w-[260px] max-w-[300px] cursor-pointer bg-white shadow-md overflow-hidden border-2 transition-all duration-200"
      style={{ borderColor: selected ? '#4F46E5' : '#e5e7eb' }}
    >
      <Handle type="target" position={Position.Top} className="!bg-gray-400 !w-2 !h-2" />
      
      {/* Content */}
      <div className="px-5 py-4">
        {/* Header with engine icon */}
        <div className="flex items-start gap-3 mb-4">
          <div className={`w-10 h-10 rounded-xl ${getEngineColor()} flex items-center justify-center flex-shrink-0 text-xl`}>
            {getEngineIcon()}
          </div>
          <div className="flex-1 min-w-0">
            <div className="font-medium text-base text-gray-900 capitalize">{database.engine}</div>
            <div className="text-xs text-gray-400 truncate font-mono">{displayHost}</div>
          </div>
        </div>

        {/* Info row */}
        <div className="flex items-center justify-between text-xs text-gray-500 mb-3">
          <span className="capitalize">{database.size}</span>
          <span>{database.volume_size_mb} MB</span>
        </div>

        {/* Status section */}
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${status.dot}`} />
          <span className={`text-sm font-medium ${status.text}`}>{status.label}</span>
          {status.label === 'Provisioning' && (
            <Loader2 className="w-3 h-3 text-blue-500 animate-spin ml-auto" />
          )}
        </div>
      </div>

      <Handle type="source" position={Position.Bottom} className="!bg-gray-400 !w-2 !h-2" />
    </div>
  )
}

export default memo(DatabaseNode)
