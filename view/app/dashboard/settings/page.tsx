'use client';

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { SetInfo } from '@/types/settings';

export default function SettingsPage() {
  const router = useRouter();
  const t = useTranslations('settings');
  const tCommon = useTranslations('common');
  const tAuth = useTranslations('auth');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [settings, setSettings] = useState<SetInfo>({
    hasInit: true,
    dbType: 'sqlite',
    cachePath: '.cache',
    sqlite: { filename: 'd-mail.db' },
    mysql: { host: 'localhost', port: 3306, dbname: 'd-mail', username: '', password: '', charset: 'utf8' },
    manage: { port: 12566, webPath: 'web', username: '', password: '' },
    api: { port: 12567 },
    isDocker: false,
  });

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getSettings();
      setSettings(data);
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert(tAuth('authFailed'));
        router.push('/');
      } else {
        setError(t('fetchError'));
        console.error('Failed to fetch settings:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setError('');
    setSuccess('');

    try {
      await apiClient.updateSettings(settings);
      setSuccess(t('saveSuccess'));
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert(tAuth('authFailed'));
        router.push('/');
      } else {
        setError(t('saveError'));
        console.error('Failed to save settings:', err);
      }
    } finally {
      setSaving(false);
    }
  };

  const handleRestart = async () => {
    if (confirm(t('restartConfirm'))) {
      try {
        await apiClient.restart();
        alert(t('restartSuccess'));
      } catch (err) {
        console.error('Failed to restart:', err);
        setError(t('restartError'));
      }
    }
  };

  const handleInputChange = (section: string, field: string, value: string | number) => {
    setSettings(prev => ({
      ...prev,
      [section]: {
        ...prev[section as keyof SetInfo] as object,
        [field]: value,
      },
    }));
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">{tCommon('loading')}</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">{t('title')}</h1>
        <p className="text-gray-600">{t('subtitle')}</p>
      </div>

      {error && (
        <Alert variant="destructive" className="mb-6">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {success && (
        <Alert className="mb-6 border-green-500 text-green-700">
          <AlertDescription>{success}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-6">
        {/* General Settings */}
        <Card>
          <CardHeader>
            <CardTitle>{t('generalSettings')}</CardTitle>
            <CardDescription>{t('generalSettingsDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="managePort">{t('managementPort')}</Label>
                <Input
                  id="managePort"
                  type="number"
                  value={settings.manage?.port || 12566}
                  onChange={(e) => handleInputChange('manage', 'port', parseInt(e.target.value) || 12566)}
                />
                <p className="text-xs text-gray-500">{t('managementPortDesc')}</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="apiPort">{t('apiPort')}</Label>
                <Input
                  id="apiPort"
                  type="number"
                  value={settings.api?.port || 12567}
                  onChange={(e) => handleInputChange('api', 'port', parseInt(e.target.value) || 12567)}
                />
                <p className="text-xs text-gray-500">{t('apiPortDesc')}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Database Configuration */}
        <Card>
          <CardHeader>
            <CardTitle>{t('databaseConfig')}</CardTitle>
            <CardDescription>{t('databaseConfigDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="dbType">{t('databaseType')}</Label>
              <select
                id="dbType"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                value={settings.dbType || 'sqlite'}
                onChange={(e) => setSettings(prev => ({ ...prev, dbType: e.target.value }))}
              >
                <option value="sqlite">SQLite</option>
                <option value="mysql">MySQL</option>
              </select>
            </div>

            {settings.dbType === 'sqlite' && settings.sqlite && (
              <div className="space-y-2">
                <Label htmlFor="sqliteFilename">{t('databaseFilename')}</Label>
                <Input
                  id="sqliteFilename"
                  value={settings.sqlite.filename}
                  onChange={(e) => handleInputChange('sqlite', 'filename', e.target.value)}
                />
              </div>
            )}

            {settings.dbType === 'mysql' && (
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="mysqlHost">{t('host')} {tCommon('required')}</Label>
                    <Input
                      id="mysqlHost"
                      value={settings.mysql?.host || ''}
                      onChange={(e) => handleInputChange('mysql', 'host', e.target.value)}
                      placeholder="localhost"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="mysqlPort">{t('port')} {tCommon('required')}</Label>
                    <Input
                      id="mysqlPort"
                      type="number"
                      value={settings.mysql?.port || 3306}
                      onChange={(e) => handleInputChange('mysql', 'port', parseInt(e.target.value) || 3306)}
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="mysqlDbname">{t('databaseName')} {tCommon('required')}</Label>
                  <Input
                    id="mysqlDbname"
                    value={settings.mysql?.dbname || ''}
                    onChange={(e) => handleInputChange('mysql', 'dbname', e.target.value)}
                    placeholder="d-mail"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="mysqlUsername">{t('dbUsername')} {tCommon('required')}</Label>
                    <Input
                      id="mysqlUsername"
                      value={settings.mysql?.username || ''}
                      onChange={(e) => handleInputChange('mysql', 'username', e.target.value)}
                      placeholder="root"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="mysqlPassword">{t('dbPassword')} {tCommon('required')}</Label>
                    <Input
                      id="mysqlPassword"
                      type="password"
                      value={settings.mysql?.password || ''}
                      onChange={(e) => handleInputChange('mysql', 'password', e.target.value)}
                      placeholder="Enter password"
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="mysqlCharset">{t('charset')}</Label>
                  <Input
                    id="mysqlCharset"
                    value={settings.mysql?.charset || 'utf8mb4'}
                    onChange={(e) => handleInputChange('mysql', 'charset', e.target.value)}
                    placeholder="utf8mb4"
                  />
                </div>
                <p className="text-xs text-gray-500">{t('dbChangeNote')}</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Account Settings */}
        <Card>
          <CardHeader>
            <CardTitle>{t('accountSettings')}</CardTitle>
            <CardDescription>{t('accountSettingsDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">{t('username')}</Label>
              <Input
                id="username"
                value={settings.manage?.username || ''}
                onChange={(e) => handleInputChange('manage', 'username', e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">{t('newPassword')}</Label>
              <Input
                id="password"
                type="password"
                value={settings.manage?.password || ''}
                onChange={(e) => handleInputChange('manage', 'password', e.target.value)}
                placeholder="Leave blank to keep current password"
              />
              <p className="text-xs text-gray-500">{t('newPasswordDesc')}</p>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="flex gap-4 mt-6">
        <Button onClick={handleSave} disabled={saving}>
          {saving ? tCommon('saving') : t('saveSettings')}
        </Button>
        <Button variant="outline" onClick={handleRestart}>
          {t('restartService')}
        </Button>
      </div>
    </div>
  );
}