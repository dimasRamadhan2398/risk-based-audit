import { defineStore } from 'pinia';
import { ref } from 'vue';
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';

export interface UploadedPerformanceReport {
  id: string;
  title: string;
  period: string; // Q1, Q2, Q3, Q4, Tahunan
  year: number;
  description: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  status: string;
  parsedKpisCount: number;
  created_at: string;
}

export const useUploadPerformanceReportStore = defineStore('upload-performance-report', () => {
  const uploadedReports = ref<UploadedPerformanceReport[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');

  const fetchUploadedReports = async (period?: string, year?: number) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const params: Record<string, any> = {};
      if (period && period !== 'Semua') {
        params.period = period;
      }
      if (year) {
        params.year = year;
      }

      const response: any = await $fetch(`${baseUrl}/uploaded-performance-reports`, {
        method: 'GET',
        params
      });

      if (Array.isArray(response)) {
        uploadedReports.value = response;
      } else if (response && Array.isArray(response.data)) {
        uploadedReports.value = response.data;
      } else {
        uploadedReports.value = [];
      }
    } catch (error: any) {
      console.error('Failed to fetch uploaded performance reports:', error);
      errorMsg.value = 'Failed to load uploaded performance report documents.';
    } finally {
      loading.value = false;
    }
  };

  const uploadReport = async (payload: {
    title: string;
    period: string;
    year: number;
    description: string;
    fileName: string;
    fileType: string;
    fileContent: string;
  }) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/uploaded-performance-reports`, {
        method: 'POST',
        body: payload
      });
      await fetchUploadedReports(payload.period, payload.year);
    } catch (error: any) {
      console.error('Failed to upload performance report:', error);
      errorMsg.value = error.data?.message || 'Failed to upload performance report document.';
      throw error;
    } finally {
      loading.value = false;
    }
  };

  const deleteReport = async (id: string, currentPeriod?: string, currentYear?: number) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/uploaded-performance-reports/${id}`, {
        method: 'DELETE'
      });
      await fetchUploadedReports(currentPeriod, currentYear);
    } catch (error: any) {
      console.error('Failed to delete performance report document:', error);
      errorMsg.value = 'Failed to delete performance report document.';
      throw error;
    } finally {
      loading.value = false;
    }
  };

  const downloadReport = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-performance-reports/${id}/download`, {
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
      console.error('Failed to download performance report document:', error);
      errorMsg.value = 'Failed to download document.';
    }
  };

  const viewDocument = async (id: string, fileName: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/uploaded-performance-reports/${id}/download`, {
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
      console.error('Failed to view performance report document:', error);
      errorMsg.value = 'Failed to view document.';
    }
  };

  return {
    uploadedReports,
    loading,
    errorMsg,
    fetchUploadedReports,
    uploadReport,
    deleteReport,
    downloadReport,
    viewDocument
  };
});
