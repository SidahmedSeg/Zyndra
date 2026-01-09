'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Header from '@/components/Header/Header'
import ServiceSelectionCard from '@/components/CreateProject/ServiceSelectionCard'

export default function CreateProjectPage() {
  const router = useRouter()

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 relative overflow-hidden">
      {/* Canvas-like background pattern */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute inset-0" style={{
          backgroundImage: `
            linear-gradient(rgba(59, 130, 246, 0.1) 1px, transparent 1px),
            linear-gradient(90deg, rgba(59, 130, 246, 0.1) 1px, transparent 1px)
          `,
          backgroundSize: '50px 50px'
        }} />
      </div>

      <div className="relative z-10">
        <Header />
        
        <div className="flex items-center justify-center min-h-[calc(100vh-80px)] px-6 py-12">
          <div className="text-center mb-8">
            <h1 className="text-4xl font-bold text-white mb-4">New project</h1>
            <p className="text-xl text-gray-300">Let&apos;s deploy your first service to production</p>
          </div>
          
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 mt-16">
            <ServiceSelectionCard onProjectCreated={(projectId) => router.push(`/canvas/${projectId}`)} />
          </div>
        </div>
      </div>
    </div>
  )
}

