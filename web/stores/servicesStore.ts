import { create } from 'zustand'
import { servicesApi, type Service } from '@/lib/api/services'

interface ServicesState {
  services: Service[]
  selectedService: Service | null
  loading: boolean
  error: string | null

  // Actions
  fetchServices: (projectId: string) => Promise<void>
  fetchService: (id: string) => Promise<void>
  createService: (projectId: string, data: any) => Promise<Service>
  updateService: (id: string, data: any) => Promise<void>
  updateServicePosition: (id: string, x: number, y: number) => Promise<void>
  deleteService: (id: string) => Promise<void>
  setSelectedService: (service: Service | null) => void
  clearError: () => void
}

export const useServicesStore = create<ServicesState>((set, get) => ({
  services: [],
  selectedService: null,
  loading: false,
  error: null,

  fetchServices: async (projectId: string) => {
    set({ loading: true, error: null })
    try {
      const services = await servicesApi.listByProject(projectId)
      set({ services: Array.isArray(services) ? services : [], loading: false })
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch services', loading: false, services: [] })
    }
  },

  fetchService: async (id: string) => {
    set({ loading: true, error: null })
    try {
      const service = await servicesApi.get(id)
      set({ loading: false })
      // Update in services array if exists
      const services = get().services
      const index = services.findIndex((s) => s.id === id)
      if (index >= 0) {
        services[index] = service
        set({ services: [...services] })
      }
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch service', loading: false })
    }
  },

  createService: async (projectId: string, data: any) => {
    set({ loading: true, error: null })
    try {
      const service = await servicesApi.create(projectId, data)
      set((state) => ({
        services: [...state.services, service],
        loading: false,
      }))
      return service
    } catch (error: any) {
      set({ error: error.message || 'Failed to create service', loading: false })
      throw error
    }
  },

  updateService: async (id: string, data: any) => {
    set({ loading: true, error: null })
    try {
      await servicesApi.update(id, data)
      set((state) => ({
        services: state.services.map((s) => (s.id === id ? { ...s, ...data } : s)),
        selectedService:
          state.selectedService?.id === id
            ? { ...state.selectedService, ...data }
            : state.selectedService,
        loading: false,
      }))
    } catch (error: any) {
      set({ error: error.message || 'Failed to update service', loading: false })
      throw error
    }
  },

  updateServicePosition: async (id: string, x: number, y: number) => {
    try {
      await servicesApi.updatePosition(id, { canvas_x: x, canvas_y: y })
      set((state) => ({
        services: state.services.map((s) =>
          s.id === id ? { ...s, canvas_x: x, canvas_y: y } : s
        ),
        selectedService:
          state.selectedService?.id === id
            ? { ...state.selectedService, canvas_x: x, canvas_y: y }
            : state.selectedService,
      }))
    } catch (error: any) {
      set({ error: error.message || 'Failed to update service position' })
    }
  },

  deleteService: async (id: string) => {
    set({ loading: true, error: null })
    try {
      await servicesApi.delete(id)
      set((state) => ({
        services: state.services.filter((s) => s.id !== id),
        selectedService: state.selectedService?.id === id ? null : state.selectedService,
        loading: false,
      }))
    } catch (error: any) {
      set({ error: error.message || 'Failed to delete service', loading: false })
      throw error
    }
  },

  setSelectedService: (service: Service | null) => {
    set({ selectedService: service })
  },

  clearError: () => {
    set({ error: null })
  },
}))

