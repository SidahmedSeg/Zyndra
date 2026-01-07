'use client'

import { useCallback, useEffect } from 'react'
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

  const { services, fetchServices, selectedService, setSelectedService } = useServicesStore()
  const { databases, fetchDatabases, selectedDatabase, setSelectedDatabase } = useDatabasesStore()
  const { volumes, fetchVolumes, selectedVolume, setSelectedVolume } = useVolumesStore()

  const [nodes, setNodes, onNodesChange] = useNodesState(storeNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(storeEdges)

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
    if (services.length > 0) {
      const serviceNodes: Node[] = services.map((service) => ({
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
  }, [services, setNodes])

  // Convert databases to nodes
  useEffect(() => {
    if (databases.length > 0) {
      const dbNodes: Node[] = databases.map((db) => ({
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
        for (const db of databases) {
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
  }, [databases, setNodes, setEdges])

  // Convert volumes to nodes
  useEffect(() => {
    if (volumes.length > 0) {
      const volNodes: Node[] = volumes.map((v) => ({
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
        for (const v of volumes) {
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
  }, [volumes, setNodes, setEdges])

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

  return (
    <div style={{ width: '100%', height: '100vh' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={handleNodesChange}
        onEdgesChange={handleEdgesChange}
        onConnect={onConnect}
        nodeTypes={nodeTypes}
        fitView
      >
        <Controls />
        <Background />
        <MiniMap />
        <Panel position="top-left" className="bg-white p-2 rounded shadow">
          <h2 className="text-lg font-semibold">Canvas</h2>
          <p className="text-sm text-gray-600">Drag nodes to reposition</p>
        </Panel>
      </ReactFlow>

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

