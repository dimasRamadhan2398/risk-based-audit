import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useToastNotification } from '~/components/shared/ToastNotification.vue';

import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';

export interface UploadedConsultingDocument {
  id: string;
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  created_at: string;
}

export const useUploadConsultingDocumentStore = defineStore('upload-consulting-document', () => {
  const uploadedDocuments = ref<UploadedConsultingDocument[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');
  const toast = useToastNotification()

  const fetchUploadedDocuments = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-consulting-documents`, {
        method: 'GET'
      });
      if (Array.isArray(response)) {
        uploadedDocuments.value = response;
      } else if (response && Array.isArray(response.data)) {
        uploadedDocuments.value = response.data;
      }
    } catch (error: any) {
      console.error('Failed to fetch uploaded consulting documents:', error);
      errorMsg.value = 'Failed to load uploaded consulting documents.';
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
      formData.append('title', payload.title);
      formData.append('description', payload.description);
      formData.append('fileName', payload.fileName);
      formData.append('fileType', payload.fileType);
      formData.append('file', payload.file);
      await $fetch(`${baseUrl}/uploaded-consulting-documents`, {
        method: 'POST',
        body: formData
      });
      await fetchUploadedDocuments();
      toast.showSuccess('Consulting document uploaded successfully!')
    } catch (error: any) {
      console.error('Failed to upload consulting document:', error);
      toast.showError(error.data?.message || 'Failed to upload consulting document.')
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
      await $fetch(`${baseUrl}/uploaded-consulting-documents/${id}`, {
        method: 'DELETE'
      });
      await fetchUploadedDocuments();
      toast.showSuccess('Consulting document deleted successfully!')
    } catch (error: any) {
      console.error('Failed to delete uploaded consulting document:', error);
      toast.showError(error.data?.message || 'Failed to delete consulting document.')
      throw error;
    } finally {
      loading.value = false;
    }
  };

  const viewDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-consulting-documents/${id}/download`, {
        responseType: 'blob'
      });

      const blob = new Blob([response], { type: response.type || 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      window.open(url, '_blank');
      setTimeout(() => window.URL.revokeObjectURL(url), 10000);
    } catch (error: any) {
      console.error('Failed to view document:', error);
      toast.showError(error.data?.message || 'Failed to view document.')
    }
  };

  const downloadDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-consulting-documents/${id}/download`, {
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
      console.error('Failed to download consulting document:', error);
      toast.showError(error.data?.message || 'Failed to download document.')
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
