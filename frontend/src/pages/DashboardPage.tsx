import { FormEvent, useState } from 'react'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'

import { deleteManagedDevice, listManagedDevices, registerInventoryEndpoint, type ManagedDevice } from '../api/deviceClient'

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
  const [devices, setDevices] = useState<ManagedDevice[]>([])

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
      await handleLoadDevices()
    } catch {
      setResult('Falha ao adicionar endpoint no inventário. Verifique token e conectividade.')
    } finally {
      setLoading(false)
    }
  }

  async function handleLoadDevices() {
    try {
      const list = await listManagedDevices(token)
      setDevices(list)
    } catch {
      setResult('Não foi possível carregar endpoints. Verifique o token JWT.')
    }
  }

  async function handleDeleteDevice(deviceId: string) {
    try {
      await deleteManagedDevice(deviceId, token)
      setResult(`Endpoint removido: ${deviceId}`)
      await handleLoadDevices()
    } catch {
      setResult('Falha ao remover endpoint.')
    }
  }

  return (
    <div style={{ fontFamily: 'Inter, sans-serif', padding: 24 }}>
      <h1>UEM Enterprise Dashboard (Tenant: Acme Corp)</h1>
      <p>Sistema de gestão de endpoints com cadastro, listagem e remoção de inventário.</p>

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

      <div style={{ marginTop: 16 }}>
        <button onClick={handleLoadDevices}>Carregar endpoints do inventário</button>
      </div>

      <h2 style={{ marginTop: 20 }}>Endpoints gerenciados</h2>
      <table cellPadding={8} style={{ borderCollapse: 'collapse', width: '100%', maxWidth: 960 }}>
        <thead>
          <tr>
            <th align="left">Hostname</th>
            <th align="left">SO</th>
            <th align="left">CPU</th>
            <th align="left">RAM</th>
            <th align="left">Antivírus</th>
            <th align="left">Criptografia</th>
            <th align="left">Ações</th>
          </tr>
        </thead>
        <tbody>
          {devices.map((device) => (
            <tr key={device.id}>
              <td>{device.hostname}</td>
              <td>
                {device.os} {device.os_version}
              </td>
              <td>{device.cpu}</td>
              <td>{device.ram_mb} MB</td>
              <td>{device.antivirus_status}</td>
              <td>{device.encryption_status}</td>
              <td>
                <button onClick={() => handleDeleteDevice(device.id)}>Remover</button>
              </td>
            </tr>
          ))}
          {devices.length === 0 ? (
            <tr>
              <td colSpan={7}>Nenhum endpoint carregado.</td>
            </tr>
          ) : null}
        </tbody>
      </table>

      {result ? <p style={{ marginTop: 12 }}>{result}</p> : null}
    </div>
  )
}
