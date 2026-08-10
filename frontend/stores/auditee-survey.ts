import { defineStore } from 'pinia';
import { ref } from 'vue';
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';

export interface AuditeeSurvey {
  id?: string;
  audit_execution_id?: string;
  auditee_name: string;
  department: string;
  year: number;
  month: number;
  rating_clarity: number;
  rating_professionalism: number;
  rating_timeliness: number;
  overall_score?: number;
  comments: string;
  created_at?: string;
  updated_at?: string;
}

export const useAuditeeSurveyStore = defineStore('auditee-survey', () => {
  const surveys = ref<AuditeeSurvey[]>([]);
  const loading = ref(false);
  const errorMsg = ref('');

  const fetchSurveys = async () => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/auditee-surveys`, {
        method: 'GET'
      });
      if (response && Array.isArray(response.items)) {
        surveys.value = response.items;
      } else if (Array.isArray(response)) {
        surveys.value = response;
      } else if (response && Array.isArray(response.data)) {
        surveys.value = response.data;
      }
    } catch (error: any) {
      console.error('Failed to fetch auditee surveys:', error);
      errorMsg.value = 'Failed to load auditee surveys.';
    } finally {
      loading.value = false;
    }
  };

  const createSurvey = async (payload: AuditeeSurvey) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response = await $fetch(`${baseUrl}/auditee-surveys`, {
        method: 'POST',
        body: payload
      });
      await fetchSurveys();
      return response;
    } catch (error: any) {
      console.error('Failed to create auditee survey:', error);
      errorMsg.value = error.data?.message || 'Failed to submit the survey.';
      throw error;
    } finally {
      loading.value = false;
    }
  };

  const deleteSurvey = async (id: string) => {
    loading.value = true;
    errorMsg.value = '';
    try {
      const baseUrl = getAuditServiceBaseUrl();
      await $fetch(`${baseUrl}/auditee-surveys/${id}`, {
        method: 'DELETE'
      });
      await fetchSurveys();
    } catch (error: any) {
      console.error('Failed to delete auditee survey:', error);
      errorMsg.value = 'Failed to delete survey.';
      throw error;
    } finally {
      loading.value = false;
    }
  };

  return {
    surveys,
    loading,
    errorMsg,
    fetchSurveys,
    createSurvey,
    deleteSurvey
  };
});
