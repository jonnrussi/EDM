import { FormEvent, useState } from 'react'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'

import { registerInventoryEndpoint } from '../api/deviceClient'

const sample = [
  { name: 'Compliant', devices: 820 },
  { name: 'Critical CVEs', devices: 73 },
  { name: 'Pending Patch', devices: 166 },
]

export function DashboardPage() {
  const [hostname, setHostname] = useState('')
  const [osName, setOsName] = useState('Windows')
  const [osVersion, setOsVersion] = useState('11')
  const [cpu, setCpu] = useState('x86_64')
  const [ramMB, setRamMB] = useState(8192)
  const [token, setToken] = useState('')
  const [result, setResult] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleRegisterEndpoint(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setLoading(true)
    setResult('')

    try {
      const data = await registerInventoryEndpoint(
        {
          hostname,
          os_name: osName,
          os_version: osVersion,
          cpu,
          ram_mb: ramMB,
        },
        token,
      )
      setResult(`Endpoint adicionado com sucesso. Device ID: ${data.device_id}`)
    } catch {
      setResult('Falha ao adicionar endpoint no inventário. Verifique token e conectividade.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ fontFamily: 'Inter, sans-serif', padding: 24 }}>
      <h1>UEM Enterprise Dashboard (Tenant: Acme Corp)</h1>
      <p>Visão multi-tenant com isolamento lógico por tenant_id e filtros por grupo dinâmico.</p>

      <div style={{ width: '100%', height: 320, marginBottom: 24 }}>
        <ResponsiveContainer>
          <BarChart data={sample}>
            <XAxis dataKey="name" />
            <YAxis />
            <Tooltip />
            <Bar dataKey="devices" fill="#2f80ed" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <h2>Adicionar endpoint ao inventário</h2>
      <form onSubmit={handleRegisterEndpoint} style={{ display: 'grid', gap: 8, maxWidth: 560 }}>
        <input value={token} onChange={(e) => setToken(e.target.value)} placeholder="JWT Bearer token" required />
        <input value={hostname} onChange={(e) => setHostname(e.target.value)} placeholder="Hostname" required />
        <input value={osName} onChange={(e) => setOsName(e.target.value)} placeholder="Sistema operacional" required />
        <input value={osVersion} onChange={(e) => setOsVersion(e.target.value)} placeholder="Versão do SO" required />
        <input value={cpu} onChange={(e) => setCpu(e.target.value)} placeholder="CPU" required />
        <input
          type="number"
          value={ramMB}
          onChange={(e) => setRamMB(Number(e.target.value))}
          placeholder="RAM (MB)"
          required
        />
        <button type="submit" disabled={loading}>
          {loading ? 'Adicionando...' : 'Adicionar endpoint'}
        </button>
      </form>

      {result ? <p style={{ marginTop: 12 }}>{result}</p> : null}
    </div>
  )
}
