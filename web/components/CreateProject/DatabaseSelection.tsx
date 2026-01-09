'use client'

import { useState } from 'react'
import { Database } from 'lucide-react'
import { useDatabasesStore } from '@/stores/databasesStore'

interface DatabaseSelectionProps {
  projectId: string
  onServiceCreated: () => void
}

const databaseOptions = [
  { engine: 'postgresql', name: 'PostgreSQL', description: 'Open-source relational database' },
  { engine: 'mysql', name: 'MySQL', description: 'Popular relational database' },
  { engine: 'mongodb', name: 'MongoDB', description: 'NoSQL document database' },
]

export default function DatabaseSelection({ projectId, onServiceCreated }: DatabaseSelectionProps) {
  const { createDatabase } = useDatabasesStore()
  const [creating, setCreating] = useState<string | null>(null)

  const handleSelectDatabase = async (engine: string) => {
    setCreating(engine)
    try {
      await createDatabase(projectId, {
        engine,
        size: 'small',
      })
      onServiceCreated()
    } catch (error) {
      console.error('Failed to create database:', error)
    } finally {
      setCreating(null)
    }
  }

  return (
    <div className="h-full flex flex-col">
      <div className="flex-1 overflow-y-auto p-6">
        <div className="space-y-3">
          {databaseOptions.map((option) => (
            <button
              key={option.engine}
              onClick={() => handleSelectDatabase(option.engine)}
              disabled={creating === option.engine}
              className="w-full text-left p-4 border border-gray-200 rounded-lg hover:bg-gray-50 hover:border-blue-300 transition-colors disabled:opacity-50"
            >
              <div className="flex items-center gap-3">
                <Database className="w-5 h-5 text-gray-600" />
                <div className="flex-1">
                  <h3 className="font-medium text-gray-900">{option.name}</h3>
                  <p className="text-sm text-gray-500 mt-1">{option.description}</p>
                </div>
                {creating === option.engine && (
                  <div className="w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                )}
              </div>
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

