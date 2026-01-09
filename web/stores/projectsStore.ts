import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { projectsApi, type Project } from '@/lib/api/projects'

interface ProjectsState {
  projects: Project[]
  selectedProject: Project | null
  loading: boolean
  error: string | null

  // Actions
  fetchProjects: () => Promise<void>
  fetchProject: (id: string) => Promise<void>
  createProject: (data: { name: string; description?: string }) => Promise<Project>
  updateProject: (id: string, data: { name?: string; description?: string }) => Promise<void>
  deleteProject: (id: string) => Promise<void>
  setSelectedProject: (project: Project | null) => void
  clearError: () => void
}

export const useProjectsStore = create<ProjectsState>()(
  persist(
    (set, get) => ({
      projects: [] as Project[],
      selectedProject: null,
      loading: false,
      error: null,

      fetchProjects: async () => {
        set({ loading: true, error: null })
        try {
          const projects = await projectsApi.list()
          set({ projects: Array.isArray(projects) ? projects : [], loading: false })
        } catch (error: any) {
          set({ error: error.message || 'Failed to fetch projects', loading: false, projects: [] })
        }
      },

      fetchProject: async (id: string) => {
        set({ loading: true, error: null })
        try {
          const project = await projectsApi.get(id)
          set({ loading: false })
          // Update in projects array if exists
          const projects = get().projects
          const index = projects.findIndex((p) => p.id === id)
          if (index >= 0) {
            projects[index] = project
            set({ projects: [...projects] })
          }
        } catch (error: any) {
          set({ error: error.message || 'Failed to fetch project', loading: false })
        }
      },

      createProject: async (data) => {
        set({ loading: true, error: null })
        try {
          const project = await projectsApi.create(data)
          set((state) => ({
            projects: [...state.projects, project],
            loading: false,
          }))
          return project
        } catch (error: any) {
          set({ error: error.message || 'Failed to create project', loading: false })
          throw error
        }
      },

      updateProject: async (id: string, data) => {
        set({ loading: true, error: null })
        try {
          await projectsApi.update(id, data)
          set((state) => ({
            projects: state.projects.map((p) => (p.id === id ? { ...p, ...data } : p)),
            selectedProject:
              state.selectedProject?.id === id
                ? { ...state.selectedProject, ...data }
                : state.selectedProject,
            loading: false,
          }))
        } catch (error: any) {
          set({ error: error.message || 'Failed to update project', loading: false })
          throw error
        }
      },

      deleteProject: async (id: string) => {
        set({ loading: true, error: null })
        try {
          await projectsApi.delete(id)
          set((state) => ({
            projects: state.projects.filter((p) => p.id !== id),
            selectedProject:
              state.selectedProject?.id === id ? null : state.selectedProject,
            loading: false,
          }))
        } catch (error: any) {
          set({ error: error.message || 'Failed to delete project', loading: false })
          throw error
        }
      },

      setSelectedProject: (project: Project | null) => {
        set({ selectedProject: project })
      },

      clearError: () => {
        set({ error: null })
      },
    }),
    {
      name: 'projects-storage',
      partialize: (state) => ({ selectedProject: state.selectedProject }),
    }
  )
)

