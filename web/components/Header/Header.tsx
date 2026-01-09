'use client'

import { User, Plus, UserPlus, Settings } from 'lucide-react'
import { useState } from 'react'

interface HeaderProps {
  onCreateProject: () => void
  onInvite: () => void
  onSettings: () => void
}

export default function Header({ onCreateProject, onInvite, onSettings }: HeaderProps) {
  return (
    <header className="border-b bg-white">
      <div className="container mx-auto px-6 py-4">
        {/* Top row: Logo and User Avatar */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center">
            <h1 className="text-2xl font-bold text-gray-900">Zyndra</h1>
          </div>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-semibold">
              <User className="w-5 h-5" />
            </div>
          </div>
        </div>

        {/* Bottom row: Project title with divider on left, Action buttons on right */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h2 className="text-lg font-semibold text-gray-800">Projects</h2>
            <div className="h-6 w-px bg-gray-300" />
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={onCreateProject}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              <Plus className="w-4 h-4" />
              <span>Create</span>
            </button>
            <button
              onClick={onInvite}
              className="flex items-center gap-2 px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
            >
              <UserPlus className="w-4 h-4" />
              <span>Invite</span>
            </button>
            <button
              onClick={onSettings}
              className="flex items-center gap-2 px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
            >
              <Settings className="w-4 h-4" />
              <span>Settings</span>
            </button>
          </div>
        </div>
      </div>
    </header>
  )
}

