'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { apiClient } from '@/lib/client-auth';
import { SystemInfo } from '@/types/auth';


export default function DashboardPage() {
  const t = useTranslations('dashboard');
  const tCommon = useTranslations('common');
  const tAuth = useTranslations('auth');
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
        alert(tAuth('authFailed'));
        router.push('/');
      } else {
        setError(t('fetchError'));
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
      setError(t('logoutError'));
      console.error('Dashboard: Logout error:', err);
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">{tCommon('loading')}</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-red-600 mb-4">{error}</div>
        <Button onClick={fetchSystemInfo}>{t('retry')}</Button>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">{t('title')}</h1>
          <p className="text-gray-600">{t('welcome')}</p>
        </div>
        <div className="flex items-center gap-4">
          <Button variant="destructive" onClick={handleLogout}>
            {tCommon('logout')}
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>{t('systemStatus')}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div>
                <span className="font-medium">{t('initialized')}:</span>
                <span className="ml-2">{systemInfo?.hasInit ? tCommon('yes') : tCommon('no')}</span>
              </div>
              <div>
                <span className="font-medium">{t('loggedIn')}:</span>
                <span className="ml-2">{systemInfo?.hasLogin ? tCommon('yes') : tCommon('no')}</span>
              </div>
              <div>
                <span className="font-medium">{t('dockerEnv')}:</span>
                <span className="ml-2">{systemInfo?.isDocker ? tCommon('yes') : tCommon('no')}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('quickActions')}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Button className="w-full" variant="default">
                {t('manageSmtp')}
              </Button>
              <Button className="w-full" variant="default">
                {t('manageRecipients')}
              </Button>
              <Button className="w-full" variant="default">
                {t('manageTokens')}
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('resources')}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Button className="w-full" variant="secondary">
                {t('apiDocs')}
              </Button>
              <Button className="w-full" variant="secondary">
                {t('viewLogs')}
              </Button>
              <Button className="w-full" variant="secondary">
                {t('systemSettings')}
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}