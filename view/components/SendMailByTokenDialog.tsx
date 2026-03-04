'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Send, Plus } from 'lucide-react';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { EmailSelectDialog } from '@/components/EmailSelectDialog';
import { MailConfig } from '@/types/mail';

interface SendMailByTokenDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  token: string;
  tokenSubject: string;
  mails: MailConfig[];
  onSuccess?: () => void;
}

export function SendMailByTokenDialog({
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
  token,
  tokenSubject,
  mails,
  onSuccess,
}: SendMailByTokenDialogProps) {
  const router = useRouter();
  const t = useTranslations('sendMail');
  const tCommon = useTranslations('common');
  const tAuth = useTranslations('auth');
  const [internalOpen, setInternalOpen] = useState(false);
  const [formData, setFormData] = useState({
    subject: tokenSubject,
    content: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [emailDialogOpen, setEmailDialogOpen] = useState(false);
  const [selectedEmailIds, setSelectedEmailIds] = useState<number[]>([]);
  const [selectedEmails, setSelectedEmails] = useState<MailConfig[]>([]);

  // Use controlled state if provided, otherwise use internal state
  const open = controlledOpen !== undefined ? controlledOpen : internalOpen;
  const setOpen = controlledOnOpenChange || setInternalOpen;

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
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
      if (!formData.content.trim()) {
        setError(t('emailContentPlaceholder'));
        return;
      }

      const recipients = selectedEmails.length > 0 ? selectedEmails.map(email => email.mail) : undefined;
      const result = await apiClient.sendMailByToken(token, formData.subject, formData.content, recipients);

      if (result.code === 200) {
        setOpen(false);
        setFormData({
          subject: tokenSubject,
          content: '',
        });
        setSelectedEmailIds([]);
        setSelectedEmails([]);
        onSuccess?.();
      } else {
        setError(result.message || t('sendError'));
      }
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert(tAuth('authFailed'));
        router.push('/');
      } else {
        setError(t('sendError'));
        console.error('Failed to send mail:', err);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleCancel = () => {
    setOpen(false);
    setFormData({
      subject: tokenSubject,
      content: '',
    });
    setError('');
    setSelectedEmailIds([]);
    setSelectedEmails([]);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>{t('sendMail')}</DialogTitle>
          <DialogDescription>
            {t('sendMailDesc')}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="space-y-2">
            <Label htmlFor="subject">{t('emailSubject')}</Label>
            <Input
              id="subject"
              name="subject"
              value={formData.subject}
              onChange={handleInputChange}
              placeholder={t('emailSubject')}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="content">{t('emailContent')} *</Label>
            <Textarea
              id="content"
              name="content"
              value={formData.content}
              onChange={handleInputChange}
              placeholder={t('emailContentPlaceholder')}
              rows={5}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="recipients">{t('to')}</Label>
            <div className="flex flex-wrap gap-2 mb-2">
              {selectedEmails.map(email => (
                <div key={email.id} className="flex items-center gap-1 bg-muted px-2 py-1 rounded-md">
                  <span className="text-sm">{email.name || email.mail}</span>
                  <button
                    type="button"
                    className="text-muted-foreground hover:text-foreground"
                    onClick={() => {
                      const updatedSelectedEmails = selectedEmails.filter(e => e.id !== email.id);
                      setSelectedEmails(updatedSelectedEmails);
                      setSelectedEmailIds(updatedSelectedEmails.map(e => e.id!));
                    }}
                  >
                    ×
                  </button>
                </div>
              ))}
              <Button type="button" variant="outline" onClick={() => setEmailDialogOpen(true)}>
                <Plus className="h-4 w-4" />
                {tCommon('add')}
              </Button>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="secondary" onClick={handleCancel} disabled={submitting}>
              {tCommon('cancel')}
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? t('sending') : t('sendMail')}
            </Button>

            <EmailSelectDialog
              open={emailDialogOpen}
              onOpenChange={setEmailDialogOpen}
              emails={mails}
              selectedEmailIds={selectedEmailIds}
              onSelectedEmailIdsChange={(ids) => {
                setSelectedEmailIds(ids);
                const emails = mails.filter(email => ids.includes(email.id!));
                setSelectedEmails(emails);
              }}
            />
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
