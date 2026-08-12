export const getAuditServiceBaseUrl = () => {
  const config = useRuntimeConfig();
  return config.public.auditServiceBaseUrl || config.public.apiBase || '/api/v1';
};

export const getMasterServiceBaseUrl = getAuditServiceBaseUrl;
export const getRiskServiceBaseUrl = getAuditServiceBaseUrl;
export const getAuthServiceBaseUrl = getAuditServiceBaseUrl;

export const getAnalyticsServiceBaseUrl = () => {
  const config = useRuntimeConfig();
  return config.public.analyticsApiBase || '/api/analytics';
};

export const getPythonAiBaseUrl = () => {
  const config = useRuntimeConfig();
  return config.public.pythonAiBaseUrl || '/api/python-ai';
};
