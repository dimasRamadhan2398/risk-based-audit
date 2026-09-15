import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useToastNotification } from '~/components/shared/ToastNotification.vue';
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';
import { extractErrorMessage } from '~/utils/error';

export interface UploadedAnnualPlan {
  id: string;
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  created_at: string;
}

export const useUploadAnnualPlanStore = defineStore('upload-annual-plan', () => {
  const uploadedDocuments = ref<UploadedAnnualPlan[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');
  const toast = useToastNotification();

  const fetchUploadedDocuments = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-annual-plans`, {
        method: 'GET'
      });
      if (Array.isArray(response)) {
        uploadedDocuments.value = response;
      } else if (response && Array.isArray(response.data)) {
        uploadedDocuments.value = response.data;
      }
    } catch (error: any) {
      console.error('Failed to fetch uploaded annual plans:', error);
      errorMsg.value = extractErrorMessage(error, 'Failed to load uploaded annual audit plans.');
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
      await $fetch(`${baseUrl}/uploaded-annual-plans`, {
        method: 'POST',
        body: formData
      });
      await fetchUploadedDocuments();
      toast.showSuccess('Successfully uploaded annual audit plan');
    } catch (error: any) {
      console.error('Failed to upload annual audit plan:', error);
      const detail = extractErrorMessage(error, 'Failed to upload annual audit plan.');
      errorMsg.value = detail;
      toast.showError('Failed to upload annual audit plan.', detail);
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
      await $fetch(`${baseUrl}/uploaded-annual-plans/${id}`, {
        method: 'DELETE'
      });
      await fetchUploadedDocuments();
      toast.showSuccess('Successfully deleted annual audit plan');
    } catch (error: any) {
      console.error('Failed to delete uploaded annual audit plan:', error);
      const detail = extractErrorMessage(error, 'Failed to delete annual audit plan.');
      errorMsg.value = detail;
      toast.showError('Failed to delete annual audit plan.', detail);
      throw error;
    } finally {
      loading.value = false;
    }
  };

  const viewDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-annual-plans/${id}/download`, {
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
      const response: any = await $fetch(`${baseUrl}/uploaded-annual-plans/${id}/download`, {
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
      console.error('Failed to download annual audit plan document:', error);
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
