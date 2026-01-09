'use client'

import { useState } from 'react'
import { HardDrive } from 'lucide-react'
import { useVolumesStore } from '@/stores/volumesStore'

interface VolumeSelectionProps {
  projectId: string
  onServiceCreated: () => void
}

export default function VolumeSelection({ projectId, onServiceCreated }: VolumeSelectionProps) {
  const { createVolume } = useVolumesStore()
  const [creating, setCreating] = useState(false)

  const handleCreateVolume = async () => {
    setCreating(true)
    try {
      await createVolume(projectId, {
        name: 'New Volume',
        size_mb: 1024,
      })
      onServiceCreated()
    } catch (error) {
      console.error('Failed to create volume:', error)
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="h-full flex flex-col">
      <div className="flex-1 overflow-y-auto p-6">
        <div className="flex flex-col items-center justify-center h-full">
          <div className="text-center mb-6">
            <HardDrive className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Create Storage Volume</h3>
            <p className="text-gray-500">Persistent storage for your services</p>
          </div>
          
          <button
            onClick={handleCreateVolume}
            disabled={creating}
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {creating ? (
              <>
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                <span>Creating...</span>
              </>
            ) : (
              <span>Create Volume</span>
            )}
          </button>
        </div>
      </div>
    </div>
  )
}

