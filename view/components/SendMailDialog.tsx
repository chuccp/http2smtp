'use client';

import { useState } from 'react';
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
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Send, Plus } from 'lucide-react';
import { SendMail } from '@/types/entity';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { EmailSelectDialog } from '@/components/EmailSelectDialog';
import { MailConfig } from '@/types/mail';
import { SMTPConfig } from '@/types/smtp';

interface SendMailDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  smtpServers: SMTPConfig[];
  mails: MailConfig[];
  onSuccess?: () => void;
  triggerButton?: boolean;
}

export function SendMailDialog({
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
  smtpServers,
  mails,
  onSuccess,
  triggerButton = true,
}: SendMailDialogProps) {
  const router = useRouter();
  const [internalOpen, setInternalOpen] = useState(false);
  const [formData, setFormData] = useState<SendMail>({
    SMTPId: smtpServers.length > 0 ? smtpServers[0].id! : 0,
    recipients: [],
    subject: '',
    content: '',
  });

  // 如果只有一个 SMTP 服务器，默认选中它
  if (smtpServers.length === 1 && formData.SMTPId === 0) {
    setFormData(prev => ({
      ...prev,
      SMTPId: smtpServers[0].id!,
    }));
  }
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [emailDialogOpen, setEmailDialogOpen] = useState(false);
  const [selectedEmailIds, setSelectedEmailIds] = useState<number[]>([]);
  const [selectedEmails, setSelectedEmails] = useState<MailConfig[]>([]);

  // Use controlled state if provided, otherwise use internal state
  const open = controlledOpen !== undefined ? controlledOpen : internalOpen;
  const setOpen = controlledOnOpenChange || setInternalOpen;

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
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
      if (formData.SMTPId === 0) {
        setError('Please select an SMTP server');
        return;
      }

      if (selectedEmails.length === 0) {
        setError('Please add at least one recipient');
        return;
      }

      const result = await apiClient.sendMailBySMTP({
        ...formData,
        recipients: selectedEmails.map(email => email.mail),
      });
      if (result.code === 200) {
        setOpen(false);
        setFormData({
          SMTPId: smtpServers.length > 0 ? smtpServers[0].id! : 0,
          recipients: [],
          subject: '',
          content: '',
        });
        onSuccess?.();
      } else {
        setError(result.message || 'Failed to send mail');
      }
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setError('Failed to send mail');
        console.error('Failed to send mail:', err);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleCancel = () => {
    setOpen(false);
    setFormData({
      SMTPId: smtpServers.length > 0 ? smtpServers[0].id! : 0,
      recipients: [],
      subject: '',
      content: '',
    });
    setError('');
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {triggerButton && (
        <DialogTrigger asChild>
          <Button className="gap-2">
            <Send className="h-4 w-4" />
            Send Mail
          </Button>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Send Mail</DialogTitle>
          <DialogDescription>
            Send an email using your configured SMTP server
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          {/* 隐藏 SMTP 服务器选择，因为我们已经在表格中选择了特定的 SMTP 服务器 */}
          <input type="hidden" id="SMTPId" name="SMTPId" value={formData.SMTPId} />
          <div className="space-y-2">
            <Label htmlFor="recipients">Recipients *</Label>
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
                Add
              </Button>
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="subject">Subject</Label>
            <Input
              id="subject"
              name="subject"
              value={formData.subject}
              onChange={handleInputChange}
              placeholder="Email subject"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="content">Content *</Label>
            <Textarea
              id="content"
              name="content"
              value={formData.content}
              onChange={handleInputChange}
              placeholder="Email content"
              rows={5}
              required
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="secondary" onClick={handleCancel} disabled={submitting}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? 'Sending...' : 'Send Mail'}
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