import { create } from 'zustand'
import { databasesApi, type Database, type CreateDatabaseRequest } from '@/lib/api/databases'

interface DatabasesState {
  databases: Database[]
  selectedDatabase: Database | null
  loading: boolean
  error: string | null

  fetchDatabases: (projectId: string) => Promise<void>
  createDatabase: (projectId: string, data: CreateDatabaseRequest) => Promise<Database>
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

  createDatabase: async (projectId: string, data: CreateDatabaseRequest) => {
    set({ loading: true, error: null })
    try {
      const database = await databasesApi.create(projectId, data)
      set((state) => ({
        databases: [...state.databases, database],
        loading: false,
      }))
      return database
    } catch (error: any) {
      set({ error: error.message || 'Failed to create database', loading: false })
      throw error
    }
  },

  setSelectedDatabase: (db: Database | null) => set({ selectedDatabase: db }),

  clearError: () => set({ error: null }),
}))


