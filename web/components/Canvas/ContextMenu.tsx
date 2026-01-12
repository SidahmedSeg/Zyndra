'use client'

import { useState, useEffect, useRef } from 'react'
import { Github, Database, HardDrive, ChevronRight } from 'lucide-react'

export type DatabaseType = 'postgresql' | 'mongodb' | 'redis'

interface ContextMenuProps {
  x: number
  y: number
  onClose: () => void
  onAddNode: (type: 'github-repo' | 'database' | 'volume', dbType?: DatabaseType) => void
}

const databaseOptions: { type: DatabaseType; label: string; icon: string; color: string }[] = [
  { type: 'postgresql', label: 'PostgreSQL', icon: '🐘', color: 'bg-blue-100 text-blue-700' },
  { type: 'mongodb', label: 'MongoDB', icon: '🍃', color: 'bg-emerald-100 text-emerald-700' },
  { type: 'redis', label: 'Redis', icon: '⚡', color: 'bg-red-100 text-red-700' },
]

export default function ContextMenu({ x, y, onClose, onAddNode }: ContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null)
  const [showDbMenu, setShowDbMenu] = useState(false)
  const [dbMenuPosition, setDbMenuPosition] = useState<'right' | 'left'>('right')

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        onClose()
      }
    }

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    document.addEventListener('keydown', handleEscape)
    
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
      document.removeEventListener('keydown', handleEscape)
    }
  }, [onClose])

  // Check if submenu should appear on the left
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const submenuWidth = 180
      if (x + 180 + submenuWidth > window.innerWidth) {
        setDbMenuPosition('left')
      }
    }
  }, [x])

  const handleAdd = (type: 'github-repo' | 'database' | 'volume', dbType?: DatabaseType) => {
    onAddNode(type, dbType)
    onClose()
  }

  return (
    <div
      ref={menuRef}
      className="fixed bg-white border border-gray-200 rounded-xl shadow-lg z-50 min-w-[200px] py-1.5 overflow-hidden"
      style={{ left: `${x}px`, top: `${y}px` }}
    >
      <div className="px-3 py-1.5 text-xs font-medium text-gray-400 uppercase tracking-wider">
        Add Service
      </div>
      
      <button
        onClick={() => handleAdd('github-repo')}
        className="w-full text-left px-3 py-2.5 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-3 transition-colors"
      >
        <div className="w-8 h-8 rounded-lg bg-gray-900 flex items-center justify-center">
          <Github className="w-4 h-4 text-white" />
        </div>
        <div>
          <div className="font-medium">GitHub Repo</div>
          <div className="text-xs text-gray-400">Deploy from repository</div>
        </div>
      </button>
      
      {/* Database with submenu */}
      <div 
        className="relative"
        onMouseEnter={() => setShowDbMenu(true)}
        onMouseLeave={() => setShowDbMenu(false)}
      >
        <button
          className="w-full text-left px-3 py-2.5 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-3 transition-colors"
        >
          <div className="w-8 h-8 rounded-lg bg-indigo-100 flex items-center justify-center">
            <Database className="w-4 h-4 text-indigo-600" />
          </div>
          <div className="flex-1">
            <div className="font-medium">Database</div>
            <div className="text-xs text-gray-400">PostgreSQL, MongoDB, Redis</div>
          </div>
          <ChevronRight className="w-4 h-4 text-gray-400" />
        </button>
        
        {/* Database submenu */}
        {showDbMenu && (
          <div 
            className={`absolute top-0 bg-white border border-gray-200 rounded-xl shadow-lg min-w-[180px] py-1.5 ${
              dbMenuPosition === 'right' ? 'left-full ml-1' : 'right-full mr-1'
            }`}
          >
            <div className="px-3 py-1.5 text-xs font-medium text-gray-400 uppercase tracking-wider">
              Select Type
            </div>
            {databaseOptions.map((db) => (
              <button
                key={db.type}
                onClick={() => handleAdd('database', db.type)}
                className="w-full text-left px-3 py-2.5 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-3 transition-colors"
              >
                <span className="text-lg">{db.icon}</span>
                <div>
                  <div className="font-medium">{db.label}</div>
                  <div className="text-xs text-gray-400">
                    {db.type === 'postgresql' && 'Relational database'}
                    {db.type === 'mongodb' && 'Document database'}
                    {db.type === 'redis' && 'In-memory cache'}
                  </div>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
      
      <button
        onClick={() => handleAdd('volume')}
        className="w-full text-left px-3 py-2.5 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-3 transition-colors"
      >
        <div className="w-8 h-8 rounded-lg bg-amber-100 flex items-center justify-center">
          <HardDrive className="w-4 h-4 text-amber-600" />
        </div>
        <div>
          <div className="font-medium">Volume</div>
          <div className="text-xs text-gray-400">Persistent storage</div>
        </div>
      </button>
    </div>
  )
}
