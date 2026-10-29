const isRemoteUrl = (val: unknown): val is string =>
  typeof val === 'string' && /^https?:\/\//.test(val) && !/\/\/(localhost|127\.0\.0\.1)(:|\/|$)/.test(val);

const getDynamicApiUrl = (fallbackConfigKey: string, fallbackPath: string) => {
  const configured = useRuntimeConfig().public[fallbackConfigKey as keyof ReturnType<typeof useRuntimeConfig>['public']];
  // An absolute non-localhost URL from .env (e.g. https://api.auditsphere.app/api/v1) wins,
  // so local dev can target a remote API. Relative values keep the behaviour below.
  if (isRemoteUrl(configured)) {
    return configured;
  }
  if (process.client) {
    const hostname = window.location.hostname;
    if (hostname !== 'localhost' && hostname !== '127.0.0.1') {
      return fallbackPath;
    } else {
      const port = fallbackPath.includes('analytics') ? '8084' : '8080';
      return `http://localhost:${port}${fallbackPath}`;
    }
  }
  const config = useRuntimeConfig();
  const val = config.public[fallbackConfigKey as keyof typeof config.public];
  if (val && typeof val === 'string' && !val.includes('localhost')) {
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
