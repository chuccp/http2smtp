'use client';

import { useState, useEffect } from 'react';
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
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setError('Failed to fetch settings');
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
      setSuccess('Settings saved successfully');
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setError('Failed to save settings');
        console.error('Failed to save settings:', err);
      }
    } finally {
      setSaving(false);
    }
  };

  const handleRestart = async () => {
    if (confirm('Are you sure you want to restart the service?')) {
      try {
        await apiClient.restart();
        alert('Restart command sent');
      } catch (err) {
        console.error('Failed to restart:', err);
        setError('Failed to restart');
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
        <div className="text-lg">Loading settings...</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Settings</h1>
        <p className="text-gray-600">Configure system parameters and database connection</p>
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
            <CardTitle>General Settings</CardTitle>
            <CardDescription>Configure service ports</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="managePort">Management Port</Label>
                <Input
                  id="managePort"
                  type="number"
                  value={settings.manage?.port || 12566}
                  onChange={(e) => handleInputChange('manage', 'port', parseInt(e.target.value) || 12566)}
                />
                <p className="text-xs text-gray-500">Web UI management interface port</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="apiPort">API Port</Label>
                <Input
                  id="apiPort"
                  type="number"
                  value={settings.api?.port || 12567}
                  onChange={(e) => handleInputChange('api', 'port', parseInt(e.target.value) || 12567)}
                />
                <p className="text-xs text-gray-500">Email sending API port</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Database Configuration */}
        <Card>
          <CardHeader>
            <CardTitle>Database Configuration</CardTitle>
            <CardDescription>Select database type and configure connection parameters</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="dbType">Database Type</Label>
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
                <Label htmlFor="sqliteFilename">Database Filename</Label>
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
                    <Label htmlFor="mysqlHost">Host *</Label>
                    <Input
                      id="mysqlHost"
                      value={settings.mysql?.host || ''}
                      onChange={(e) => handleInputChange('mysql', 'host', e.target.value)}
                      placeholder="localhost"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="mysqlPort">Port *</Label>
                    <Input
                      id="mysqlPort"
                      type="number"
                      value={settings.mysql?.port || 3306}
                      onChange={(e) => handleInputChange('mysql', 'port', parseInt(e.target.value) || 3306)}
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="mysqlDbname">Database Name *</Label>
                  <Input
                    id="mysqlDbname"
                    value={settings.mysql?.dbname || ''}
                    onChange={(e) => handleInputChange('mysql', 'dbname', e.target.value)}
                    placeholder="d-mail"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="mysqlUsername">Username *</Label>
                    <Input
                      id="mysqlUsername"
                      value={settings.mysql?.username || ''}
                      onChange={(e) => handleInputChange('mysql', 'username', e.target.value)}
                      placeholder="root"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="mysqlPassword">Password *</Label>
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
                  <Label htmlFor="mysqlCharset">Charset</Label>
                  <Input
                    id="mysqlCharset"
                    value={settings.mysql?.charset || 'utf8mb4'}
                    onChange={(e) => handleInputChange('mysql', 'charset', e.target.value)}
                    placeholder="utf8mb4"
                  />
                </div>
                <p className="text-xs text-gray-500">Note: Service restart is required after changing database type</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Account Settings */}
        <Card>
          <CardHeader>
            <CardTitle>Account Settings</CardTitle>
            <CardDescription>Modify administrator account information</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                value={settings.manage?.username || ''}
                onChange={(e) => handleInputChange('manage', 'username', e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">New Password</Label>
              <Input
                id="password"
                type="password"
                value={settings.manage?.password || ''}
                onChange={(e) => handleInputChange('manage', 'password', e.target.value)}
                placeholder="Leave blank to keep current password"
              />
              <p className="text-xs text-gray-500">Enter a new password to change, leave blank to keep the current password</p>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="flex gap-4 mt-6">
        <Button onClick={handleSave} disabled={saving}>
          {saving ? 'Saving...' : 'Save Settings'}
        </Button>
        <Button variant="outline" onClick={handleRestart}>
          Restart Service
        </Button>
      </div>
    </div>
  );
}