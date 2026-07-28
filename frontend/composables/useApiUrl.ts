export const getAuditServiceBaseUrl = () => {
  if (import.meta.client) {
    const hostname = window.location.hostname;
    if (hostname === 'localhost' || hostname === '127.0.0.1') {
      return 'http://localhost:8080/api/v1';
    }
  }
  const config = useRuntimeConfig();
  const envUrl = config.public.auditServiceBaseUrl;
  if (envUrl && typeof envUrl === 'string' && envUrl.length > 0) {
    if (import.meta.client) {
      const hostname = window.location.hostname;
      if (hostname === 'localhost' || hostname === '127.0.0.1') {
        return 'http://localhost:8080/api/v1';
      }
    }
    return envUrl;
  }
  return import.meta.env.PROD ? 'https://api.auditsphere.app/api/v1' : 'http://localhost:8080/api/v1';
};
