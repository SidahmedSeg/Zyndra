'use client'

import { Folder, Calendar, ArrowRight } from 'lucide-react'
import { format } from 'date-fns'
import type { Project } from '@/lib/api/projects'

interface ProjectCardProps {
  project: Project
  onClick: () => void
}

export default function ProjectCard({ project, onClick }: ProjectCardProps) {
  return (
    <div
      onClick={onClick}
      className="bg-white border border-gray-200 rounded-lg p-6 hover:shadow-lg transition-all cursor-pointer group"
    >
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3">
          <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
            <Folder className="w-6 h-6 text-white" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-gray-900 group-hover:text-blue-600 transition-colors">
              {project.name}
            </h3>
            {project.slug && (
              <p className="text-sm text-gray-500">/{project.slug}</p>
            )}
          </div>
        </div>
        <ArrowRight className="w-5 h-5 text-gray-400 group-hover:text-blue-600 transition-colors" />
      </div>

      {project.description && (
        <p className="text-gray-600 mb-4 line-clamp-2">{project.description}</p>
      )}

      <div className="flex items-center gap-4 text-sm text-gray-500">
        <div className="flex items-center gap-1">
          <Calendar className="w-4 h-4" />
          <span>
            {format(new Date(project.created_at), 'MMM d, yyyy')}
          </span>
        </div>
      </div>
    </div>
  )
}

