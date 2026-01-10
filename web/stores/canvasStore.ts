import { create } from 'zustand'
import { Node, Edge, Connection, addEdge, applyNodeChanges, applyEdgeChanges } from 'reactflow'

interface CanvasState {
  nodes: Node[]
  edges: Edge[]
  selectedNodeId: string | null

  // Actions
  setNodes: (nodes: Node[]) => void
  setEdges: (edges: Edge[]) => void
  onNodesChange: (changes: any) => void
  onEdgesChange: (changes: any) => void
  onConnect: (connection: Connection) => void
  addNode: (node: Node) => void
  updateNode: (id: string, data: Partial<Node>) => void
  deleteNode: (id: string) => void
  setSelectedNodeId: (id: string | null) => void
  clearCanvas: () => void
}

const initialNodes: Node[] = []
const initialEdges: Edge[] = []

export const useCanvasStore = create<CanvasState>()((set, get) => ({
  nodes: initialNodes,
  edges: initialEdges,
  selectedNodeId: null,

  setNodes: (nodes: Node[]) => {
    set({ nodes })
  },

  setEdges: (edges: Edge[]) => {
    set({ edges })
  },

  onNodesChange: (changes: any) => {
    set({
      nodes: applyNodeChanges(changes, get().nodes),
    })
  },

  onEdgesChange: (changes: any) => {
    set({
      edges: applyEdgeChanges(changes, get().edges),
    })
  },

  onConnect: (connection: Connection) => {
    set({
      edges: addEdge(connection, get().edges),
    })
  },

  addNode: (node: Node) => {
    set((state) => ({
      nodes: [...state.nodes, node],
    }))
  },

  updateNode: (id: string, data: Partial<Node>) => {
    set((state) => ({
      nodes: state.nodes.map((node) =>
        node.id === id ? { ...node, ...data } : node
      ),
    }))
  },

  deleteNode: (id: string) => {
    set((state) => ({
      nodes: state.nodes.filter((node) => node.id !== id),
      edges: state.edges.filter(
        (edge) => edge.source !== id && edge.target !== id
      ),
      selectedNodeId: state.selectedNodeId === id ? null : state.selectedNodeId,
    }))
  },

  setSelectedNodeId: (id: string | null) => {
    set({ selectedNodeId: id })
  },

  clearCanvas: () => {
    set({ nodes: [], edges: [], selectedNodeId: null })
  },
}))

