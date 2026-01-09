import { create } from 'zustand'
import { volumesApi, type Volume, type CreateVolumeRequest } from '@/lib/api/volumes'

interface VolumesState {
  volumes: Volume[]
  selectedVolume: Volume | null
  loading: boolean
  error: string | null

  fetchVolumes: (projectId: string) => Promise<void>
  createVolume: (projectId: string, data: CreateVolumeRequest) => Promise<Volume>
  setSelectedVolume: (v: Volume | null) => void
  clearError: () => void
}

export const useVolumesStore = create<VolumesState>((set) => ({
  volumes: [],
  selectedVolume: null,
  loading: false,
  error: null,

  fetchVolumes: async (projectId: string) => {
    set({ loading: true, error: null })
    try {
      const volumes = await volumesApi.listByProject(projectId)
      set({ volumes: Array.isArray(volumes) ? volumes : [], loading: false })
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch volumes', loading: false, volumes: [] })
    }
  },

  createVolume: async (projectId: string, data: CreateVolumeRequest) => {
    set({ loading: true, error: null })
    try {
      const volume = await volumesApi.create(projectId, data)
      set((state) => ({
        volumes: [...state.volumes, volume],
        loading: false,
      }))
      return volume
    } catch (error: any) {
      set({ error: error.message || 'Failed to create volume', loading: false })
      throw error
    }
  },

  setSelectedVolume: (v: Volume | null) => set({ selectedVolume: v }),

  clearError: () => set({ error: null }),
}))


