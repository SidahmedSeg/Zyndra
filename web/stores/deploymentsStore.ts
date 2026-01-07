import { create } from 'zustand'
import { deploymentsApi, type Deployment, type DeploymentLog } from '@/lib/api/deployments'

interface DeploymentsState {
  deployments: Record<string, Deployment[]> // keyed by serviceId
  logs: Record<string, DeploymentLog[]> // keyed by deploymentId
  loading: boolean
  error: string | null

  // Actions
  fetchDeployments: (serviceId: string) => Promise<void>
  fetchDeployment: (id: string) => Promise<Deployment>
  fetchLogs: (id: string) => Promise<void>
  triggerDeployment: (serviceId: string, data: any) => Promise<Deployment>
  cancelDeployment: (id: string) => Promise<void>
  clearError: () => void
}

export const useDeploymentsStore = create<DeploymentsState>((set, get) => ({
  deployments: {},
  logs: {},
  loading: false,
  error: null,

  fetchDeployments: async (serviceId: string) => {
    set({ loading: true, error: null })
    try {
      const deployments = await deploymentsApi.listByService(serviceId)
      set((state) => ({
        deployments: { ...state.deployments, [serviceId]: deployments },
        loading: false,
      }))
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch deployments', loading: false })
    }
  },

  fetchDeployment: async (id: string) => {
    set({ loading: true, error: null })
    try {
      const deployment = await deploymentsApi.get(id)
      set({ loading: false })
      return deployment
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch deployment', loading: false })
      throw error
    }
  },

  fetchLogs: async (id: string) => {
    try {
      const logs = await deploymentsApi.getLogs(id)
      set((state) => ({
        logs: { ...state.logs, [id]: logs },
      }))
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch logs' })
    }
  },

  triggerDeployment: async (serviceId: string, data: any) => {
    set({ loading: true, error: null })
    try {
      const deployment = await deploymentsApi.trigger(serviceId, data)
      set((state) => ({
        deployments: {
          ...state.deployments,
          [serviceId]: [deployment, ...(state.deployments[serviceId] || [])],
        },
        loading: false,
      }))
      return deployment
    } catch (error: any) {
      set({ error: error.message || 'Failed to trigger deployment', loading: false })
      throw error
    }
  },

  cancelDeployment: async (id: string) => {
    set({ loading: true, error: null })
    try {
      await deploymentsApi.cancel(id)
      set({ loading: false })
      // Update deployment status in store
      const deployments = get().deployments
      for (const serviceId in deployments) {
        const index = deployments[serviceId].findIndex((d) => d.id === id)
        if (index >= 0) {
          deployments[serviceId][index] = {
            ...deployments[serviceId][index],
            status: 'cancelled',
          }
          set((state) => ({
            deployments: { ...state.deployments, [serviceId]: deployments[serviceId] },
          }))
          break
        }
      }
    } catch (error: any) {
      set({ error: error.message || 'Failed to cancel deployment', loading: false })
      throw error
    }
  },

  clearError: () => {
    set({ error: null })
  },
}))

