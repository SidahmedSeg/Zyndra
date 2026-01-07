import { apiClient } from './client'

export interface ConnectTokenResponse {
  token: string
  ws_url: string
}

export interface SubscriptionTokenResponse {
  token: string
}

export const realtimeApi = {
  getConnectToken: () => apiClient.get<ConnectTokenResponse>('/realtime/connect-token'),

  getSubscriptionToken: (channel: string) =>
    apiClient.get<SubscriptionTokenResponse>(
      `/realtime/subscription-token?channel=${encodeURIComponent(channel)}`
    ),
}


