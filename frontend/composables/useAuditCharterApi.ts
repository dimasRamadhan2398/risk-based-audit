export type AuditCharterPayload = {
  filename: string
  version: string
  title: string
  content: string
  is_active: boolean
}

export const useAuditCharterApi = () => {
  const config = useRuntimeConfig()
  const baseUrl = config.public.auditServiceBaseUrl

  const getAuditCharters = async () => {
    return await $fetch(`${baseUrl}/audit-charters`, {
      method: 'GET'
    })
  }

  const getAuditCharterById = async (id: string | number) => {
    return await $fetch(`${baseUrl}/audit-charters/${id}`, {
      method: 'GET'
    })
  }

  const createAuditCharter = async (payload: AuditCharterPayload) => {
    return await $fetch(`${baseUrl}/audit-charters`, {
      method: 'POST',
      body: payload
    })
  }

  const updateAuditCharter = async (id: string | number, payload: AuditCharterPayload) => {
    return await $fetch(`${baseUrl}/audit-charters/${id}`, {
      method: 'PUT',
      body: payload
    })
  }

  const deleteAuditCharter = async (id: string | number) => {
    return await $fetch(`${baseUrl}/audit-charters/${id}`, {
      method: 'DELETE'
    })
  }

  const downloadAuditCharter = async (id: string | number, filename: string) => {
    const response = await $fetch<Blob>(`${baseUrl}/audit-charters/${id}/download`, {
      method: 'GET',
      responseType: 'blob'
    })

    // Create a blob URL and trigger download
    const blob = new Blob([response])
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  }

  return {
    getAuditCharters,
    getAuditCharterById,
    createAuditCharter,
    updateAuditCharter,
    deleteAuditCharter,
    downloadAuditCharter
  }
}
