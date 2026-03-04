'use client';

import { useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useTranslations } from 'next-intl';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { LanguageSwitcher } from '@/lib/LanguageSwitcher';
import { apiClient } from '@/lib/client-auth';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const router = useRouter();
  const pathname = usePathname();
  const t = useTranslations('auth');
  const tCommon = useTranslations('common');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await apiClient.login(username, password);

      if (response.code === 200) {
        // Login successful, redirect to root which will handle redirect to dashboard
        router.push('/');
      } else {
        setError(response.message || 'Login failed');
      }
    } catch (err) {
      setError('Failed to connect to server');
      console.error('Login error:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="w-full max-w-md space-y-8">
        <div className="flex justify-end">
          <LanguageSwitcher />
        </div>
        <div>
          <h2 className="mt-6 text-center text-3xl font-bold tracking-tight text-gray-900">
            {t('signIn')}
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            {t('signInDesc')}
          </p>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>{t('login')}</CardTitle>
            <CardDescription>
              {t('loginDesc')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-4" onSubmit={handleLogin}>
              {error && (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}
              <div className="space-y-2">
                <Label htmlFor="username">{t('username')}</Label>
                <Input
                  id="username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                  disabled={loading}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">{t('password')}</Label>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  disabled={loading}
                />
              </div>
              <Button
                type="submit"
                className="w-full"
                disabled={loading || !username || !password}
              >
                {loading ? t('signingIn') : t('signIn')}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
