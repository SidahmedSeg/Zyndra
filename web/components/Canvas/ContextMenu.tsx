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

const databaseOptions: { type: DatabaseType; label: string; color: string }[] = [
  { type: 'postgresql', label: 'PostgreSQL', color: 'text-blue-600' },
  { type: 'mongodb', label: 'MongoDB', color: 'text-emerald-600' },
  { type: 'redis', label: 'Redis', color: 'text-red-600' },
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
      const submenuWidth = 140
      if (x + 160 + submenuWidth > window.innerWidth) {
        setDbMenuPosition('left')
      }
    }
  }, [x])

  const handleAdd = (type: 'github-repo' | 'database' | 'volume', dbType?: DatabaseType) => {
    onAddNode(type, dbType)
    onClose()
  }

  const toggleDbMenu = () => {
    setShowDbMenu(!showDbMenu)
  }

  return (
    <div
      ref={menuRef}
      className="fixed bg-white border border-gray-200 rounded-lg shadow-lg z-50 min-w-[160px] py-1 overflow-visible"
      style={{ left: `${x}px`, top: `${y}px` }}
    >
      <button
        onClick={() => handleAdd('github-repo')}
        className="w-full text-left px-3 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2.5 transition-colors"
      >
        <Github className="w-4 h-4 text-gray-700" />
        <span>GitHub Repo</span>
      </button>
      
      {/* Database with submenu */}
      <div className="relative">
        <button
          onClick={toggleDbMenu}
          className="w-full text-left px-3 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2.5 transition-colors"
        >
          <Database className="w-4 h-4 text-gray-700" />
          <span className="flex-1">Database</span>
          <ChevronRight className={`w-3.5 h-3.5 text-gray-400 transition-transform ${showDbMenu ? 'rotate-90' : ''}`} />
        </button>
        
        {/* Database submenu */}
        {showDbMenu && (
          <div 
            className={`absolute top-0 bg-white border border-gray-200 rounded-lg shadow-lg min-w-[130px] py-1 ${
              dbMenuPosition === 'right' ? 'left-full ml-1' : 'right-full mr-1'
            }`}
          >
            {databaseOptions.map((db) => (
              <button
                key={db.type}
                onClick={() => handleAdd('database', db.type)}
                className="w-full text-left px-3 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2 transition-colors"
              >
                <span className={`font-medium ${db.color}`}>{db.label}</span>
              </button>
            ))}
          </div>
        )}
      </div>
      
      <button
        onClick={() => handleAdd('volume')}
        className="w-full text-left px-3 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2.5 transition-colors"
      >
        <HardDrive className="w-4 h-4 text-gray-700" />
        <span>Volume</span>
      </button>
    </div>
  )
}
