'use client'

import { memo } from 'react'
import { Handle, Position, NodeProps } from 'reactflow'
import { Loader2 } from 'lucide-react'
import type { Database as DatabaseType } from '@/lib/api/databases'
import { useDatabasesStore } from '@/stores/databasesStore'

interface DatabaseNodeData {
  label: string
  database: DatabaseType
}

function DatabaseNode({ data, selected }: NodeProps<DatabaseNodeData>) {
  const { database } = data
  const { setSelectedDatabase } = useDatabasesStore()
  
  // Engine-specific icons (SVG or image paths)
  const getEngineConfig = () => {
    switch (database.engine) {
      case 'postgresql':
        return { 
          icon: '/db-icons/postgres.svg', 
          label: 'Postgres',
          fallbackEmoji: '🐘',
        }
      case 'mongodb':
        return { 
          icon: '/db-icons/mongodb.svg', 
          label: 'MongoDB',
          fallbackEmoji: '🍃',
        }
      case 'redis':
        return { 
          icon: '/db-icons/redis.svg', 
          label: 'Redis',
          fallbackEmoji: '⚡',
        }
      case 'mysql':
        return { 
          icon: '/db-icons/mysql.svg', 
          label: 'MySQL',
          fallbackEmoji: '🐬',
        }
      default:
        return { 
          icon: null, 
          label: database.engine,
          fallbackEmoji: '🗄️',
        }
    }
  }

  const getStatusConfig = () => {
    switch (database.status) {
      case 'active':
        return { dot: 'bg-emerald-500', text: 'text-emerald-500', label: 'Online' }
      case 'error':
      case 'failed':
        return { dot: 'bg-red-500', text: 'text-red-500', label: 'Failed' }
      case 'pending':
      case 'provisioning':
        return { dot: 'bg-amber-500 animate-pulse', text: 'text-amber-500', label: 'Provisioning' }
      default:
        return { dot: 'bg-gray-400', text: 'text-gray-400', label: database.status }
    }
  }

  const engineConfig = getEngineConfig()
  const status = getStatusConfig()

  return (
    <div
      onClick={() => setSelectedDatabase(database)}
      className="rounded-2xl min-w-[320px] max-w-[400px] cursor-pointer bg-white shadow-md overflow-hidden border-2 transition-all duration-200"
      style={{ borderColor: selected ? '#4F46E5' : '#e5e7eb' }}
    >
      <Handle type="target" position={Position.Top} className="!bg-gray-400 !w-2 !h-2" />
      
      {/* Content */}
      <div className="px-6 py-5">
        {/* Header with engine icon and name */}
        <div className="flex items-start gap-4 mb-8">
          {/* Engine icon */}
          {engineConfig.icon ? (
            <img 
              src={engineConfig.icon} 
              alt={engineConfig.label} 
              className="w-10 h-10 flex-shrink-0"
              onError={(e) => {
                // Fallback to emoji if image fails
                const target = e.target as HTMLImageElement
                target.style.display = 'none'
                const parent = target.parentElement
                if (parent) {
                  const span = document.createElement('span')
                  span.className = 'text-2xl'
                  span.textContent = engineConfig.fallbackEmoji
                  parent.appendChild(span)
                }
              }}
            />
          ) : (
            <span className="text-2xl flex-shrink-0">{engineConfig.fallbackEmoji}</span>
          )}
          <div className="flex-1 min-w-0">
            <div className="font-semibold text-lg text-gray-900">{engineConfig.label}</div>
            <div className="text-sm text-gray-400">{database.volume_size_mb || 500} MB volume</div>
          </div>
        </div>

        {/* Status section */}
        <div className="flex items-center gap-3">
          <span className={`w-2.5 h-2.5 rounded-full ${status.dot}`} />
          <span className={`text-base font-medium ${status.text}`}>
            {status.label}
          </span>
          {status.label === 'Provisioning' && (
            <Loader2 className="w-4 h-4 text-amber-500 animate-spin" />
          )}
        </div>
      </div>

      <Handle type="source" position={Position.Bottom} className="!bg-gray-400 !w-2 !h-2" />
    </div>
  )
}

export default memo(DatabaseNode)
