'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { LanguageSwitcher } from '@/lib/LanguageSwitcher';
import type { SetInfo, SystemInfo } from '@/types/auth';

interface SetupFormData {
  dbType: 'sqlite' | 'mysql';
  sqlite: {
    filename: string;
  };
  mysql: {
    host: string;
    port: string;
    dbname: string;
    username: string;
    password: string;
    charset: string;
  };
  manage: {
    username: string;
    password: string;
    confirmPassword: string;
    port: string;
  };
  api: {
    port: string;
  };
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:12566';

export default function SetupPage() {
  const [formData, setFormData] = useState<SetupFormData>({
    dbType: 'sqlite',
    sqlite: {
      filename: 'd-mail.db',
    },
    mysql: {
      host: 'localhost',
      port: '3306',
      dbname: 'd-main',
      username: '',
      password: '',
      charset: 'utf8',
    },
    manage: {
      username: 'admin',
      password: '',
      confirmPassword: '',
      port: '12566',
    },
    api: {
      port: '12567',
    },
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [testingConnection, setTestingConnection] = useState(false);
  const [checkingInit, setCheckingInit] = useState(true);
  const router = useRouter();
  const t = useTranslations('setup');
  const tCommon = useTranslations('common');

  // Check if system is already initialized on mount
  useEffect(() => {
    const checkInitStatus = async () => {
      try {
        const response = await fetch(`${API_URL}/set`, {
          credentials: 'include',
        });
        if (response.ok) {
          const responseData = await response.json();
          console.log('API Response:', responseData);

          // The response is wrapped in { code: 200, data: {...}, msg: "ok" }
          let systemInfo: SystemInfo;
          if (responseData.data && responseData.data.hasInit !== undefined) {
            systemInfo = responseData.data;
          } else if (responseData.hasInit !== undefined) {
            systemInfo = responseData;
          } else {
            // Fallback - assume not initialized
            setCheckingInit(false);
            return;
          }

          console.log('System info - hasInit:', systemInfo.hasInit);

          if (systemInfo.hasInit) {
            // Already initialized, redirect to root
            console.log('Already initialized, redirecting to root');
            router.push('/');
            return;
          }
        }
      } catch (err) {
        console.error('Failed to check init status:', err);
      } finally {
        setCheckingInit(false);
      }
    };
    checkInitStatus();
  }, [router]);

  // Load default settings on mount
  useEffect(() => {
    if (checkingInit) return;

    const loadDefaults = async () => {
      try {
        const response = await fetch(`${API_URL}/defaultSet`, {
          credentials: 'include',
        });
        if (response.ok) {
          const data: SetInfo = await response.json();
          if (data) {
            setFormData((prev) => ({
              ...prev,
              dbType: (data.dbType as 'sqlite' | 'mysql') || 'sqlite',
              sqlite: {
                filename: data.sqlite?.filename || 'd-mail.db',
              },
              mysql: {
                host: data.mysql?.host || 'localhost',
                port: String(data.mysql?.port || 3306),
                dbname: data.mysql?.dbname || 'd-main',
                username: data.mysql?.username || '',
                password: data.mysql?.password || '',
                charset: data.mysql?.charset || 'utf8',
              },
              manage: {
                ...prev.manage,
                port: String(data.manage?.port || 12566),
              },
              api: {
                port: String(data.api?.port || 12567),
              },
            }));
          }
        }
      } catch (err) {
        console.error('Failed to load defaults:', err);
      }
    };
    loadDefaults();
  }, [checkingInit]);

  const handleTestConnection = async () => {
    setTestingConnection(true);
    setError('');
    try {
      const testData: SetInfo = {
        dbType: formData.dbType,
        sqlite: formData.dbType === 'sqlite' ? formData.sqlite : undefined,
        mysql: formData.dbType === 'mysql' ? {
          ...formData.mysql,
          port: parseInt(formData.mysql.port),
        } : undefined,
      };
      const response = await fetch(`${API_URL}/testConnection`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(testData),
        credentials: 'include',
      });
      if (response.ok) {
        const result = await response.text();
        alert('Connection test successful: ' + result);
      } else {
        setError('Connection test failed');
      }
    } catch (err) {
      setError('Failed to test connection');
      console.error('Test connection error:', err);
    } finally {
      setTestingConnection(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    // Validate form
    if (!formData.manage.username.trim()) {
      setError('Username is required');
      setLoading(false);
      return;
    }

    if (!formData.manage.password || formData.manage.password.length < 6) {
      setError('Password must be at least 6 characters long');
      setLoading(false);
      return;
    }

    if (formData.manage.password !== formData.manage.confirmPassword) {
      setError('Passwords do not match');
      setLoading(false);
      return;
    }

    try {
      const submitData: SetInfo = {
        dbType: formData.dbType,
        sqlite: formData.dbType === 'sqlite' ? formData.sqlite : undefined,
        mysql: formData.dbType === 'mysql' ? {
          ...formData.mysql,
          port: parseInt(formData.mysql.port),
        } : undefined,
        manage: {
          username: formData.manage.username,
          password: formData.manage.password,
          port: parseInt(formData.manage.port),
        },
        api: {
          port: parseInt(formData.api.port),
        },
      };

      const response = await fetch(`${API_URL}/set`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(submitData),
        credentials: 'include',
      });

      if (response.ok) {
        // Setup successful, redirect to root
        router.push('/');
      } else {
        const errorData = await response.json().catch(() => null);
        setError(errorData?.message || 'Failed to initialize system');
      }
    } catch (err) {
      setError('Failed to connect to server');
      console.error('Setup error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    const parts = name.split('.');
    setFormData((prev) => {
      if (parts.length === 1) {
        return { ...prev, [parts[0]]: value };
      } else if (parts.length === 2) {
        return {
          ...prev,
          [parts[0]]: {
            ...(prev as any)[parts[0]],
            [parts[1]]: value,
          },
        };
      }
      return prev;
    });
  };

  const handleDbTypeChange = (value: 'sqlite' | 'mysql') => {
    setFormData((prev) => ({
      ...prev,
      dbType: value,
    }));
  };

  if (checkingInit) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">{tCommon('loading')}</div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="w-full max-w-2xl space-y-8">
        <div className="flex justify-end">
          <LanguageSwitcher />
        </div>
        <div>
          <h2 className="mt-6 text-center text-3xl font-bold tracking-tight text-gray-900">
            {t('systemSetup')}
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            {t('setupDesc')}
          </p>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>{t('initialConfig')}</CardTitle>
            <CardDescription>
              {t('initialConfigDesc')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-6" onSubmit={handleSubmit}>
              {error && (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              {/* Database Type */}
              <div className="space-y-2">
                <Label htmlFor="dbType">{t('dbType')}</Label>
                <Select
                  value={formData.dbType}
                  onValueChange={handleDbTypeChange}
                  disabled={loading}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder={t('selectDbType')} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="sqlite">SQLite</SelectItem>
                    <SelectItem value="mysql">MySQL</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* SQLite Configuration */}
              {formData.dbType === 'sqlite' && (
                <div className="space-y-4 border rounded-lg p-4 bg-gray-50">
                  <h3 className="text-sm font-medium">{t('sqliteConfig')}</h3>
                  <div className="space-y-2">
                    <Label htmlFor="sqlite.filename">{t('dbFilename')}</Label>
                    <Input
                      id="sqlite.filename"
                      name="sqlite.filename"
                      type="text"
                      value={formData.sqlite.filename}
                      onChange={handleChange}
                      disabled={loading}
                      placeholder="d-mail.db"
                    />
                  </div>
                </div>
              )}

              {/* MySQL Configuration */}
              {formData.dbType === 'mysql' && (
                <div className="space-y-4 border rounded-lg p-4 bg-gray-50">
                  <h3 className="text-sm font-medium">{t('mysqlConfig')}</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="mysql.host">{t('host')}</Label>
                      <Input
                        id="mysql.host"
                        name="mysql.host"
                        type="text"
                        value={formData.mysql.host}
                        onChange={handleChange}
                        disabled={loading}
                        placeholder="localhost"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="mysql.port">{t('port')}</Label>
                      <Input
                        id="mysql.port"
                        name="mysql.port"
                        type="number"
                        value={formData.mysql.port}
                        onChange={handleChange}
                        disabled={loading}
                        placeholder="3306"
                      />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="mysql.dbname">{t('dbName')}</Label>
                    <Input
                      id="mysql.dbname"
                      name="mysql.dbname"
                      type="text"
                      value={formData.mysql.dbname}
                      onChange={handleChange}
                      disabled={loading}
                      placeholder="d-main"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="mysql.username">{t('username')}</Label>
                      <Input
                        id="mysql.username"
                        name="mysql.username"
                        type="text"
                        value={formData.mysql.username}
                        onChange={handleChange}
                        disabled={loading}
                        placeholder={t('dbUsernamePlaceholder')}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="mysql.password">{t('password')}</Label>
                      <Input
                        id="mysql.password"
                        name="mysql.password"
                        type="password"
                        value={formData.mysql.password}
                        onChange={handleChange}
                        disabled={loading}
                        placeholder={t('dbPasswordPlaceholder')}
                      />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="mysql.charset">{t('charset')}</Label>
                    <Input
                      id="mysql.charset"
                      name="mysql.charset"
                      type="text"
                      value={formData.mysql.charset}
                      onChange={handleChange}
                      disabled={loading}
                      placeholder="utf8"
                    />
                  </div>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={handleTestConnection}
                    disabled={loading || testingConnection}
                  >
                    {testingConnection ? tCommon('testing') : tCommon('testConnection')}
                  </Button>
                </div>
              )}

              <hr className="my-6" />

              {/* Port Configuration */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="manage.port">{t('managePort')} (12566)</Label>
                  <Input
                    id="manage.port"
                    name="manage.port"
                    type="number"
                    value={formData.manage.port}
                    onChange={handleChange}
                    disabled={loading}
                    placeholder="12566"
                  />
                  <p className="text-xs text-gray-500">
                    {t('managePortDesc')}
                  </p>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="api.port">{t('apiPort')} (12567)</Label>
                  <Input
                    id="api.port"
                    name="api.port"
                    type="number"
                    value={formData.api.port}
                    onChange={handleChange}
                    disabled={loading}
                    placeholder="12567"
                  />
                  <p className="text-xs text-gray-500">
                    {t('apiPortDesc')}
                  </p>
                </div>
              </div>

              <hr className="my-6" />

              {/* Admin Account */}
              <div className="space-y-4">
                <h3 className="text-lg font-medium">{t('adminAccount')}</h3>
                <div className="space-y-2">
                  <Label htmlFor="manage.username">{t('username')}</Label>
                  <Input
                    id="manage.username"
                    name="manage.username"
                    type="text"
                    value={formData.manage.username}
                    onChange={handleChange}
                    required
                    disabled={loading}
                    placeholder={t('adminUsernamePlaceholder')}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="manage.password">{t('password')}</Label>
                  <Input
                    id="manage.password"
                    name="manage.password"
                    type="password"
                    value={formData.manage.password}
                    onChange={handleChange}
                    required
                    disabled={loading}
                    placeholder={t('passwordPlaceholder')}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="manage.confirmPassword">{t('confirmPassword')}</Label>
                  <Input
                    id="manage.confirmPassword"
                    name="manage.confirmPassword"
                    type="password"
                    value={formData.manage.confirmPassword}
                    onChange={handleChange}
                    required
                    disabled={loading}
                    placeholder={t('confirmPasswordPlaceholder')}
                  />
                </div>
              </div>

              <Button
                type="submit"
                className="w-full"
                disabled={loading}
              >
                {loading ? t('settingUp') : t('completeSetup')}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
