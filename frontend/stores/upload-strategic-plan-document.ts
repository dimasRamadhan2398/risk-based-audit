import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useToastNotification } from '~/components/shared/ToastNotification.vue';
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';
import { extractErrorMessage } from '~/utils/error';
import { viewDocumentInNewTab } from '~/utils/documentViewer';

export interface UploadedStrategicPlanDocument {
  id: string;
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  created_at: string;
}

export const useUploadStrategicPlanDocumentStore = defineStore('upload-strategic-plan-document', () => {
  const uploadedDocuments = ref<UploadedStrategicPlanDocument[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');
  const toast = useToastNotification();

  const fetchUploadedDocuments = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-strategic-plan-documents`, {
        method: 'GET'
      });
      if (Array.isArray(response)) {
        uploadedDocuments.value = response;
      } else if (response && Array.isArray(response.data)) {
        uploadedDocuments.value = response.data;
      }
    } catch (error: any) {
      console.error('Failed to fetch uploaded strategic plan documents:', error);
      errorMsg.value = extractErrorMessage(error, 'Failed to load uploaded strategic plan documents.');
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
      Object.keys(payload).forEach(key => {
        const value = (payload as any)[key];
        if (value !== undefined && value !== null) {
          formData.append(key, value);
        }
      });
      await $fetch(`${baseUrl}/uploaded-strategic-plan-documents`, {
        method: 'POST',
        body: formData
      });
      toast.showSuccess('Document uploaded successfully');
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to upload strategic plan document:', error);
      const detail = extractErrorMessage(error, 'Failed to upload strategic plan document.');
      errorMsg.value = detail;
      toast.showError('Failed to upload strategic plan document.', detail);
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
      await $fetch(`${baseUrl}/uploaded-strategic-plan-documents/${id}`, {
        method: 'DELETE'
      });
      toast.showSuccess('Document deleted successfully');
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to delete uploaded strategic plan document:', error);
      const detail = extractErrorMessage(error, 'Failed to delete strategic plan document.');
      errorMsg.value = detail;
      toast.showError('Failed to delete strategic plan document.', detail);
      throw error;
    } finally {
      loading.value = false;
    }
  };

  const viewDocument = async (id: string, fileName?: string, title?: string) => {
    const doc = uploadedDocuments.value.find(d => d.id === id);
    const resolvedName = fileName || doc?.fileName || 'document.pdf';
    const resolvedFileName = fileName || doc?.fileName || 'document.pdf';

    try {
      const baseUrl = getAuditServiceBaseUrl();
      await viewDocumentInNewTab(
        async () => {
          try {
            return await $fetch(`${baseUrl}/uploaded-strategic-plan-documents/${id}/view`, {
              responseType: 'blob'
            });
          } catch (_err) {
            return await $fetch(`${baseUrl}/uploaded-strategic-plan-documents/${id}/download`, {
              responseType: 'blob'
            });
          }
        },
        resolvedName,
        resolvedFileName
      );
    } catch (error: any) {
      console.error('Failed to view strategic plan document:', error);
      const detail = extractErrorMessage(error, 'Failed to view document.');
      errorMsg.value = detail;
      toast.showError('Failed to view document.', detail);
    }
  };

  const downloadDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-strategic-plan-documents/${id}/download`, {
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
      console.error('Failed to download strategic plan document:', error);
      const detail = extractErrorMessage(error, 'Failed to download document.');
      errorMsg.value = detail;
      toast.showError('Failed to download strategic plan document.', detail);
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
