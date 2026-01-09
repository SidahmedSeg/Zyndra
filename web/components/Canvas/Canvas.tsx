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
  Panel,
  ReactFlowInstance,
} from 'reactflow'
import 'reactflow/dist/style.css'

import { useCanvasStore } from '@/stores/canvasStore'
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

const nodeTypes = {
  service: ServiceNode,
  database: DatabaseNode,
  volume: VolumeNode,
}

interface CanvasProps {
  projectId: string
}

export default function Canvas({ projectId }: CanvasProps) {
  const {
    nodes: storeNodes,
    edges: storeEdges,
    onNodesChange: storeOnNodesChange,
    onEdgesChange: storeOnEdgesChange,
    onConnect: storeOnConnect,
    setNodes: storeSetNodes,
    setEdges: storeSetEdges,
  } = useCanvasStore()

  const { services, fetchServices, selectedService, setSelectedService, createService } = useServicesStore()
  const { databases, fetchDatabases, selectedDatabase, setSelectedDatabase, createDatabase } = useDatabasesStore()
  const { volumes, fetchVolumes, selectedVolume, setSelectedVolume, createVolume } = useVolumesStore()
  
  // Ensure arrays are always arrays (handle null/undefined)
  const servicesList = useMemo(() => Array.isArray(services) ? services : [], [services])
  const databasesList = useMemo(() => Array.isArray(databases) ? databases : [], [databases])
  const volumesList = useMemo(() => Array.isArray(volumes) ? volumes : [], [volumes])

  const [nodes, setNodes, onNodesChange] = useNodesState(storeNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(storeEdges)
  
  // Context menu state
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number } | null>(null)
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null)

  // Sync with store
  useEffect(() => {
    setNodes(storeNodes)
  }, [storeNodes, setNodes])

  useEffect(() => {
    setEdges(storeEdges)
  }, [storeEdges, setEdges])

  // Update store when nodes/edges change
  useEffect(() => {
    storeSetNodes(nodes)
  }, [nodes, storeSetNodes])

  useEffect(() => {
    storeSetEdges(edges)
  }, [edges, storeSetEdges])

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
      }))

      // Only add nodes that don't already exist
      setNodes((currentNodes) => {
        const existingIds = new Set(currentNodes.map((n) => n.id))
        const newNodes = serviceNodes.filter((n) => !existingIds.has(n.id))
        return newNodes.length > 0 ? [...currentNodes, ...newNodes] : currentNodes
      })
    }
  }, [servicesList, setNodes])

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
      const newEdges = addEdge(params, edges)
      setEdges(newEdges)
      storeOnConnect(params)
    },
    [edges, setEdges, storeOnConnect]
  )

  const handleNodesChange = useCallback(
    (changes: any) => {
      onNodesChange(changes)
      storeOnNodesChange(changes)
    },
    [onNodesChange, storeOnNodesChange]
  )

  const handleEdgesChange = useCallback(
    (changes: any) => {
      onEdgesChange(changes)
      storeOnEdgesChange(changes)
    },
    [onEdgesChange, storeOnEdgesChange]
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
    async (type: 'service' | 'database' | 'volume') => {
      if (!contextMenu || !reactFlowInstance) return

      const position = reactFlowInstance.screenToFlowPosition({
        x: contextMenu.x,
        y: contextMenu.y,
      })

      try {
        if (type === 'service') {
          const service = await createService(projectId, {
            name: 'New Service',
            type: 'app',
            instance_size: 'small',
            port: 8080,
            canvas_x: position.x,
            canvas_y: position.y,
          })
          setSelectedService(service)
        } else if (type === 'database') {
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
    [contextMenu, reactFlowInstance, projectId, createService, createDatabase, createVolume, setSelectedService, setSelectedDatabase, setSelectedVolume]
  )

  return (
    <div className="w-full h-screen flex flex-col">
      <CanvasHeader projectId={projectId} />
      <div className="flex-1 relative">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={handleNodesChange}
          onEdgesChange={handleEdgesChange}
          onConnect={onConnect}
          onInit={setReactFlowInstance}
          onPaneContextMenu={onPaneContextMenu}
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

