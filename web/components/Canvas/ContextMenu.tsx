'use client'

import { useState, useEffect, useRef } from 'react'
import { Server, Database, HardDrive } from 'lucide-react'

interface ContextMenuProps {
  x: number
  y: number
  onClose: () => void
  onAddNode: (type: 'service' | 'database' | 'volume') => void
}

export default function ContextMenu({ x, y, onClose, onAddNode }: ContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null)

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

  const handleAdd = (type: 'service' | 'database' | 'volume') => {
    onAddNode(type)
    onClose()
  }

  return (
    <div
      ref={menuRef}
      className="fixed bg-white border border-gray-200 rounded-md shadow-lg z-50 min-w-[180px] py-1"
      style={{ left: `${x}px`, top: `${y}px` }}
    >
      <button
        onClick={() => handleAdd('service')}
        className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2 transition-colors"
      >
        <Server className="w-4 h-4" />
        <span>Add Service</span>
      </button>
      <button
        onClick={() => handleAdd('database')}
        className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2 transition-colors"
      >
        <Database className="w-4 h-4" />
        <span>Add Database</span>
      </button>
      <button
        onClick={() => handleAdd('volume')}
        className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2 transition-colors"
      >
        <HardDrive className="w-4 h-4" />
        <span>Add Volume</span>
      </button>
    </div>
  )
}

