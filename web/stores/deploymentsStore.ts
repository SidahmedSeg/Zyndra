import { create } from 'zustand'
import { deploymentsApi, type Deployment } from '@/lib/api/deployments'

interface DeploymentsState {
  deployments: Record<string, Deployment[]> // serviceId -> deployments
  activeDeployments: Record<string, Deployment> // serviceId -> current deployment
  loading: boolean
  error: string | null

  // Actions
  triggerDeployment: (serviceId: string) => Promise<Deployment>
  cancelDeployment: (deploymentId: string) => Promise<void>
  fetchDeployments: (serviceId: string) => Promise<void>
  fetchDeployment: (deploymentId: string) => Promise<Deployment>
  pollDeploymentStatus: (deploymentId: string, serviceId: string, onUpdate?: (deployment: Deployment) => void) => () => void
  setActiveDeployment: (serviceId: string, deployment: Deployment) => void
  clearError: () => void
}

export const useDeploymentsStore = create<DeploymentsState>((set, get) => ({
  deployments: {},
  activeDeployments: {},
  loading: false,
  error: null,

  triggerDeployment: async (serviceId: string) => {
    set({ loading: true, error: null })
    try {
      const deployment = await deploymentsApi.trigger(serviceId)
      set((state) => ({
        activeDeployments: {
          ...state.activeDeployments,
          [serviceId]: deployment,
        },
        loading: false,
      }))
      return deployment
    } catch (error: any) {
      set({ error: error.message || 'Failed to trigger deployment', loading: false })
      throw error
    }
  },

  cancelDeployment: async (deploymentId: string) => {
    set({ loading: true, error: null })
    try {
      await deploymentsApi.cancel(deploymentId)
      set({ loading: false })
    } catch (error: any) {
      set({ error: error.message || 'Failed to cancel deployment', loading: false })
      throw error
    }
  },

  fetchDeployments: async (serviceId: string) => {
    set({ loading: true, error: null })
    try {
      const deployments = await deploymentsApi.listByService(serviceId)
      set((state) => ({
        deployments: {
          ...state.deployments,
          [serviceId]: Array.isArray(deployments) ? deployments : [],
        },
        loading: false,
      }))
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch deployments', loading: false })
    }
  },

  fetchDeployment: async (deploymentId: string) => {
    try {
      const deployment = await deploymentsApi.get(deploymentId)
      return deployment
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch deployment' })
      throw error
    }
  },

  pollDeploymentStatus: (deploymentId: string, serviceId: string, onUpdate?: (deployment: Deployment) => void) => {
    let intervalId: NodeJS.Timeout | null = null
    let isPolling = true

    const poll = async () => {
      if (!isPolling) return

      try {
        const deployment = await deploymentsApi.get(deploymentId)
        
        set((state) => ({
          activeDeployments: {
            ...state.activeDeployments,
            [serviceId]: deployment,
          },
        }))

        if (onUpdate) {
          onUpdate(deployment)
        }

        // Stop polling if deployment is finished
        if (['success', 'failed', 'cancelled'].includes(deployment.status)) {
          if (intervalId) {
            clearInterval(intervalId)
            intervalId = null
          }
        }
      } catch (error) {
        console.error('Error polling deployment status:', error)
      }
    }

    // Initial poll
    poll()
    
    // Poll every 2 seconds
    intervalId = setInterval(poll, 2000)

    // Return cleanup function
    return () => {
      isPolling = false
      if (intervalId) {
        clearInterval(intervalId)
      }
    }
  },

  setActiveDeployment: (serviceId: string, deployment: Deployment) => {
    set((state) => ({
      activeDeployments: {
        ...state.activeDeployments,
        [serviceId]: deployment,
      },
    }))
  },

  clearError: () => {
    set({ error: null })
  },
}))
