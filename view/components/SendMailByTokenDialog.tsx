'use client';

import { useState } from 'react';
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
        setError('Please enter email content');
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
          <DialogTitle>Send Mail</DialogTitle>
          <DialogDescription>
            Send an email using the selected API token
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
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
          <div className="space-y-2">
            <Label htmlFor="recipients">Recipients (optional)</Label>
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
