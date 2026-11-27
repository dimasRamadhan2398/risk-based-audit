import { getAuthServiceBaseUrl } from '~/composables/useApiUrl';

export const useApi = () => {
  const apiFetch = $fetch.create({
    baseURL: getAuthServiceBaseUrl(),
    onRequest({ options }) {
      // Add headers if needed (e.g., auth token)
      // options.headers = {
      //   ...options.headers,
      //   Authorization: `Bearer ${token}`,
      // };
    },
  });

  return {
    apiFetch,
  };
};
