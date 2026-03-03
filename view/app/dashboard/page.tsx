'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { apiClient } from '@/lib/client-auth';
import { SystemInfo } from '@/types/auth';

export default function DashboardPage() {
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const router = useRouter();

  useEffect(() => {
    fetchSystemInfo();
  }, []);

  const fetchSystemInfo = async () => {
    try {
      console.log('Dashboard: Loading system info...');
      const info = await apiClient.getSystemInfo();
      console.log('Dashboard: System info:', info);
      setSystemInfo(info);

      // If not initialized or not logged in, redirect back to root
      if (!info.hasInit) {
        console.log('Dashboard: Not initialized, redirecting to root');
        router.push('/');
        return;
      }

      if (!info.hasLogin) {
        console.log('Dashboard: Not logged in, redirecting to root');
        router.push('/');
        return;
      }
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('权限验证失败，将跳转到首页');
        router.push('/');
      } else {
        setError('Failed to load dashboard data');
        console.error('Dashboard: Failed to fetch system info:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    try {
      await apiClient.logout();
      router.push('/');
    } catch (err) {
      setError('Failed to logout');
      console.error('Dashboard: Logout error:', err);
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-red-600 mb-4">{error}</div>
        <Button onClick={fetchSystemInfo}>Retry</Button>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">Dashboard</h1>
          <p className="text-gray-600">Welcome to HTTP2SMTP Administration Panel</p>
        </div>
        <Button variant="destructive" onClick={handleLogout}>
          Logout
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>System Status</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div>
                <span className="font-medium">Initialized:</span>
                <span className="ml-2">{systemInfo?.hasInit ? 'Yes' : 'No'}</span>
              </div>
              <div>
                <span className="font-medium">Logged In:</span>
                <span className="ml-2">{systemInfo?.hasLogin ? 'Yes' : 'No'}</span>
              </div>
              <div>
                <span className="font-medium">Docker Environment:</span>
                <span className="ml-2">{systemInfo?.isDocker ? 'Yes' : 'No'}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Button className="w-full" variant="default">
                Manage SMTP Servers
              </Button>
              <Button className="w-full" variant="default">
                Manage Email Recipients
              </Button>
              <Button className="w-full" variant="default">
                Manage API Tokens
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Resources</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Button className="w-full" variant="secondary">
                API Documentation
              </Button>
              <Button className="w-full" variant="secondary">
                View Logs
              </Button>
              <Button className="w-full" variant="secondary">
                System Settings
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
