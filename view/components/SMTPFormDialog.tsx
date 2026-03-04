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
import { SMTPConfig } from '@/types/smtp';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';

interface SMTPFormDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  editingServer?: SMTPConfig | null;
  onSuccess?: () => void;
  triggerButton?: boolean;
}

export function SMTPFormDialog({
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
  editingServer: externalEditingServer,
  onSuccess,
  triggerButton = true,
}: SMTPFormDialogProps) {
  const router = useRouter();
  const [internalOpen, setInternalOpen] = useState(false);
  const [editingServer, setEditingServer] = useState<SMTPConfig | null>(null);
  const [formData, setFormData] = useState<SMTPConfig>({
    name: '',
    host: '',
    port: 587,
    mail: '',
    username: '',
    password: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [testing, setTesting] = useState(false);
  const [error, setError] = useState('');
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);

  // Use controlled state if provided, otherwise use internal state
  const open = controlledOpen !== undefined ? controlledOpen : internalOpen;
  const setOpen = controlledOnOpenChange || setInternalOpen;

  // Sync external editingServer
  useEffect(() => {
    if (externalEditingServer !== undefined && externalEditingServer !== null) {
      setEditingServer(externalEditingServer);
      setFormData({
        name: externalEditingServer.name || '',
        host: externalEditingServer.host || '',
        port: externalEditingServer.port || 587,
        mail: externalEditingServer.mail || '',
        username: externalEditingServer.username || '',
        password: externalEditingServer.password || '',
      });
    }
  }, [externalEditingServer]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'number' ? parseInt(value) || 0 : value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError('');

    try {
      if (!formData.host.trim() || !formData.mail.trim()) {
        setError('Please fill in all required fields');
        return;
      }

      if (formData.port <= 0 || formData.port > 65535) {
        setError('Port must be between 1 and 65535');
        return;
      }

      if (!editingServer && (!formData.username.trim() || !formData.password || !formData.password.trim())) {
        setError('Username and password are required for new servers');
        return;
      }

      if (editingServer) {
        await apiClient.updateSMTPServer(formData);
      } else {
        await apiClient.createSMTPServer(formData);
      }

      setOpen(false);
      setEditingServer(null);
      setFormData({
        name: '',
        host: '',
        port: 587,
        mail: '',
        username: '',
        password: '',
      });
      onSuccess?.();
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setError('Failed to save SMTP server');
        console.error('Failed to save SMTP server:', err);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleCancel = () => {
    setOpen(false);
    setEditingServer(null);
    setFormData({
      name: '',
      host: '',
      port: 587,
      mail: '',
      username: '',
      password: '',
    });
    setError('');
    setTestResult(null);
  };

  const handleTest = async () => {
    setTesting(true);
    setTestResult(null);
    setError('');

    try {
      if (!formData.host.trim() || !formData.mail.trim()) {
        setError('Please fill in all required fields');
        setTesting(false);
        return;
      }

      if (formData.port <= 0 || formData.port > 65535) {
        setError('Port must be between 1 and 65535');
        setTesting(false);
        return;
      }

      const result = await apiClient.testSMTPServer(formData);
      if (result.code === 200) {
        setTestResult({ success: true, message: result.message || 'Test successful' });
      } else {
        setTestResult({ success: false, message: result.message || 'Test failed' });
      }
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setTestResult({ success: false, message: 'Failed to test SMTP server' });
        console.error('Failed to test SMTP server:', err);
      }
    } finally {
      setTesting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {triggerButton && (
        <DialogTrigger asChild>
          <Button className="gap-2">
            <Plus className="h-4 w-4" />
            Add Server
          </Button>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle>{editingServer ? 'Edit SMTP Server' : 'Add SMTP Server'}</DialogTitle>
          <DialogDescription>
            {editingServer
              ? 'Modify the settings of an existing SMTP server.'
              : 'Configure a new SMTP server for email delivery.'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          {testResult && (
            <Alert variant={testResult.success ? "default" : "destructive"}>
              <AlertDescription>{testResult.message}</AlertDescription>
            </Alert>
          )}
          <div className="space-y-2">
            <Label htmlFor="name">Server Name</Label>
            <Input
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              placeholder="e.g., Gmail, Outlook"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="host">SMTP Host *</Label>
            <Input
              id="host"
              name="host"
              value={formData.host}
              onChange={handleInputChange}
              placeholder="smtp.gmail.com"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="port">Port *</Label>
            <Input
              id="port"
              name="port"
              type="number"
              value={formData.port}
              onChange={handleInputChange}
              min="1"
              max="65535"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="mail">From Address *</Label>
            <Input
              id="mail"
              name="mail"
              value={formData.mail}
              onChange={handleInputChange}
              placeholder="your-email@example.com"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="username">Username</Label>
            <Input
              id="username"
              name="username"
              value={formData.username}
              onChange={handleInputChange}
              placeholder="your-email@example.com"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              name="password"
              type="password"
              value={formData.password}
              onChange={handleInputChange}
              placeholder="********"
              autoComplete="new-password"
            />
          </div>
          <DialogFooter className="flex justify-between">
            <Button type="button" variant="outline" onClick={handleTest} disabled={submitting || testing}>
              {testing ? 'Testing...' : 'Test Connection'}
            </Button>
            <div className="flex gap-2">
              <Button type="button" variant="secondary" onClick={handleCancel} disabled={submitting || testing}>
                Cancel
              </Button>
              <Button type="submit" disabled={submitting || testing}>
                {submitting ? 'Saving...' : 'Save Server'}
              </Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}