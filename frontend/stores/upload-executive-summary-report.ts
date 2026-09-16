import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useToastNotification } from '~/components/shared/ToastNotification.vue';
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';
import { extractErrorMessage } from '~/utils/error';

export interface UploadedExecutiveSummaryReport {
  id: string;
  title: string;
  description: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  created_at: string;
}

export const useUploadExecutiveSummaryReportStore = defineStore('upload-executive-summary-report', () => {
  const uploadedDocuments = ref<UploadedExecutiveSummaryReport[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');
  const toast = useToastNotification();

  const fetchUploadedDocuments = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-executive-summary-reports`, {
        method: 'GET'
      });
      if (Array.isArray(response)) {
        uploadedDocuments.value = response;
      } else if (response && Array.isArray(response.data)) {
        uploadedDocuments.value = response.data;
      }
    } catch (error: any) {
      console.error('Failed to fetch uploaded executive summary reports:', error);
      errorMsg.value = extractErrorMessage(error, 'Failed to load uploaded executive summary reports.');
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
        const val = (payload as any)[key];
        if (val !== undefined && val !== null) {
          formData.append(key, val);
        }
      });
      await $fetch(`${baseUrl}/uploaded-executive-summary-reports`, {
        method: 'POST',
        body: formData
      });
      toast.showSuccess('Executive summary report uploaded successfully');
      await fetchUploadedDocuments();
      toast.showSuccess('Laporan berhasil diunggah')
    } catch (error: any) {
      console.error('Failed to upload executive summary report:', error);
      const detail = extractErrorMessage(error, 'Failed to upload executive summary report.');
      errorMsg.value = detail;
      toast.showError('Failed to upload executive summary report.', detail);
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
      await $fetch(`${baseUrl}/uploaded-executive-summary-reports/${id}`, {
        method: 'DELETE'
      });
      toast.showSuccess('Executive summary report deleted successfully');
      await fetchUploadedDocuments();
      toast.showSuccess('Laporan berhasil dihapus')
    } catch (error: any) {
      console.error('Failed to delete uploaded executive summary report:', error);
      const detail = extractErrorMessage(error, 'Failed to delete executive summary report.');
      errorMsg.value = detail;
      toast.showError('Failed to delete executive summary report.', detail);
      throw error;
    } finally {
      loading.value = false;
    }
  };

  const viewDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-executive-summary-reports/${id}/download`, {
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
      const response: any = await $fetch(`${baseUrl}/uploaded-executive-summary-reports/${id}/download`, {
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
      console.error('Failed to download executive summary report:', error);
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
