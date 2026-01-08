'use client'

import { useEffect, useMemo, useRef, useState } from 'react'
import { Centrifuge, type Subscription } from 'centrifuge'
import { deploymentsApi, type DeploymentLog } from '@/lib/api/deployments'
import { realtimeApi } from '@/lib/api/realtime'

interface LogStreamProps {
  deploymentId: string
  height?: string
}

export default function LogStream({ deploymentId, height = '400px' }: LogStreamProps) {
  const [logs, setLogs] = useState<DeploymentLog[]>([])
  const [status, setStatus] = useState<'connecting' | 'live' | 'polling' | 'offline'>('connecting')
  const centrifugeRef = useRef<Centrifuge | null>(null)
  const subRef = useRef<Subscription | null>(null)

  const channel = useMemo(() => `deployment:${deploymentId}`, [deploymentId])

  useEffect(() => {
    let cancelled = false

    async function loadInitial() {
      const initial = await deploymentsApi.getLogs(deploymentId, 500, 0)
      if (!cancelled) setLogs(initial)
    }

    loadInitial().catch(() => {
      // ignore
    })

    return () => {
      cancelled = true
    }
  }, [deploymentId])

  useEffect(() => {
    let pollTimer: any = null
    let stopped = false

    async function startPolling() {
      setStatus('polling')
      pollTimer = setInterval(async () => {
        try {
          const latest = await deploymentsApi.getLogs(deploymentId, 500, 0)
          if (!stopped) setLogs(latest)
        } catch {
          // ignore
        }
      }, 2000)
    }

    async function startCentrifugo() {
      try {
        const connect = await realtimeApi.getConnectToken()
        const centrifuge = new Centrifuge(connect.ws_url, {
          token: connect.token,
        })

        centrifuge.on('connecting', () => setStatus('connecting'))
        centrifuge.on('connected', () => setStatus('live'))
        centrifuge.on('disconnected', () => setStatus('offline'))

        centrifugeRef.current = centrifuge

        const subToken = await realtimeApi.getSubscriptionToken(channel)
        const sub = centrifuge.newSubscription(channel, { token: subToken.token })
        subRef.current = sub

        sub.on('publication', (ctx) => {
          // ctx.data is whatever backend published
          const data: any = ctx.data
          // normalize to DeploymentLog shape-ish
          const entry: DeploymentLog = {
            id: Date.now(),
            deployment_id: deploymentId,
            timestamp: data.timestamp || new Date().toISOString(),
            phase: data.phase || 'deploy',
            level: data.level || 'info',
            message: data.message || '',
            metadata: data.metadata,
          }
          setLogs((prev) => [...prev, entry])
        })

        sub.subscribe()
        centrifuge.connect()
      } catch (e) {
        // No centrifugo configured or failed - fallback to polling
        await startPolling()
      }
    }

    startCentrifugo()

    return () => {
      stopped = true
      if (pollTimer) clearInterval(pollTimer)
      try {
        subRef.current?.unsubscribe()
      } catch {}
      try {
        centrifugeRef.current?.disconnect()
      } catch {}
      subRef.current = null
      centrifugeRef.current = null
    }
  }, [deploymentId, channel])

  return (
    <div className="border border-gray-200 rounded-md overflow-hidden">
      <div className="px-3 py-2 border-b bg-gray-50 flex items-center justify-between">
        <div className="text-sm font-medium">Logs</div>
        <div className="text-xs text-gray-500">
          {status === 'live' && 'Live (Centrifugo)'}
          {status === 'polling' && 'Polling'}
          {status === 'connecting' && 'Connecting...'}
          {status === 'offline' && 'Offline'}
        </div>
      </div>
      <div
        className="bg-black text-white font-mono text-xs p-3 overflow-y-auto"
        style={{ height }}
      >
        {logs.length === 0 ? (
          <div className="text-gray-400">No logs yet.</div>
        ) : (
          logs.map((l, idx) => (
            <div key={`${l.id}-${idx}`} className="whitespace-pre-wrap">
              <span className="text-gray-400">{new Date(l.timestamp).toLocaleTimeString()}</span>{' '}
              <span className="text-blue-300">{l.phase}</span>{' '}
              <span
                className={
                  l.level === 'error'
                    ? 'text-red-300'
                    : l.level === 'warn'
                      ? 'text-yellow-300'
                      : 'text-green-300'
                }
              >
                {l.level}
              </span>{' '}
              <span>{l.message}</span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}


