import axios from 'axios'

export type RegisterInventoryEndpointRequest = {
  hostname: string
  os_name: string
  os_version: string
  cpu: string
  ram_mb: number
}

export type RegisterInventoryEndpointResponse = {
  device_id: string
}

export type ManagedDevice = {
  id: string
  hostname: string
  os: string
  os_version: string
  cpu: string
  ram_mb: number
  antivirus_status: string
  encryption_status: string
  last_seen: string
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

function authHeaders(token: string) {
  return {
    Authorization: `Bearer ${token}`,
  }
}

export async function registerInventoryEndpoint(
  payload: RegisterInventoryEndpointRequest,
  token: string,
): Promise<RegisterInventoryEndpointResponse> {
  const response = await axios.post<RegisterInventoryEndpointResponse>(`${API_BASE_URL}/devices/v1/devices`, payload, {
    headers: authHeaders(token),
  })

  return response.data
}

export async function listManagedDevices(token: string): Promise<ManagedDevice[]> {
  const response = await axios.get<ManagedDevice[]>(`${API_BASE_URL}/devices/v1/devices`, {
    headers: authHeaders(token),
  })
  return response.data
}

export async function deleteManagedDevice(deviceId: string, token: string): Promise<void> {
  await axios.delete(`${API_BASE_URL}/devices/v1/devices/${deviceId}`, {
    headers: authHeaders(token),
  })
}
