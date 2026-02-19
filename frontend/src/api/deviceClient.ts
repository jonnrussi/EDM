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

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export async function registerInventoryEndpoint(
  payload: RegisterInventoryEndpointRequest,
  token: string,
): Promise<RegisterInventoryEndpointResponse> {
  const response = await axios.post<RegisterInventoryEndpointResponse>(
    `${API_BASE_URL}/devices/v1/devices`,
    payload,
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    },
  )

  return response.data
}
