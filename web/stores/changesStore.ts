import { create } from 'zustand'

export type ChangeType = 
  | 'root_dir'
  | 'branch'
  | 'env_var_add'
  | 'env_var_update'
  | 'env_var_delete'
  | 'resource_cpu'
  | 'resource_memory'
  | 'custom_domain_add'
  | 'custom_domain_remove'
  | 'port'
  | 'start_command'

export interface Change {
  id: string
  type: ChangeType
  field: string
  oldValue?: string | number
  newValue?: string | number
  description: string
  timestamp: number
}

export interface ServiceChanges {
  serviceId: string
  serviceName: string
  changes: Change[]
  originalConfig: ServiceConfig
  currentConfig: ServiceConfig
}

export interface ServiceConfig {
  rootDir: string
  branch: string
  port: number
  cpu: string
  memory: string
  startCommand: string
  envVars: Record<string, string>
  customDomains: string[]
}

interface ChangesState {
  // Map of serviceId -> ServiceChanges
  serviceChanges: Record<string, ServiceChanges>
  
  // Initialize tracking for a service
  initializeService: (serviceId: string, serviceName: string, config: ServiceConfig) => void
  
  // Add a change
  addChange: (serviceId: string, change: Omit<Change, 'id' | 'timestamp'>) => void
  
  // Remove a change
  removeChange: (serviceId: string, changeId: string) => void
  
  // Clear all changes for a service
  clearChanges: (serviceId: string) => void
  
  // Get changes count for a service
  getChangesCount: (serviceId: string) => number
  
  // Get all changes for a service
  getChanges: (serviceId: string) => Change[]
  
  // Update current config
  updateConfig: (serviceId: string, field: keyof ServiceConfig, value: any) => void
  
  // Check if there are unsaved changes
  hasChanges: (serviceId: string) => boolean
  
  // Get change summary for deploy
  getChangeSummary: (serviceId: string) => { type: ChangeType; count: number }[]
  
  // Reset to original config (discard changes)
  discardChanges: (serviceId: string) => void
}

const generateId = () => Math.random().toString(36).substring(2, 9)

const getChangeDescription = (type: ChangeType, field: string, oldValue?: any, newValue?: any): string => {
  switch (type) {
    case 'root_dir':
      return `Root directory changed to "${newValue}"`
    case 'branch':
      return `Branch changed to "${newValue}"`
    case 'env_var_add':
      return `Added environment variable "${field}"`
    case 'env_var_update':
      return `Updated environment variable "${field}"`
    case 'env_var_delete':
      return `Removed environment variable "${field}"`
    case 'resource_cpu':
      return `CPU changed to ${newValue}`
    case 'resource_memory':
      return `Memory changed to ${newValue}`
    case 'custom_domain_add':
      return `Added custom domain "${newValue}"`
    case 'custom_domain_remove':
      return `Removed custom domain "${oldValue}"`
    case 'port':
      return `Port changed to ${newValue}`
    case 'start_command':
      return `Start command updated`
    default:
      return `Configuration changed`
  }
}

export const useChangesStore = create<ChangesState>((set, get) => ({
  serviceChanges: {},

  initializeService: (serviceId, serviceName, config) => {
    set((state) => ({
      serviceChanges: {
        ...state.serviceChanges,
        [serviceId]: {
          serviceId,
          serviceName,
          changes: [],
          originalConfig: { ...config },
          currentConfig: { ...config },
        },
      },
    }))
  },

  addChange: (serviceId, change) => {
    const id = generateId()
    const timestamp = Date.now()
    const description = getChangeDescription(
      change.type,
      change.field,
      change.oldValue,
      change.newValue
    )
    
    set((state) => {
      const serviceData = state.serviceChanges[serviceId]
      if (!serviceData) return state

      // Check if there's already a change for the same field, replace it
      const existingIndex = serviceData.changes.findIndex(
        (c) => c.type === change.type && c.field === change.field
      )

      let newChanges: Change[]
      if (existingIndex >= 0) {
        // Replace existing change
        newChanges = [...serviceData.changes]
        newChanges[existingIndex] = { ...change, id, timestamp, description }
      } else {
        // Add new change
        newChanges = [...serviceData.changes, { ...change, id, timestamp, description }]
      }

      return {
        serviceChanges: {
          ...state.serviceChanges,
          [serviceId]: {
            ...serviceData,
            changes: newChanges,
          },
        },
      }
    })
  },

  removeChange: (serviceId, changeId) => {
    set((state) => {
      const serviceData = state.serviceChanges[serviceId]
      if (!serviceData) return state

      return {
        serviceChanges: {
          ...state.serviceChanges,
          [serviceId]: {
            ...serviceData,
            changes: serviceData.changes.filter((c) => c.id !== changeId),
          },
        },
      }
    })
  },

  clearChanges: (serviceId) => {
    set((state) => {
      const serviceData = state.serviceChanges[serviceId]
      if (!serviceData) return state

      return {
        serviceChanges: {
          ...state.serviceChanges,
          [serviceId]: {
            ...serviceData,
            changes: [],
            originalConfig: { ...serviceData.currentConfig },
          },
        },
      }
    })
  },

  getChangesCount: (serviceId) => {
    const serviceData = get().serviceChanges[serviceId]
    return serviceData?.changes.length || 0
  },

  getChanges: (serviceId) => {
    const serviceData = get().serviceChanges[serviceId]
    return serviceData?.changes || []
  },

  updateConfig: (serviceId, field, value) => {
    set((state) => {
      const serviceData = state.serviceChanges[serviceId]
      if (!serviceData) return state

      return {
        serviceChanges: {
          ...state.serviceChanges,
          [serviceId]: {
            ...serviceData,
            currentConfig: {
              ...serviceData.currentConfig,
              [field]: value,
            },
          },
        },
      }
    })
  },

  hasChanges: (serviceId) => {
    const serviceData = get().serviceChanges[serviceId]
    return (serviceData?.changes.length || 0) > 0
  },

  getChangeSummary: (serviceId) => {
    const changes = get().getChanges(serviceId)
    const summary: Record<ChangeType, number> = {} as Record<ChangeType, number>
    
    changes.forEach((change) => {
      summary[change.type] = (summary[change.type] || 0) + 1
    })

    return Object.entries(summary).map(([type, count]) => ({
      type: type as ChangeType,
      count,
    }))
  },

  discardChanges: (serviceId) => {
    set((state) => {
      const serviceData = state.serviceChanges[serviceId]
      if (!serviceData) return state

      return {
        serviceChanges: {
          ...state.serviceChanges,
          [serviceId]: {
            ...serviceData,
            changes: [],
            currentConfig: { ...serviceData.originalConfig },
          },
        },
      }
    })
  },
}))

