'use client'

import { useEffect, useState } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { format, subHours } from 'date-fns'
import { metricsApi, type MetricsResponse } from '@/lib/api/metrics'

interface MetricsTabProps {
  resourceId: string
  resourceType: 'service' | 'database' | 'volume'
}

export default function MetricsTab({ resourceId, resourceType }: MetricsTabProps) {
  const [metrics, setMetrics] = useState<MetricsResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState<'1h' | '6h' | '24h' | '7d'>('1h')

  useEffect(() => {
    const fetchMetrics = async () => {
      setLoading(true)
      try {
        const end = new Date()
        const hours = timeRange === '1h' ? 1 : timeRange === '6h' ? 6 : timeRange === '24h' ? 24 : 168
        const start = subHours(end, hours)
        
        let data: MetricsResponse
        if (resourceType === 'service') {
          data = await metricsApi.getServiceMetrics(
            resourceId,
            start.toISOString(),
            end.toISOString(),
            '30s'
          )
        } else if (resourceType === 'database') {
          data = await metricsApi.getDatabaseMetrics(
            resourceId,
            start.toISOString(),
            end.toISOString(),
            '30s'
          )
        } else {
          data = await metricsApi.getVolumeMetrics(
            resourceId,
            start.toISOString(),
            end.toISOString(),
            '30s'
          )
        }
        setMetrics(data)
      } catch (error) {
        console.error('Failed to fetch metrics:', error)
      } finally {
        setLoading(false)
      }
    }

    if (resourceId) {
      fetchMetrics()
      // Refresh every 30 seconds
      const interval = setInterval(fetchMetrics, 30000)
      return () => clearInterval(interval)
    }
  }, [resourceId, resourceType, timeRange])

  if (loading && !metrics) {
    return (
      <div className="flex items-center justify-center h-96">
        <p className="text-gray-500">Loading metrics...</p>
      </div>
    )
  }

  if (!metrics) {
    return (
      <div className="flex items-center justify-center h-96">
        <p className="text-gray-500">No metrics available</p>
      </div>
    )
  }

  // Transform data for charts
  const transformData = (points: { timestamp: string; value: number }[]) => {
    return points.map((p) => ({
      time: format(new Date(p.timestamp), 'HH:mm:ss'),
      value: p.value,
    }))
  }

  // Merge network data for combined chart
  const mergeNetworkData = () => {
    if (!metrics.network_in && !metrics.network_out) return []
    
    const inData = metrics.network_in ? transformData(metrics.network_in) : []
    const outData = metrics.network_out ? transformData(metrics.network_out) : []
    
    // Merge by timestamp
    const merged: Record<string, { time: string; inbound?: number; outbound?: number }> = {}
    
    inData.forEach((point) => {
      if (!merged[point.time]) {
        merged[point.time] = { time: point.time }
      }
      merged[point.time].inbound = point.value
    })
    
    outData.forEach((point) => {
      if (!merged[point.time]) {
        merged[point.time] = { time: point.time }
      }
      merged[point.time].outbound = point.value
    })
    
    return Object.values(merged)
  }

  const networkData = mergeNetworkData()

  return (
    <div className="space-y-6">
      {/* Time Range Selector */}
      <div className="flex items-center gap-2">
        <label className="text-sm font-medium text-gray-700">Time Range:</label>
        <select
          value={timeRange}
          onChange={(e) => setTimeRange(e.target.value as '1h' | '6h' | '24h' | '7d')}
          className="rounded-md border border-gray-300 px-3 py-1 text-sm"
        >
          <option value="1h">Last Hour</option>
          <option value="6h">Last 6 Hours</option>
          <option value="24h">Last 24 Hours</option>
          <option value="7d">Last 7 Days</option>
        </select>
      </div>

      {/* CPU Usage */}
      {metrics.cpu && metrics.cpu.length > 0 && (
        <div className="border border-gray-200 rounded-lg p-4">
          <h3 className="text-lg font-semibold mb-4">CPU Usage (%)</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={transformData(metrics.cpu)}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line
                type="monotone"
                dataKey="value"
                stroke="#3B82F6"
                strokeWidth={2}
                dot={false}
                name="CPU %"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Memory Usage */}
      {metrics.memory && metrics.memory.length > 0 && (
        <div className="border border-gray-200 rounded-lg p-4">
          <h3 className="text-lg font-semibold mb-4">Memory Usage</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={transformData(metrics.memory)}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip
                formatter={(value: number | undefined) => {
                  if (value === undefined) return 'N/A'
                  if (resourceType === 'volume') {
                    return `${(value / 1024 / 1024 / 1024).toFixed(2)} GB`
                  }
                  return `${(value / 1024 / 1024).toFixed(2)} MB`
                }}
              />
              <Legend />
              <Line
                type="monotone"
                dataKey="value"
                stroke="#10B981"
                strokeWidth={2}
                dot={false}
                name="Memory"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Network Traffic */}
      {networkData.length > 0 && (
        <div className="border border-gray-200 rounded-lg p-4">
          <h3 className="text-lg font-semibold mb-4">
            {resourceType === 'volume' ? 'I/O Traffic' : 'Network Traffic'}
          </h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={networkData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip
                formatter={(value: number | undefined) => {
                  if (value === undefined) return 'N/A'
                  return `${(value / 1024 / 1024).toFixed(2)} MB/s`
                }}
              />
              <Legend />
              {metrics.network_in && metrics.network_in.length > 0 && (
                <Line
                  type="monotone"
                  dataKey="inbound"
                  stroke="#8B5CF6"
                  strokeWidth={2}
                  dot={false}
                  name={resourceType === 'volume' ? 'Read' : 'Inbound'}
                />
              )}
              {metrics.network_out && metrics.network_out.length > 0 && (
                <Line
                  type="monotone"
                  dataKey="outbound"
                  stroke="#F59E0B"
                  strokeWidth={2}
                  dot={false}
                  name={resourceType === 'volume' ? 'Write' : 'Outbound'}
                />
              )}
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Request Count (Services only) */}
      {resourceType === 'service' && metrics.request_count && metrics.request_count.length > 0 && (
        <div className="border border-gray-200 rounded-lg p-4">
          <h3 className="text-lg font-semibold mb-4">Request Rate (req/s)</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={transformData(metrics.request_count)}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line
                type="monotone"
                dataKey="value"
                stroke="#EF4444"
                strokeWidth={2}
                dot={false}
                name="Requests/s"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Response Time (Services only) */}
      {resourceType === 'service' && metrics.response_time && metrics.response_time.length > 0 && (
        <div className="border border-gray-200 rounded-lg p-4">
          <h3 className="text-lg font-semibold mb-4">Response Time (ms)</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={transformData(metrics.response_time)}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip
                formatter={(value: number | undefined) => {
                  if (value === undefined) return 'N/A'
                  return `${(value * 1000).toFixed(2)} ms`
                }}
              />
              <Legend />
              <Line
                type="monotone"
                dataKey="value"
                stroke="#06B6D4"
                strokeWidth={2}
                dot={false}
                name="Response Time"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Error Rate (Services only) */}
      {resourceType === 'service' && metrics.error_rate && metrics.error_rate.length > 0 && (
        <div className="border border-gray-200 rounded-lg p-4">
          <h3 className="text-lg font-semibold mb-4">Error Rate (errors/s)</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={transformData(metrics.error_rate)}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line
                type="monotone"
                dataKey="value"
                stroke="#DC2626"
                strokeWidth={2}
                dot={false}
                name="Errors/s"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  )
}
