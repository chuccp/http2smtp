'use client';

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Plus, Trash2 } from 'lucide-react';
import { ScheduleConfig, ScheduleHeader } from '@/types/schedule';
import { TokenConfig } from '@/types/token';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { TokenSelectDialog } from '@/components/TokenSelectDialog';

interface ScheduleFormDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  editingSchedule?: ScheduleConfig | null;
  tokens: TokenConfig[];
  onSuccess?: () => void;
  triggerButton?: boolean;
}

export function ScheduleFormDialog({
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
  editingSchedule: externalEditingSchedule,
  tokens,
  onSuccess,
  triggerButton = true,
}: ScheduleFormDialogProps) {
  const router = useRouter();
  const t = useTranslations('schedule');
  const tCommon = useTranslations('common');
  const tAuth = useTranslations('auth');
  const [internalOpen, setInternalOpen] = useState(false);
  const [editingSchedule, setEditingSchedule] = useState<ScheduleConfig | null>(null);
  const [formData, setFormData] = useState<ScheduleConfig>({
    name: '',
    token: tokens.length > 0 ? tokens[0].token || '' : '',
    cron: '0 0 * * *',
    url: '',
    method: 'GET',
    headers: [],
    body: '',
    useTemplate: false,
    template: '',
    isUse: true,
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [tokenDialogOpen, setTokenDialogOpen] = useState(false);
  const [selectedTokenId, setSelectedTokenId] = useState<number | null>(null);
  const [selectedToken, setSelectedToken] = useState<TokenConfig | null>(null);

  // Use controlled state if provided, otherwise use internal state
  const open = controlledOpen !== undefined ? controlledOpen : internalOpen;
  const setOpen = controlledOnOpenChange || setInternalOpen;

  // Sync external editingSchedule
  useEffect(() => {
    if (externalEditingSchedule !== undefined && externalEditingSchedule !== null) {
      setEditingSchedule(externalEditingSchedule);
      // Parse headers from headerStr if available
      let headers: ScheduleHeader[] = [];
      if (externalEditingSchedule.headerStr) {
        try {
          headers = JSON.parse(externalEditingSchedule.headerStr);
        } catch {
          headers = [];
        }
      }
      // Find the token by token string to get the id
      const token = tokens.find(t => t.token === externalEditingSchedule.token);
      setSelectedTokenId(token?.id || null);
      setSelectedToken(token || null);
      setFormData({
        name: externalEditingSchedule.name || '',
        token: externalEditingSchedule.token || '',
        cron: externalEditingSchedule.cron || '0 0 * * *',
        url: externalEditingSchedule.url || '',
        method: externalEditingSchedule.method || 'GET',
        headers: headers,
        body: externalEditingSchedule.body || '',
        useTemplate: externalEditingSchedule.useTemplate ?? false,
        template: externalEditingSchedule.template || '',
        isUse: externalEditingSchedule.isUse ?? true,
      });
    }
  }, [externalEditingSchedule, tokens]);

  // Reset form when dialog opens for new schedule
  useEffect(() => {
    if (open && !externalEditingSchedule) {
      const defaultToken = tokens.length > 0 ? tokens[0] : null;
      setSelectedTokenId(defaultToken?.id || null);
      setSelectedToken(defaultToken || null);
      setFormData({
        name: '',
        token: defaultToken?.token || '',
        cron: '0 0 * * *',
        url: '',
        method: 'GET',
        headers: [],
        body: '',
        useTemplate: false,
        template: '',
        isUse: true,
      });
      setEditingSchedule(null);
    }
  }, [open, externalEditingSchedule, tokens]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target;
    if (name === 'useTemplate' || name === 'isUse') {
      setFormData(prev => ({
        ...prev,
        [name]: (e.target as HTMLInputElement).checked,
      }));
    } else {
      setFormData(prev => ({
        ...prev,
        [name]: value,
      }));
    }
  };

  const handleAddHeader = () => {
    setFormData(prev => ({
      ...prev,
      headers: [...(prev.headers || []), { name: '', value: '' }],
    }));
  };

  const handleRemoveHeader = (index: number) => {
    setFormData(prev => ({
      ...prev,
      headers: prev.headers?.filter((_, i) => i !== index),
    }));
  };

  const handleHeaderChange = (index: number, field: 'name' | 'value', value: string) => {
    setFormData(prev => ({
      ...prev,
      headers: prev.headers?.map((h, i) => (i === index ? { ...h, [field]: value } : h)),
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError('');

    try {
      if (!formData.name.trim()) {
        setError(t('nameRequired'));
        return;
      }

      if (!formData.token.trim()) {
        setError(t('tokenRequired'));
        return;
      }

      if (!formData.url.trim()) {
        setError(t('urlRequired'));
        return;
      }

      if (!formData.cron.trim()) {
        setError(t('cronRequired'));
        return;
      }

      // Prepare data for submission
      // 后端期望的是 headers 数组，而不是 headerStr 字符串
      const submitData: ScheduleConfig = {
        ...formData,
        headers: formData.headers || [],
      };

      // 删除 headerStr，因为后端会自动处理
      delete submitData.headerStr;

      if (editingSchedule) {
        submitData.id = editingSchedule.id;
        await apiClient.updateSchedule(submitData);
      } else {
        await apiClient.createSchedule(submitData);
      }

      setOpen(false);
      setEditingSchedule(null);
      onSuccess?.();
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert(tAuth('authFailed'));
        router.push('/');
      } else {
        setError(t('saveError'));
        console.error('Failed to save schedule:', err);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleCancel = () => {
    setOpen(false);
    setEditingSchedule(null);
    setError('');
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {triggerButton && (
        <DialogTrigger asChild>
          <Button className="gap-2">
            <Plus className="h-4 w-4" />
            {t('addSchedule')}
          </Button>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-[650px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{editingSchedule ? t('editSchedule') : t('addScheduleTitle')}</DialogTitle>
          <DialogDescription>
            {editingSchedule ? t('editScheduleDesc') : t('addScheduleDesc')}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="name">{t('scheduleName')} *</Label>
              <Input
                id="name"
                name="name"
                value={formData.name}
                onChange={handleInputChange}
                placeholder={t('scheduleNamePlaceholder')}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="token">{t('token')} *</Label>
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  className="flex-1 justify-start text-left"
                  onClick={() => setTokenDialogOpen(true)}
                >
                  {selectedToken ? (
                    <span>{selectedToken.subject || selectedToken.token?.substring(0, 12)}...</span>
                  ) : (
                    <span>{t('selectToken')}</span>
                  )}
                </Button>
                {selectedToken && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => {
                      setSelectedTokenId(null);
                      setSelectedToken(null);
                      setFormData(prev => ({ ...prev, token: '' }));
                    }}
                  >
                    ×
                  </Button>
                )}
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="cron">{t('cron')} *</Label>
              <Input
                id="cron"
                name="cron"
                value={formData.cron}
                onChange={handleInputChange}
                placeholder={t('cronPlaceholder')}
              />
              <p className="text-xs text-gray-500" dangerouslySetInnerHTML={{__html: t('cronExample')}} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="method">{t('method')}</Label>
              <select
                id="method"
                name="method"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                value={formData.method}
                onChange={handleInputChange}
              >
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="DELETE">DELETE</option>
              </select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="url">{t('url')} *</Label>
            <Input
              id="url"
              name="url"
              value={formData.url}
              onChange={handleInputChange}
              placeholder={t('urlPlaceholder')}
            />
          </div>

          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <Label>{t('headers')}</Label>
              <Button type="button" variant="outline" size="sm" onClick={handleAddHeader}>
                {t('addHeader')}
              </Button>
            </div>
            {formData.headers && formData.headers.length > 0 && (
              <div className="space-y-2">
                {formData.headers.map((header, index) => (
                  <div key={index} className="flex gap-2 items-center">
                    <Input
                      placeholder={t('headerNamePlaceholder')}
                      value={header.name}
                      onChange={(e) => handleHeaderChange(index, 'name', e.target.value)}
                      className="flex-1"
                    />
                    <Input
                      placeholder={t('headerValuePlaceholder')}
                      value={header.value}
                      onChange={(e) => handleHeaderChange(index, 'value', e.target.value)}
                      className="flex-1"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() => handleRemoveHeader(index)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="body">{t('requestBody')}</Label>
            <textarea
              id="body"
              name="body"
              className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              value={formData.body || ''}
              onChange={handleInputChange}
              placeholder='{"key": "value"}'
            />
          </div>

          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="useTemplate"
                name="useTemplate"
                checked={formData.useTemplate}
                onChange={handleInputChange}
                className="h-4 w-4"
              />
              <Label htmlFor="useTemplate">{t('useTemplate')}</Label>
            </div>

            {formData.useTemplate && (
              <div className="space-y-2">
                <Label htmlFor="template">{t('emailTemplate')}</Label>
                <textarea
                  id="template"
                  name="template"
                  className="flex min-h-[100px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  value={formData.template || ''}
                  onChange={handleInputChange}
                  placeholder="Use {{.field}} for template variables"
                />
              </div>
            )}
          </div>

          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="isUse"
              name="isUse"
              checked={formData.isUse}
              onChange={handleInputChange}
              className="h-4 w-4"
            />
            <Label htmlFor="isUse">{tCommon('active')}</Label>
          </div>

          <DialogFooter>
            <Button type="button" variant="secondary" onClick={handleCancel} disabled={submitting}>
              {tCommon('cancel')}
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? tCommon('saving') : t('saveSchedule')}
            </Button>

            <TokenSelectDialog
              open={tokenDialogOpen}
              onOpenChange={setTokenDialogOpen}
              tokens={tokens}
              selectedTokenId={selectedTokenId}
              onSelectedTokenIdChange={(id) => {
                setSelectedTokenId(id);
                const token = tokens.find(t => t.id === id);
                setSelectedToken(token || null);
                setFormData(prev => ({ ...prev, token: token?.token || '' }));
              }}
            />
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}