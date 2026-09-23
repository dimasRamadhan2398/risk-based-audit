const getDynamicApiUrl = (fallbackConfigKey: string, fallbackPath: string) => {
  if (process.client) {
    const hostname = window.location.hostname;
    if (hostname.includes('auditsphere.app')) {
      return 'https://auditsphere.app' + fallbackPath;
    } else if (hostname === 'localhost' || hostname === '127.0.0.1') {
      const port = fallbackPath.includes('analytics') ? '8084' : '8080';
      return `http://localhost:${port}${fallbackPath}`;
    }
  }
  const config = useRuntimeConfig();
  return config.public[fallbackConfigKey as keyof typeof config.public] || fallbackPath;
};

export const getAuditServiceBaseUrl = () => getDynamicApiUrl('auditServiceBaseUrl', '/api/v1');
export const getMasterServiceBaseUrl = () => getDynamicApiUrl('masterServiceBaseUrl', '/api/v1');
export const getRiskServiceBaseUrl = () => getDynamicApiUrl('riskServiceBaseUrl', '/api/v1');
export const getAuthServiceBaseUrl = () => getDynamicApiUrl('authServiceBaseUrl', '/api/v1');
export const getAnalyticsServiceBaseUrl = () => getDynamicApiUrl('analyticsApiBase', '/api/analytics');
export const getPythonAiBaseUrl = () => getDynamicApiUrl('pythonAiBaseUrl', '/api/python-ai');
