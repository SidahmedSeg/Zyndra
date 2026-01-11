'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import ReactFlow, {
  Node,
  Edge,
  Controls,
  Background,
  MiniMap,
  Connection,
  useNodesState,
  useEdgesState,
  addEdge,
  ReactFlowInstance,
} from 'reactflow'
import 'reactflow/dist/style.css'

import { useServicesStore } from '@/stores/servicesStore'
import { useDatabasesStore } from '@/stores/databasesStore'
import { useVolumesStore } from '@/stores/volumesStore'
import ServiceNode from './nodes/ServiceNode'
import DatabaseNode from './nodes/DatabaseNode'
import VolumeNode from './nodes/VolumeNode'
import ServiceDrawer from '../Drawer/ServiceDrawer'
import DatabaseDrawer from '../Drawer/DatabaseDrawer'
import VolumeDrawer from '../Drawer/VolumeDrawer'
import ContextMenu from './ContextMenu'
import CanvasHeader from './CanvasHeader'
import RepoSelectionModal from './RepoSelectionModal'
import type { GitRepository } from '@/lib/api/git'
import { servicesApi } from '@/lib/api/services'

const nodeTypes = {
  service: ServiceNode,
  database: DatabaseNode,
  volume: VolumeNode,
}

interface CanvasProps {
  projectId: string
}

export default function Canvas({ projectId }: CanvasProps) {
  const { services, fetchServices, selectedService, setSelectedService, createService } = useServicesStore()
  const { databases, fetchDatabases, selectedDatabase, setSelectedDatabase, createDatabase } = useDatabasesStore()
  const { volumes, fetchVolumes, selectedVolume, setSelectedVolume, createVolume } = useVolumesStore()
  
  // Ensure arrays are always arrays (handle null/undefined)
  const servicesList = useMemo(() => Array.isArray(services) ? services : [], [services])
  const databasesList = useMemo(() => Array.isArray(databases) ? databases : [], [databases])
  const volumesList = useMemo(() => Array.isArray(volumes) ? volumes : [], [volumes])

  // Use ReactFlow's built-in state management (no external store sync)
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  
  // Context menu state
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number } | null>(null)
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null)
  const [repoModalOpen, setRepoModalOpen] = useState(false)
  const [pendingNodePosition, setPendingNodePosition] = useState<{ x: number; y: number } | null>(null)

  // When a drawer opens, pan the canvas to ensure the node is visible and not covered
  useEffect(() => {
    if (!reactFlowInstance) return
    
    const drawerWidth = 560 // Drawer width in pixels
    const hasDrawerOpen = selectedService || selectedDatabase || selectedVolume
    
    if (hasDrawerOpen) {
      // Get current viewport
      const { x, y, zoom } = reactFlowInstance.getViewport()
      // Pan left by half the drawer width to make room
      reactFlowInstance.setViewport({ x: x - (drawerWidth / 2) / zoom, y, zoom }, { duration: 300 })
    }
  }, [selectedService, selectedDatabase, selectedVolume, reactFlowInstance])

  // Load services and create nodes
  useEffect(() => {
    if (projectId) {
      fetchServices(projectId)
      fetchDatabases(projectId)
      fetchVolumes(projectId)
    }
  }, [projectId, fetchServices, fetchDatabases, fetchVolumes])

  // Convert services to nodes
  useEffect(() => {
    if (servicesList.length > 0) {
      const serviceNodes: Node[] = servicesList.map((service) => ({
        id: service.id,
        type: 'service',
        position: {
          x: service.canvas_x || Math.random() * 500,
          y: service.canvas_y || Math.random() * 500,
        },
        data: {
          label: service.name,
          service,
        },
        selected: selectedService?.id === service.id,
      }))

      // Only add nodes that don't already exist
      setNodes((currentNodes) => {
        const existingIds = new Set(currentNodes.map((n) => n.id))
        const newNodes = serviceNodes.filter((n) => !existingIds.has(n.id))
        if (newNodes.length > 0) {
          return [...currentNodes, ...newNodes]
        }
        // Update selected state for existing nodes
        return currentNodes.map((node) => ({
          ...node,
          selected: selectedService?.id === node.id,
        }))
      })
    }
  }, [servicesList, selectedService, setNodes])

  // Convert databases to nodes
  useEffect(() => {
    if (databasesList.length > 0) {
      const dbNodes: Node[] = databasesList.map((db) => ({
        id: `db:${db.id}`,
        type: 'database',
        position: {
          x: Math.random() * 500 + 600,
          y: Math.random() * 500,
        },
        data: {
          label: `${db.engine}`,
          database: db,
        },
      }))

      setNodes((currentNodes) => {
        const existingIds = new Set(currentNodes.map((n) => n.id))
        const newNodes = dbNodes.filter((n) => !existingIds.has(n.id))
        return newNodes.length > 0 ? [...currentNodes, ...newNodes] : currentNodes
      })

      // Create edges service -> database if service_id exists
      setEdges((currentEdges) => {
        const existing = new Set(currentEdges.map((e) => e.id))
        const newEdges: Edge[] = []
        for (const db of databasesList) {
          if (db.service_id) {
            const id = `edge:svc:${db.service_id}->db:${db.id}`
            if (!existing.has(id)) {
              newEdges.push({
                id,
                source: db.service_id,
                target: `db:${db.id}`,
                label: 'db',
                animated: false,
              })
            }
          }
        }
        return newEdges.length > 0 ? [...currentEdges, ...newEdges] : currentEdges
      })
    }
  }, [databasesList, setNodes, setEdges])

  // Convert volumes to nodes
  useEffect(() => {
    if (volumesList.length > 0) {
      const volNodes: Node[] = volumesList.map((v) => ({
        id: `vol:${v.id}`,
        type: 'volume',
        position: {
          x: Math.random() * 500,
          y: Math.random() * 500 + 600,
        },
        data: {
          label: v.name,
          volume: v,
        },
      }))

      setNodes((currentNodes) => {
        const existingIds = new Set(currentNodes.map((n) => n.id))
        const newNodes = volNodes.filter((n) => !existingIds.has(n.id))
        return newNodes.length > 0 ? [...currentNodes, ...newNodes] : currentNodes
      })

      // Create edges volume -> attached target
      setEdges((currentEdges) => {
        const existing = new Set(currentEdges.map((e) => e.id))
        const newEdges: Edge[] = []
        for (const v of volumesList) {
          if (v.attached_to_service_id) {
            const id = `edge:vol:${v.id}->svc:${v.attached_to_service_id}`
            if (!existing.has(id)) {
              newEdges.push({
                id,
                source: `vol:${v.id}`,
                target: v.attached_to_service_id,
                label: v.mount_path || 'volume',
                animated: false,
              })
            }
          }
          if (v.attached_to_database_id) {
            const id = `edge:vol:${v.id}->db:${v.attached_to_database_id}`
            if (!existing.has(id)) {
              newEdges.push({
                id,
                source: `vol:${v.id}`,
                target: `db:${v.attached_to_database_id}`,
                label: v.mount_path || 'volume',
                animated: false,
              })
            }
          }
        }
        return newEdges.length > 0 ? [...currentEdges, ...newEdges] : currentEdges
      })
    }
  }, [volumesList, setNodes, setEdges])

  const onConnect = useCallback(
    (params: Connection) => {
      setEdges((eds) => addEdge(params, eds))
    },
    [setEdges]
  )

  // Handle right-click on pane
  const onPaneContextMenu = useCallback(
    (event: React.MouseEvent) => {
      event.preventDefault()
      if (reactFlowInstance) {
        const point = reactFlowInstance.screenToFlowPosition({
          x: event.clientX,
          y: event.clientY,
        })
        setContextMenu({ x: event.clientX, y: event.clientY })
      }
    },
    [reactFlowInstance]
  )

  // Handle adding a node from context menu
  const handleAddNode = useCallback(
    async (type: 'github-repo' | 'database' | 'volume') => {
      if (!contextMenu || !reactFlowInstance) return

      const position = reactFlowInstance.screenToFlowPosition({
        x: contextMenu.x,
        y: contextMenu.y,
      })

      if (type === 'github-repo') {
        // Open repo selection modal
        setPendingNodePosition(position)
        setRepoModalOpen(true)
        return
      }

      try {
        if (type === 'database') {
          const database = await createDatabase(projectId, {
            engine: 'postgresql',
            size: 'small',
          })
          setSelectedDatabase(database)
        } else if (type === 'volume') {
          const volume = await createVolume(projectId, {
            name: 'New Volume',
            size_mb: 1024,
          })
          setSelectedVolume(volume)
        }
      } catch (error) {
        console.error('Failed to create node:', error)
      }
    },
    [contextMenu, reactFlowInstance, projectId, createDatabase, createVolume, setSelectedDatabase, setSelectedVolume]
  )

  // Handle repo selection from modal
  const handleRepoSelect = useCallback(
    async (repo: GitRepository) => {
      if (!pendingNodePosition) return

      try {
        // Use owner from repo or parse from full_name
        const owner = repo.owner || repo.full_name.split('/')[0]
        const repoName = repo.name
        
        const service = await createService(projectId, {
          name: repoName,
          type: 'app',
          instance_size: 'small',
          port: 8080,
          canvas_x: Math.round(pendingNodePosition.x),
          canvas_y: Math.round(pendingNodePosition.y),
          git_source: {
            provider: 'github',
            repo_owner: owner,
            repo_name: repoName,
            branch: repo.default_branch || 'main',
          },
        })
        
        // Trigger deployment
        try {
          await servicesApi.triggerDeployment(service.id)
        } catch (deployErr) {
          console.warn('Failed to trigger deployment:', deployErr)
        }
        
        setSelectedService(service)
        setPendingNodePosition(null)
        // Refresh services to show the new one
        fetchServices(projectId)
      } catch (error) {
        console.error('Failed to create service from repo:', error)
      }
    },
    [pendingNodePosition, projectId, createService, setSelectedService, fetchServices]
  )

  return (
    <div className="w-full h-screen flex flex-col">
      <CanvasHeader projectId={projectId} />
      <div className="flex-1 relative">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onInit={setReactFlowInstance}
          onPaneContextMenu={onPaneContextMenu}
          onNodeClick={(_, node) => {
            if (node.type === 'service') {
              setSelectedService(node.data.service)
            } else if (node.type === 'database') {
              setSelectedDatabase(node.data.database)
            } else if (node.type === 'volume') {
              setSelectedVolume(node.data.volume)
            }
          }}
          nodeTypes={nodeTypes}
          fitView
        >
          <Controls />
          <Background />
          <MiniMap />
        </ReactFlow>
        
        {contextMenu && (
          <ContextMenu
            x={contextMenu.x}
            y={contextMenu.y}
            onClose={() => setContextMenu(null)}
            onAddNode={handleAddNode}
          />
        )}
      </div>

      <RepoSelectionModal
        open={repoModalOpen}
        onOpenChange={setRepoModalOpen}
        onSelectRepo={handleRepoSelect}
        projectId={projectId}
      />

      <ServiceDrawer
        service={selectedService}
        isOpen={!!selectedService}
        onClose={() => setSelectedService(null)}
      />

      <DatabaseDrawer
        database={selectedDatabase}
        isOpen={!!selectedDatabase}
        onClose={() => setSelectedDatabase(null)}
      />

      <VolumeDrawer
        volume={selectedVolume}
        isOpen={!!selectedVolume}
        onClose={() => setSelectedVolume(null)}
      />
    </div>
  )
}

