const defaultApiUrl = 'http://localhost:8090';

export function apiUrl(): string {
  return import.meta.env.VITE_API_URL || defaultApiUrl;
}

export function usesProxy(): boolean {
  return import.meta.env.VITE_API_PROXY !== 'false';
}

export function apiBaseUrl(): string {
  if (usesProxy()) {
    return `${window.location.origin}/api`;
  }
  return apiUrl().replace(/\/+$/, '');
}
