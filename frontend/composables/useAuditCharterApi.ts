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

  return {
    getAuditCharters,
    getAuditCharterById,
    createAuditCharter,
    updateAuditCharter,
    deleteAuditCharter
  }
}