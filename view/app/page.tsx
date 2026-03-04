'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { SystemInfo } from '@/types/auth';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:12566';

export default function RootPage() {
  const router = useRouter();
  const tCommon = useTranslations('common');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const checkAndRedirect = async () => {
      try {
        console.log('Root: Checking system status...');
        const response = await fetch(`${API_URL}/set`, {
          credentials: 'include',
        });

        if (response.status === 401) {
          console.log('Root: Unauthorized, redirecting to login');
          router.push('/auth/login');
          return;
        }

        const responseData = await response.json();
        console.log('Root: API response:', responseData);

        // Parse system info
        let systemInfo: SystemInfo;
        if (responseData.data && responseData.data.hasInit !== undefined) {
          systemInfo = responseData.data;
        } else if (responseData.hasInit !== undefined) {
          systemInfo = responseData;
        } else {
          systemInfo = { hasInit: false, hasLogin: false, isDocker: false };
        }

        console.log('Root: System info:', systemInfo);

        if (!systemInfo.hasInit) {
          console.log('Root: Not initialized, redirecting to setup');
          router.push('/setup');
        } else if (!systemInfo.hasLogin) {
          console.log('Root: Initialized but not logged in, redirecting to login');
          router.push('/auth/login');
        } else {
          console.log('Root: Already logged in, redirecting to dashboard');
          router.push('/dashboard');
        }
      } catch (err) {
        console.error('Root: Error checking system status:', err);
        // On error, default to setup page
        router.push('/setup');
      } finally {
        setLoading(false);
      }
    };

    checkAndRedirect();
  }, [router]);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">{tCommon('loading')}</div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-lg">{tCommon('loading')}</div>
    </div>
  );
}
