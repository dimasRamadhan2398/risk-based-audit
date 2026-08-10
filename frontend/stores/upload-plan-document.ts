import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface UploadedPlanDocument {
  id: string;
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  created_at: string;
}

export const useUploadPlanDocumentStore = defineStore('upload-plan-document', () => {
  const uploadedDocuments = ref<UploadedPlanDocument[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');

  const getAuditServiceBaseUrlLocal = () => {
    return getAuditServiceBaseUrl();
  };

  const fetchUploadedDocuments = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-plan-documents`, {
        method: 'GET'
      });
      if (Array.isArray(response)) {
        uploadedDocuments.value = response;
      } else if (response && Array.isArray(response.data)) {
        uploadedDocuments.value = response.data;
      }
    } catch (error: any) {
      console.error('Failed to fetch uploaded plan documents:', error);
      errorMsg.value = 'Failed to load uploaded plan documents.';
    } finally {
      loading.value = false;
    }
  }

  const uploadDocument = async (payload: { title: string; description: string; fileName: string; fileType: string; fileContent: string }) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/uploaded-plan-documents`, {
        method: 'POST',
        body: payload
      });
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to upload plan document:', error);
      errorMsg.value = error.data?.message || 'Failed to upload plan document.';
      throw error;
    } finally {
      loading.value = false;
    }
  }

  const deleteDocument = async (id: string) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/uploaded-plan-documents/${id}`, {
        method: 'DELETE'
      });
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to delete uploaded plan document:', error);
      errorMsg.value = 'Failed to delete plan document.';
      throw error;
    } finally {
      loading.value = false;
    }
  }

    const viewDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-plan-documents/${id}/download`, {
        responseType: 'blob'
      });
      
      const blob = new Blob([response], { type: response.type || 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      window.open(url, '_blank');
      setTimeout(() => window.URL.revokeObjectURL(url), 10000);
    } catch (error: any) {
      console.error('Failed to view document:', error);
      errorMsg.value = 'Failed to view document.';
    }
  };

  const downloadDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-plan-documents/${id}/download`, {
        responseType: 'blob'
      });
      
      const blob = new Blob([response], { type: response.type || 'application/octet-stream' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = fileName;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error: any) {
      console.error('Failed to download plan document:', error);
      errorMsg.value = 'Failed to download plan document.';
    }
  }

  return {
    uploadedDocuments,
    loading,
    errorMsg,
    fetchUploadedDocuments,
    uploadDocument,
    deleteDocument,
    downloadDocument,
    viewDocument
  };
});
