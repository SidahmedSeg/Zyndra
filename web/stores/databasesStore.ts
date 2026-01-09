import { create } from 'zustand'
import { databasesApi, type Database } from '@/lib/api/databases'

interface DatabasesState {
  databases: Database[]
  selectedDatabase: Database | null
  loading: boolean
  error: string | null

  fetchDatabases: (projectId: string) => Promise<void>
  setSelectedDatabase: (db: Database | null) => void
  clearError: () => void
}

export const useDatabasesStore = create<DatabasesState>((set) => ({
  databases: [],
  selectedDatabase: null,
  loading: false,
  error: null,

  fetchDatabases: async (projectId: string) => {
    set({ loading: true, error: null })
    try {
      const databases = await databasesApi.listByProject(projectId)
      set({ databases: Array.isArray(databases) ? databases : [], loading: false })
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch databases', loading: false, databases: [] })
    }
  },

  setSelectedDatabase: (db: Database | null) => set({ selectedDatabase: db }),

  clearError: () => set({ error: null }),
}))


