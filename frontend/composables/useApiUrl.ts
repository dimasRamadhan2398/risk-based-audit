export const getAuditServiceBaseUrl = () => {
  const config = useRuntimeConfig();
  if (import.meta.client && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
    return 'https://api.auditsphere.app/api/v1';
  }
  return config.public.apiBase || config.public.auditServiceBaseUrl || (import.meta.env.PROD ? 'https://api.auditsphere.app/api/v1' : 'http://localhost:8080/api/v1');
};

export const getMasterServiceBaseUrl = getAuditServiceBaseUrl;
export const getRiskServiceBaseUrl = getAuditServiceBaseUrl;
export const getAuthServiceBaseUrl = getAuditServiceBaseUrl;
export const getAnalyticsServiceBaseUrl = () => {
  const config = useRuntimeConfig();
  if (import.meta.client && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
    return 'https://api.auditsphere.app/api/analytics';
  }
  return config.public.analyticsApiBase || (import.meta.env.PROD ? 'https://api.auditsphere.app/api/analytics' : 'http://localhost:8084/api/analytics');
};
export const getPythonAiBaseUrl = () => {
  const config = useRuntimeConfig();
  if (import.meta.client && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
    return 'https://api.auditsphere.app/api/python-ai';
  }
  return config.public.pythonAiBaseUrl || (import.meta.env.PROD ? 'https://api.auditsphere.app/api/python-ai' : 'http://localhost:8000');
};
