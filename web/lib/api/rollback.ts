import { apiClient } from './client'
import { Deployment } from './deployments'

export interface RollbackCandidate extends Deployment {
  // Same as Deployment, but filtered to only successful ones
}

export const rollbackApi = {
  // Get rollback candidates (successful deployments)
  getRollbackCandidates: (serviceId: string) =>
    apiClient.get<RollbackCandidate[]>(
      `/services/${serviceId}/rollback-candidates`
    ),

  // Rollback to a specific deployment
  rollbackToDeployment: (serviceId: string, deploymentId: string) =>
    apiClient.post<Deployment>(
      `/services/${serviceId}/rollback/${deploymentId}`
    ),
}

