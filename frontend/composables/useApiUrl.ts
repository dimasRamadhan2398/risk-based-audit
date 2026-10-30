const isRemoteUrl = (val: unknown): val is string =>
  typeof val === 'string' && /^https?:\/\//.test(val) && !/\/\/(localhost|127\.0\.0\.1)(:|\/|$)/.test(val);

const getDynamicApiUrl = (fallbackConfigKey: string, fallbackPath: string) => {
  const config = useRuntimeConfig();
  const configured = config.public[fallbackConfigKey as keyof typeof config.public] as string | undefined;
  if (isRemoteUrl(configured)) {
    return configured;
  }
  const val =
    configured ||
    (!fallbackPath.includes('analytics') && !fallbackPath.includes('python-ai')
      ? (config.public.apiBase as string | undefined)
      : undefined);

  if (val && typeof val === 'string') {
    return val;
  }

  return fallbackPath;
};

export const getAuditServiceBaseUrl = () => getDynamicApiUrl('auditServiceBaseUrl', '/api/v1');
export const getMasterServiceBaseUrl = () => getDynamicApiUrl('masterServiceBaseUrl', '/api/v1');
export const getRiskServiceBaseUrl = () => getDynamicApiUrl('riskServiceBaseUrl', '/api/v1');
export const getAuthServiceBaseUrl = () => getDynamicApiUrl('authServiceBaseUrl', '/api/v1');
export const getAnalyticsServiceBaseUrl = () => getDynamicApiUrl('analyticsApiBase', '/api/analytics');
export const getPythonAiBaseUrl = () => getDynamicApiUrl('pythonAiBaseUrl', '/api/python-ai');
export const getDataHubBaseUrl = () => getDynamicApiUrl('dataHubBaseUrl', '/api/v1');
