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
        setError('Please enter a mail address');
        return;
      }

      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formData.mail)) {
        setError('Please enter a valid mail address');
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
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setError('Failed to save mail address');
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
            Add Address
          </Button>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-[450px]">
        <DialogHeader>
          <DialogTitle>{editingMail ? 'Edit Mail Address' : 'Add Mail Address'}</DialogTitle>
          <DialogDescription>
            {editingMail ? 'Modify recipient mail address information' : 'Add a new recipient mail address'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              placeholder="e.g., John, Support"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="mail">Mail Address *</Label>
            <Input
              id="mail"
              name="mail"
              type="email"
              value={formData.mail}
              onChange={handleInputChange}
              placeholder="example@email.com"
              required
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="secondary" onClick={handleCancel} disabled={submitting}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}