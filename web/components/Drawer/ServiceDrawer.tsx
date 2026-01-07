'use client'

import { useEffect, useMemo, useState } from 'react'
import Drawer from './Drawer'
import Tabs from './Tabs'
import { Service } from '@/lib/api/services'
import { useServicesStore } from '@/stores/servicesStore'
import { useDeploymentsStore } from '@/stores/deploymentsStore'
import type { Deployment } from '@/lib/api/deployments'
import LogStream from '@/components/Logs/LogStream'
import { rollbackApi, type RollbackCandidate } from '@/lib/api/rollback'
import MetricsTab from '@/components/Metrics/MetricsTab'

interface ServiceDrawerProps {
  service: Service | null
  isOpen: boolean
  onClose: () => void
}

export default function ServiceDrawer({
  service,
  isOpen,
  onClose,
}: ServiceDrawerProps) {
  const { updateService } = useServicesStore()
  const { deployments, fetchDeployments, triggerDeployment, cancelDeployment } = useDeploymentsStore()
  const [selectedDeploymentId, setSelectedDeploymentId] = useState<string | null>(null)
  const serviceId = service?.id || null
  const serviceDeployments = serviceId ? deployments[serviceId] || [] : []
  const latestDeployment = serviceDeployments[0] || null

  useEffect(() => {
    if (isOpen && serviceId) {
      fetchDeployments(serviceId)
    }
  }, [isOpen, serviceId, fetchDeployments])

  useEffect(() => {
    if (!selectedDeploymentId && latestDeployment) {
      setSelectedDeploymentId(latestDeployment.id)
    }
  }, [latestDeployment, selectedDeploymentId])

  if (!service) return null

  const tabs = [
    {
      id: 'source',
      label: 'Source',
      content: <SourceTab service={service} />,
    },
    {
      id: 'instance',
      label: 'Instance',
      content: <InstanceTab service={service} onUpdate={updateService} />,
    },
    {
      id: 'variables',
      label: 'Variables',
      content: <VariablesTab service={service} />,
    },
    {
      id: 'domains',
      label: 'Domains',
      content: <DomainsTab service={service} />,
    },
    {
      id: 'deploy',
      label: 'Deploy',
      content: (
        <DeployTab
          service={service}
          deployments={serviceDeployments}
          selectedDeploymentId={selectedDeploymentId}
          onSelectDeployment={setSelectedDeploymentId}
          onTrigger={async () => {
            const d = await triggerDeployment(service.id, { triggered_by: 'manual' })
            setSelectedDeploymentId(d.id)
            await fetchDeployments(service.id)
          }}
          onCancel={async (deploymentId) => {
            await cancelDeployment(deploymentId)
            await fetchDeployments(service.id)
          }}
        />
      ),
    },
    {
      id: 'logs',
      label: 'Logs',
      content: <LogsTab deploymentId={selectedDeploymentId} />,
    },
    {
      id: 'metrics',
      label: 'Metrics',
      content: <MetricsTab resourceId={service.id} resourceType="service" />,
    },
  ]

  return (
    <Drawer
      isOpen={isOpen}
      onClose={onClose}
      title={service.name}
      width="900px"
    >
      <Tabs tabs={tabs} />
    </Drawer>
  )
}

// Tab Components
function SourceTab({ service }: { service: Service }) {
  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Git Source
        </label>
        <p className="text-sm text-gray-500">
          {service.git_source_id ? 'Connected' : 'Not connected'}
        </p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Service Type
        </label>
        <p className="text-sm text-gray-500 capitalize">{service.type}</p>
      </div>
    </div>
  )
}

function InstanceTab({
  service,
  onUpdate,
}: {
  service: Service
  onUpdate: (id: string, data: any) => Promise<void>
}) {
  const [instanceSize, setInstanceSize] = useState(service.instance_size)
  const [port, setPort] = useState(service.port?.toString() || '')

  const handleSave = async () => {
    await onUpdate(service.id, {
      instance_size: instanceSize,
      port: port ? parseInt(port) : undefined,
    })
  }

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Instance Size
        </label>
        <select
          value={instanceSize}
          onChange={(e) => setInstanceSize(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md"
        >
          <option value="small">Small</option>
          <option value="medium">Medium</option>
          <option value="large">Large</option>
        </select>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Port
        </label>
        <input
          type="number"
          value={port}
          onChange={(e) => setPort(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md"
          placeholder="8080"
        />
      </div>
      <button
        onClick={handleSave}
        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
      >
        Save Changes
      </button>
    </div>
  )
}

function VariablesTab({ service }: { service: Service }) {
  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-500">Environment variables will be listed here</p>
    </div>
  )
}

function DomainsTab({ service }: { service: Service }) {
  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Generated URL
        </label>
        <p className="text-sm text-blue-600">{service.generated_url || 'Not available'}</p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Custom Domain
        </label>
        <input
          type="text"
          className="w-full px-3 py-2 border border-gray-300 rounded-md"
          placeholder="example.com"
        />
      </div>
    </div>
  )
}

function DeployTab({
  service,
  deployments,
  selectedDeploymentId,
  onSelectDeployment,
  onTrigger,
  onCancel,
}: {
  service: Service
  deployments: Deployment[]
  selectedDeploymentId: string | null
  onSelectDeployment: (id: string) => void
  onTrigger: () => Promise<void>
  onCancel: (deploymentId: string) => Promise<void>
}) {
  const [rollbackCandidates, setRollbackCandidates] = useState<RollbackCandidate[]>([])
  const [loadingRollback, setLoadingRollback] = useState(false)
  const [rollingBack, setRollingBack] = useState<string | null>(null)

  useEffect(() => {
    const fetchCandidates = async () => {
      try {
        const candidates = await rollbackApi.getRollbackCandidates(service.id)
        setRollbackCandidates(candidates || [])
      } catch (error) {
        console.error('Failed to fetch rollback candidates:', error)
      }
    }
    fetchCandidates()
  }, [service.id, deployments])

  const selected = useMemo(
    () => deployments.find((d) => d.id === selectedDeploymentId) || null,
    [deployments, selectedDeploymentId]
  )

  const handleRollback = async (deploymentId: string) => {
    setRollingBack(deploymentId)
    try {
      await rollbackApi.rollbackToDeployment(service.id, deploymentId)
      // Refresh deployments after rollback
      await onTrigger()
    } catch (error) {
      console.error('Failed to rollback:', error)
      alert('Failed to rollback deployment')
    } finally {
      setRollingBack(null)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <button
          onClick={onTrigger}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Trigger Deployment
        </button>
        {selected && (selected.status === 'queued' || selected.status === 'building' || selected.status === 'pushing') && (
          <button
            onClick={() => onCancel(selected.id)}
            className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700"
          >
            Cancel
          </button>
        )}
      </div>

      {selected && <DeploymentProgress deployment={selected} />}

      {rollbackCandidates.length > 0 && (
        <div className="border border-gray-200 rounded-md p-4 bg-yellow-50">
          <div className="text-sm font-medium mb-2">Rollback</div>
          <div className="text-xs text-gray-600 mb-3">
            Rollback to a previous successful deployment:
          </div>
          <div className="space-y-2 max-h-48 overflow-y-auto">
            {rollbackCandidates.slice(0, 5).map((candidate) => (
              <div
                key={candidate.id}
                className="flex items-center justify-between p-2 bg-white rounded border border-gray-200"
              >
                <div className="flex-1 min-w-0">
                  <div className="text-xs font-medium truncate">
                    {candidate.commit_message || 'No commit message'}
                  </div>
                  <div className="text-xs text-gray-500">
                    {candidate.image_tag} • {new Date(candidate.created_at).toLocaleString()}
                  </div>
                </div>
                <button
                  onClick={() => handleRollback(candidate.id)}
                  disabled={rollingBack === candidate.id}
                  className="ml-2 px-3 py-1 text-xs bg-orange-600 text-white rounded hover:bg-orange-700 disabled:opacity-50"
                >
                  {rollingBack === candidate.id ? 'Rolling back...' : 'Rollback'}
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      <div>
        <div className="text-sm font-medium mb-2">History</div>
        {deployments.length === 0 ? (
          <div className="text-sm text-gray-500">No deployments yet.</div>
        ) : (
          <div className="border border-gray-200 rounded-md divide-y">
            {deployments.map((d) => (
              <button
                key={d.id}
                onClick={() => onSelectDeployment(d.id)}
                className={`w-full text-left px-3 py-2 hover:bg-gray-50 ${
                  d.id === selectedDeploymentId ? 'bg-blue-50' : ''
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="text-sm font-medium">{d.status}</div>
                  <div className="text-xs text-gray-500">
                    {new Date(d.created_at).toLocaleString()}
                  </div>
                </div>
                {d.image_tag && (
                  <div className="text-xs text-gray-600 truncate">{d.image_tag}</div>
                )}
                {d.error_message && (
                  <div className="text-xs text-red-600 truncate">{d.error_message}</div>
                )}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function DeploymentProgress({ deployment }: { deployment: Deployment }) {
  const steps = [
    { id: 'queued', label: 'Queued' },
    { id: 'building', label: 'Build' },
    { id: 'pushing', label: 'Push' },
    { id: 'deploying', label: 'Deploy' },
    { id: 'success', label: 'Done' },
  ]

  const statusIndex = steps.findIndex((s) => s.id === deployment.status)
  const isFailed = deployment.status === 'failed'
  const isCancelled = deployment.status === 'cancelled'

  return (
    <div className="border border-gray-200 rounded-md p-3">
      <div className="text-sm font-medium mb-2">Progress</div>
      <div className="flex items-center gap-2 flex-wrap">
        {steps.map((s, idx) => {
          const done = statusIndex >= idx && statusIndex !== -1
          return (
            <div key={s.id} className="flex items-center gap-2">
              <div
                className={`px-2 py-1 rounded text-xs ${
                  done ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                }`}
              >
                {s.label}
              </div>
              {idx !== steps.length - 1 && <div className="text-gray-300">→</div>}
            </div>
          )
        })}
      </div>
      {(isFailed || isCancelled) && (
        <div className="mt-2 text-sm text-red-600">
          {isCancelled ? 'Cancelled' : `Failed: ${deployment.error_message || 'Unknown error'}`}
        </div>
      )}
    </div>
  )
}

function LogsTab({ deploymentId }: { deploymentId: string | null }) {
  return (
    <div className="space-y-4">
      {!deploymentId ? (
        <p className="text-sm text-gray-500">Select a deployment to view logs.</p>
      ) : (
        <LogStream deploymentId={deploymentId} height="520px" />
      )}
    </div>
  )
}

