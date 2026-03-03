'use client';

import { useState, useEffect } from 'react';
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
import { TokenConfig } from '@/types/token';
import { SMTPConfig } from '@/types/smtp';
import { MailConfig } from '@/types/mail';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';

interface TokenFormDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  editingToken?: TokenConfig | null;
  smtpServers: SMTPConfig[];
  mails: MailConfig[];
  onSuccess?: () => void;
  triggerButton?: boolean;
}

const generateToken = () => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < 32; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
};

export function TokenFormDialog({
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
  editingToken: externalEditingToken,
  smtpServers,
  mails,
  onSuccess,
  triggerButton = true,
}: TokenFormDialogProps) {
  const router = useRouter();
  const [internalOpen, setInternalOpen] = useState(false);
  const [editingToken, setEditingToken] = useState<TokenConfig | null>(null);
  const [formData, setFormData] = useState<TokenConfig>({
    token: generateToken(),
    subject: '',
    receiveEmailIds: '',
    SMTPId: smtpServers.length > 0 ? smtpServers[0].id! : 0,
    isUse: true,
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  // Use controlled state if provided, otherwise use internal state
  const open = controlledOpen !== undefined ? controlledOpen : internalOpen;
  const setOpen = controlledOnOpenChange || setInternalOpen;

  // Sync external editingToken
  useEffect(() => {
    if (externalEditingToken !== undefined && externalEditingToken !== null) {
      setEditingToken(externalEditingToken);
      setFormData({
        token: externalEditingToken.token || '',
        subject: externalEditingToken.subject || '',
        receiveEmailIds: externalEditingToken.receiveEmailIds || '',
        SMTPId: externalEditingToken.SMTPId || 0,
        isUse: externalEditingToken.isUse ?? true,
      });
    }
  }, [externalEditingToken]);

  // Reset form when dialog opens for new token
  useEffect(() => {
    if (open && !externalEditingToken) {
      setFormData({
        token: generateToken(),
        subject: '',
        receiveEmailIds: '',
        SMTPId: smtpServers.length > 0 ? smtpServers[0].id! : 0,
        isUse: true,
      });
      setEditingToken(null);
    }
  }, [open, externalEditingToken, smtpServers]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    if (name === 'isUse') {
      setFormData(prev => ({
        ...prev,
        [name]: (e.target as HTMLInputElement).checked,
      }));
    } else if (name === 'SMTPId') {
      setFormData(prev => ({
        ...prev,
        [name]: parseInt(value) || 0,
      }));
    } else {
      setFormData(prev => ({
        ...prev,
        [name]: value,
      }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError('');

    try {
      if (!formData.token.trim()) {
        setError('Token is required');
        return;
      }

      if (formData.SMTPId === 0) {
        setError('Please select an SMTP server');
        return;
      }

      if (editingToken) {
        await apiClient.updateToken(formData);
      } else {
        await apiClient.createToken(formData);
      }

      setOpen(false);
      setEditingToken(null);
      onSuccess?.();
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setError('Failed to save token');
        console.error('Failed to save token:', err);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleCancel = () => {
    setOpen(false);
    setEditingToken(null);
    setError('');
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {triggerButton && (
        <DialogTrigger asChild>
          <Button className="gap-2">
            <Plus className="h-4 w-4" />
            Generate Token
          </Button>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle>{editingToken ? 'Edit Token' : 'Generate New Token'}</DialogTitle>
          <DialogDescription>
            {editingToken ? 'Modify token settings' : 'Create a new API token for email sending'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="space-y-2">
            <Label htmlFor="token">Token *</Label>
            <div className="flex gap-2">
              <Input
                id="token"
                name="token"
                value={formData.token}
                onChange={handleInputChange}
                className="flex-1"
                readOnly={!editingToken}
              />
              {!editingToken && (
                <Button type="button" variant="outline" onClick={() => setFormData(prev => ({ ...prev, token: generateToken() }))}>
                  Regenerate
                </Button>
              )}
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="subject">Subject</Label>
            <Input
              id="subject"
              name="subject"
              value={formData.subject || ''}
              onChange={handleInputChange}
              placeholder="Default email subject"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="SMTPId">SMTP Server *</Label>
            <select
              id="SMTPId"
              name="SMTPId"
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              value={formData.SMTPId || ''}
              onChange={handleInputChange}
            >
              <option value="">Select SMTP Server</option>
              {smtpServers.map(server => (
                <option key={server.id} value={server.id}>
                  {server.name || server.host} ({server.host}:{server.port})
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="receiveEmailIds">Recipient Emails</Label>
            <Input
              id="receiveEmailIds"
              name="receiveEmailIds"
              value={formData.receiveEmailIds || ''}
              onChange={handleInputChange}
              placeholder="Email IDs, comma separated (e.g., 1,2,3)"
            />
            {mails.length > 0 && (
              <p className="text-xs text-gray-500">
                Available: {mails.map(m => `${m.id}:${m.name || m.mail}`).join(', ')}
              </p>
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
            <Label htmlFor="isUse">Active</Label>
          </div>
          <DialogFooter>
            <Button type="button" variant="secondary" onClick={handleCancel} disabled={submitting}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? 'Saving...' : 'Save Token'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}