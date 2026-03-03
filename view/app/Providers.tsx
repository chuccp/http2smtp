'use client';

import { useState, useEffect } from 'react';
import { SystemInfo } from '@/types/auth';
import { apiClient } from '@/lib/client-auth';

export interface AuthProviderProps {
  children: React.ReactNode;
}

export interface AppState {
  systemInfo: SystemInfo | null;
  loading: boolean;
  error: string | null;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [appState, setAppState] = useState<AppState>({
    systemInfo: null,
    loading: true,
    error: null,
  });

  useEffect(() => {
    fetchSystemInfo();
  }, []);

  const fetchSystemInfo = async () => {
    try {
      const systemInfo = await apiClient.getSystemInfo();
      setAppState({
        systemInfo,
        loading: false,
        error: null,
      });
    } catch (err) {
      setAppState({
        systemInfo: null,
        loading: false,
        error: 'Failed to load system information',
      });
      console.error('Failed to fetch system info:', err);
    }
  };

  const refreshSystemInfo = () => {
    fetchSystemInfo();
  };

  if (appState.loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  if (appState.error && !appState.systemInfo) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-4">Connection Error</h2>
          <p className="text-red-600 mb-4">{appState.error}</p>
          <button
            onClick={refreshSystemInfo}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div>
      {children}
    </div>
  );
}