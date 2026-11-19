import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useToastNotification } from '~/components/shared/ToastNotification.vue';

import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';

import { useI18n } from '~/composables/useI18n';
import { extractErrorMessage } from '~/utils/error';
import { viewDocumentInNewTab } from '~/utils/documentViewer';

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
  const { t } = useI18n()


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
      errorMsg.value = extractErrorMessage(error, t('toasts.document.loadFailed', { name: t('toasts.documentTypes.assignmentLetter') }));
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
      toast.showSuccess(t('toasts.document.uploaded', { name: t('toasts.documentTypes.assignmentLetter') }))
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to upload assignment letter document:', error);
      const detail = extractErrorMessage(error, t('toasts.document.uploadFailed', { name: t('toasts.documentTypes.assignmentLetter') }));
      errorMsg.value = detail;
      toast.showError(t('toasts.document.uploadFailed', { name: t('toasts.documentTypes.assignmentLetter') }), detail);
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
      toast.showSuccess(t('toasts.document.deleted', { name: t('toasts.documentTypes.assignmentLetter') }))
      await fetchUploadedDocuments();
    } catch (error: any) {
      console.error('Failed to delete uploaded assignment letter:', error);
      const detail = extractErrorMessage(error, t('toasts.document.deleteFailed', { name: t('toasts.documentTypes.assignmentLetter') }));
      errorMsg.value = detail;
      toast.showError(t('toasts.document.deleteFailed', { name: t('toasts.documentTypes.assignmentLetter') }), detail);
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
            return await $fetch(`${baseUrl}/uploaded-assignment-letters/${id}/view`, {
              responseType: 'blob'
            });
          } catch (_err) {
            return await $fetch(`${baseUrl}/uploaded-assignment-letters/${id}/download`, {
              responseType: 'blob'
            });
          }
        },
        resolvedName,
        resolvedFileName
      );
    } catch (error: any) {
      console.error('Failed to view document:', error);
      const detail = extractErrorMessage(error, t('toasts.document.viewFailed'));
      errorMsg.value = detail;
      toast.showError(t('toasts.document.viewFailed'), detail);
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
      const detail = extractErrorMessage(error, t('toasts.document.downloadFailed'));
      errorMsg.value = detail;
      toast.showError(t('toasts.document.downloadFailed'), detail);
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
