import { defineStore } from 'pinia';
import { ref } from 'vue';

import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';

export interface UploadedExecutiveSummary {
  id: string;
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  created_at: string;
}

export const useUploadExecutiveSummaryStore = defineStore('upload-executive-summary', () => {
  const uploadedDocuments = ref<UploadedExecutiveSummary[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');

  const fetchUploadedDocuments = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const formData = new FormData()
      Object.keys(payload).forEach(key => {
        if (payload[key] !== undefined && payload[key] !== null) {
          formData.append(key, payload[key])
        }
      })
      const response: any = await $fetch(`${baseUrl}/uploaded-executive-summaries`, {
        method: 'GET'
      });
      if (Array.isArray(response)) {
        uploadedDocuments.value = response;
      } else if (response && Array.isArray(response.data)) {
        uploadedDocuments.value = response.data;
      }
    } catch (error: any) {
      console.error('Failed to fetch uploaded executive summaries:', error);
      errorMsg.value = 'Failed to load uploaded executive summary documents.';
    } finally {
      loading.value = false;
    }
  };

  const uploadDocument = async (payload: { title: string; description: string; fileName: string; fileType: string; file: File }) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/uploaded-executive-summaries`, {
        method: 'POST',
        body: formData
      });
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to upload executive summary document:', error);
      errorMsg.value = error.data?.message || 'Failed to upload executive summary document.';
      throw error;
    } finally {
      loading.value = false;
    }
  };

  const deleteDocument = async (id: string) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/uploaded-executive-summaries/${id}`, {
        method: 'DELETE'
      });
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to delete uploaded executive summary document:', error);
      errorMsg.value = 'Failed to delete executive summary document.';
      throw error;
    } finally {
      loading.value = false;
    }
  };

    const viewDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const formData = new FormData()
      Object.keys(payload).forEach(key => {
        if (payload[key] !== undefined && payload[key] !== null) {
          formData.append(key, payload[key])
        }
      })
      const response: any = await $fetch(`${baseUrl}/uploaded-executive-summaries/${id}/download`, {
        responseType: 'blob'
      });
      let mimeType = 'application/pdf'
      if (fileName) {
        const lowerName = fileName.toLowerCase()
        if (lowerName.endsWith('.png')) mimeType = 'image/png'
        else if (lowerName.endsWith('.jpg') || lowerName.endsWith('.jpeg')) mimeType = 'image/jpeg'
        else if (lowerName.endsWith('.txt')) mimeType = 'text/plain'
      }
      
      const blob = new Blob([response], { type: mimeType });
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
      const formData = new FormData()
      Object.keys(payload).forEach(key => {
        if (payload[key] !== undefined && payload[key] !== null) {
          formData.append(key, payload[key])
        }
      })
      const response: any = await $fetch(`${baseUrl}/uploaded-executive-summaries/${id}/download`, {
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
      console.error('Failed to download executive summary document:', error);
      errorMsg.value = 'Failed to download document.';
    }
  };

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
