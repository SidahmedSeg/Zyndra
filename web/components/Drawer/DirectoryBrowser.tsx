'use client'

import { useState, useEffect, useCallback } from 'react'
import { Folder, ChevronRight, ChevronDown, Loader2, File, ArrowLeft } from 'lucide-react'
import { gitApi, TreeEntry } from '@/lib/api/git'

interface DirectoryBrowserProps {
  owner: string
  repo: string
  branch: string
  currentPath: string
  onSelect: (path: string) => void
  onClose: () => void
}

export default function DirectoryBrowser({ 
  owner, 
  repo, 
  branch, 
  currentPath, 
  onSelect,
  onClose
}: DirectoryBrowserProps) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [entries, setEntries] = useState<TreeEntry[]>([])
  const [browsePath, setBrowsePath] = useState(currentPath === '/' ? '' : currentPath)
  const [pathHistory, setPathHistory] = useState<string[]>([''])

  const loadDirectory = useCallback(async (path: string) => {
    try {
      setLoading(true)
      setError(null)
      const tree = await gitApi.getRepoTree(owner, repo, branch, path)
      // Filter to only show directories (trees)
      const directories = tree.filter(entry => entry.type === 'tree')
      // Sort alphabetically
      directories.sort((a, b) => a.name.localeCompare(b.name))
      setEntries(directories)
    } catch (err) {
      console.error('Failed to load directory:', err)
      setError('Failed to load repository contents')
    } finally {
      setLoading(false)
    }
  }, [owner, repo, branch])

  useEffect(() => {
    if (owner && repo && branch) {
      loadDirectory(browsePath)
    }
  }, [owner, repo, branch, browsePath, loadDirectory])

  const handleNavigate = (path: string) => {
    setPathHistory([...pathHistory, path])
    setBrowsePath(path)
  }

  const handleBack = () => {
    if (pathHistory.length > 1) {
      const newHistory = [...pathHistory]
      newHistory.pop()
      const previousPath = newHistory[newHistory.length - 1]
      setPathHistory(newHistory)
      setBrowsePath(previousPath)
    }
  }

  const handleSelect = (path: string) => {
    // Convert path to format expected (with leading /)
    const selectedPath = path ? `/${path}` : '/'
    onSelect(selectedPath)
    onClose()
  }

  const displayPath = browsePath || '/'

  return (
    <div className="absolute top-full left-0 right-0 mt-1 bg-white border border-gray-200 rounded-xl shadow-lg z-20 overflow-hidden max-h-[300px]">
      {/* Header */}
      <div className="flex items-center gap-2 px-3 py-2 bg-gray-50 border-b border-gray-200">
        {pathHistory.length > 1 && (
          <button
            onClick={handleBack}
            className="p-1 hover:bg-gray-200 rounded transition-colors"
          >
            <ArrowLeft className="w-4 h-4 text-gray-500" />
          </button>
        )}
        <Folder className="w-4 h-4 text-gray-400" />
        <span className="text-sm text-gray-600 font-mono flex-1 truncate">{displayPath}</span>
        <button
          onClick={() => handleSelect(browsePath)}
          className="px-2 py-1 text-xs font-medium text-indigo-600 hover:bg-indigo-50 rounded transition-colors"
        >
          Select
        </button>
      </div>

      {/* Content */}
      <div className="overflow-y-auto max-h-[240px]">
        {loading ? (
          <div className="flex items-center justify-center py-6">
            <Loader2 className="w-5 h-5 animate-spin text-gray-400" />
          </div>
        ) : error ? (
          <div className="px-4 py-6 text-center text-sm text-red-500">
            {error}
          </div>
        ) : entries.length === 0 ? (
          <div className="px-4 py-6 text-center text-sm text-gray-400">
            No subdirectories found
          </div>
        ) : (
          <div className="py-1">
            {entries.map((entry) => (
              <button
                key={entry.path}
                onClick={() => handleNavigate(entry.path)}
                className="w-full flex items-center gap-3 px-3 py-2 hover:bg-gray-50 transition-colors text-left"
              >
                <Folder className="w-4 h-4 text-amber-500" />
                <span className="text-sm text-gray-700 flex-1">{entry.name}</span>
                <ChevronRight className="w-4 h-4 text-gray-300" />
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

