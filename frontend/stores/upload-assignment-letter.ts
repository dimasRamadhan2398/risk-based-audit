import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useToastNotification } from '~/components/shared/ToastNotification.vue';

import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';

import { extractErrorMessage } from '~/utils/error';

export interface UploadedAssignmentLetter {
  id: string;
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  created_at: string;
}

export const useUploadAssignmentLetterStore = defineStore('upload-assignment-letter', () => {
  const uploadedDocuments = ref<UploadedAssignmentLetter[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');
  const toast = useToastNotification()


  const fetchUploadedDocuments = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-assignment-letters`, {
        method: 'GET'
      });
      if (Array.isArray(response)) {
        uploadedDocuments.value = response;
      } else if (response && Array.isArray(response.data)) {
        uploadedDocuments.value = response.data;
      }
    } catch (error: any) {
      console.error('Failed to fetch uploaded assignment letters:', error);
      errorMsg.value = extractErrorMessage(error, 'Failed to load uploaded assignment letters.');
    } finally {
      loading.value = false;
    }
  };

  const uploadDocument = async (payload: { title: string; description: string; fileName: string; fileType: string; file: File }) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const formData = new FormData();
      Object.entries(payload).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          formData.append(key, value as any);
        }
      });
      await $fetch(`${baseUrl}/uploaded-assignment-letters`, {
        method: 'POST',
        body: formData
      });
      toast.showSuccess('Assignment letter document uploaded successfully')
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to upload assignment letter document:', error);
      const detail = extractErrorMessage(error, 'Failed to upload assignment letter document.');
      errorMsg.value = detail;
      toast.showError('Failed to upload assignment letter document.', detail);
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
      await $fetch(`${baseUrl}/uploaded-assignment-letters/${id}`, {
        method: 'DELETE'
      });
      toast.showSuccess('Assignment letter document deleted successfully')
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to delete uploaded assignment letter:', error);
      const detail = extractErrorMessage(error, 'Failed to delete assignment letter.');
      errorMsg.value = detail;
      toast.showError('Failed to delete assignment letter.', detail);
      throw error;
    } finally {
      loading.value = false;
    }
  };

  const viewDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-assignment-letters/${id}/download`, {
        responseType: 'blob'
      });

      const blob = new Blob([response], { type: response.type || 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      window.open(url, '_blank');
      setTimeout(() => window.URL.revokeObjectURL(url), 10000);
    } catch (error: any) {
      console.error('Failed to view document:', error);
      const detail = extractErrorMessage(error, 'Failed to view document.');
      errorMsg.value = detail;
      toast.showError('Failed to view document.', detail);
    }
  };

  const downloadDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-assignment-letters/${id}/download`, {
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
      console.error('Failed to download assignment letter document:', error);
      const detail = extractErrorMessage(error, 'Failed to download document.');
      errorMsg.value = detail;
      toast.showError('Failed to download document.', detail);
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
