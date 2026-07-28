import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';

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
      errorMsg.value = 'Failed to load uploaded annual audit plans.';
    } finally {
      loading.value = false;
    }
  };

  const uploadDocument = async (payload: { title: string; description: string; fileName: string; fileType: string; fileContent: string }) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/uploaded-annual-plans`, {
        method: 'POST',
        body: payload
      });
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to upload annual audit plan:', error);
      errorMsg.value = error.data?.message || 'Failed to upload annual audit plan.';
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
    } catch (error: any) {
      console.error('Failed to delete uploaded annual audit plan:', error);
      errorMsg.value = 'Failed to delete annual audit plan.';
      throw error;
    } finally {
      loading.value = false;
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
    downloadDocument
  };
});
