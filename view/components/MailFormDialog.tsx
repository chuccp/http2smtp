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
import { Plus } from 'lucide-react';
import { MailConfig } from '@/types/mail';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';

interface MailFormDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  editingMail?: MailConfig | null;
  onSuccess?: () => void;
  triggerButton?: boolean;
}

export function MailFormDialog({
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
  editingMail: externalEditingMail,
  onSuccess,
  triggerButton = true,
}: MailFormDialogProps) {
  const router = useRouter();
  const t = useTranslations('mail');
  const tCommon = useTranslations('common');
  const tAuth = useTranslations('auth');
  const [internalOpen, setInternalOpen] = useState(false);
  const [editingMail, setEditingMail] = useState<MailConfig | null>(null);
  const [formData, setFormData] = useState<MailConfig>({
    name: '',
    mail: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  // Use controlled state if provided, otherwise use internal state
  const open = controlledOpen !== undefined ? controlledOpen : internalOpen;
  const setOpen = controlledOnOpenChange || setInternalOpen;

  // Sync external editingMail
  useEffect(() => {
    if (externalEditingMail !== undefined && externalEditingMail !== null) {
      setEditingMail(externalEditingMail);
      setFormData({
        id: externalEditingMail.id,
        name: externalEditingMail.name || '',
        mail: externalEditingMail.mail || '',
      });
    }
  }, [externalEditingMail]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError('');

    try {
      if (!formData.mail.trim()) {
        setError(t('enterMail'));
        return;
      }

      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formData.mail)) {
        setError(t('invalidMail'));
        return;
      }

      if (editingMail) {
        await apiClient.updateMail(formData);
      } else {
        await apiClient.createMail(formData);
      }

      setOpen(false);
      setEditingMail(null);
      setFormData({
        name: '',
        mail: '',
      });
      onSuccess?.();
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert(tAuth('authFailed'));
        router.push('/');
      } else {
        setError(t('saveError'));
        console.error('Failed to save mail address:', err);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleCancel = () => {
    setOpen(false);
    setEditingMail(null);
    setFormData({
      name: '',
      mail: '',
    });
    setError('');
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {triggerButton && (
        <DialogTrigger asChild>
          <Button className="gap-2">
            <Plus className="h-4 w-4" />
            {t('addAddress')}
          </Button>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-[450px]">
        <DialogHeader>
          <DialogTitle>{editingMail ? t('editAddress') : t('addAddressTitle')}</DialogTitle>
          <DialogDescription>
            {editingMail ? t('editAddressDesc') : t('addAddressDesc')}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="space-y-2">
            <Label htmlFor="name">{tCommon('name')}</Label>
            <Input
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              placeholder={t('namePlaceholder')}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="mail">{t('mailAddress')} *</Label>
            <Input
              id="mail"
              name="mail"
              type="email"
              value={formData.mail}
              onChange={handleInputChange}
              placeholder={t('mailPlaceholder')}
              required
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="secondary" onClick={handleCancel} disabled={submitting}>
              {t('cancel')}
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? t('saving') : tCommon('save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}