import { defineStore } from 'pinia';
import { ref } from 'vue';

import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';

export interface UploadedAuditResultReport {
  id: string;
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  created_at: string;
}

export const useUploadAuditResultReportStore = defineStore('upload-audit-result-report', () => {
  const uploadedDocuments = ref<UploadedAuditResultReport[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');

  const fetchUploadedDocuments = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-audit-result-reports`, {
        method: 'GET'
      });
      if (Array.isArray(response)) {
        uploadedDocuments.value = response;
      } else if (response && Array.isArray(response.data)) {
        uploadedDocuments.value = response.data;
      }
    } catch (error: any) {
      console.error('Failed to fetch uploaded audit result reports:', error);
      errorMsg.value = 'Failed to load uploaded LHA documents.';
    } finally {
      loading.value = false;
    }
  };

  const uploadDocument = async (payload: { title: string; description: string; fileName: string; fileType: string; fileContent: string }) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/uploaded-audit-result-reports`, {
        method: 'POST',
        body: payload
      });
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to upload LHA document:', error);
      errorMsg.value = error.data?.message || 'Failed to upload LHA document.';
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
      await $fetch(`${baseUrl}/uploaded-audit-result-reports/${id}`, {
        method: 'DELETE'
      });
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to delete uploaded LHA document:', error);
      errorMsg.value = 'Failed to delete LHA document.';
      throw error;
    } finally {
      loading.value = false;
    }
  };

    const viewDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-audit-result-reports/${id}/download`, {
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
      const response: any = await $fetch(`${baseUrl}/uploaded-audit-result-reports/${id}/download`, {
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
      console.error('Failed to download LHA document:', error);
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
