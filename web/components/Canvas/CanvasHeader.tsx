'use client'

import AppHeader from '@/components/Header/AppHeader'

type TabType = 'architecture' | 'logs' | 'settings'

interface CanvasHeaderProps {
  projectId: string
  onEnvironmentChange?: (env: string) => void
  activeTab?: TabType
  onTabChange?: (tab: TabType) => void
}

export default function CanvasHeader({ 
  projectId, 
  onEnvironmentChange,
  activeTab = 'architecture',
  onTabChange
}: CanvasHeaderProps) {
  return (
    <AppHeader 
      variant="canvas"
      projectId={projectId}
      onEnvironmentChange={onEnvironmentChange}
      activeTab={activeTab}
      onTabChange={onTabChange}
    />
  )
}
