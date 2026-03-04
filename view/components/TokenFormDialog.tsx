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
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Pagination } from '@/components/ui/pagination';
import { EmailSelectDialog } from '@/components/EmailSelectDialog';
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
  const [smtpDialogOpen, setSmtpDialogOpen] = useState(false);
  const [emailDialogOpen, setEmailDialogOpen] = useState(false);
  const [selectedEmailIds, setSelectedEmailIds] = useState<number[]>([]);
  const [selectedEmails, setSelectedEmails] = useState<MailConfig[]>([]);

  // Pagination state
  const [smtpPage, setSmtpPage] = useState(1);
  const pageSize = 10;

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
      // Parse and set selected emails
      const emailIds = externalEditingToken.receiveEmailIds
        ? externalEditingToken.receiveEmailIds.split(',').map(id => parseInt(id.trim())).filter(id => !isNaN(id))
        : [];
      setSelectedEmailIds(emailIds);
      const emails = mails.filter(mail => emailIds.includes(mail.id!));
      setSelectedEmails(emails);
    }
  }, [externalEditingToken, mails]);

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
      setSelectedEmailIds([]);
      setSelectedEmails([]);
    }
  }, [open, externalEditingToken, smtpServers]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
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
    <>
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
            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                className="flex-1 justify-start text-left"
                onClick={() => setSmtpDialogOpen(true)}
              >
                {smtpServers.find(server => server.id === formData.SMTPId) ? (
                  <span>{smtpServers.find(server => server.id === formData.SMTPId)?.name || smtpServers.find(server => server.id === formData.SMTPId)?.host} ({smtpServers.find(server => server.id === formData.SMTPId)?.host}:{smtpServers.find(server => server.id === formData.SMTPId)?.port})</span>
                ) : (
                  <span>Select SMTP Server</span>
                )}
              </Button>
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="receiveEmailIds">Recipient Emails</Label>
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
                      setFormData(prev => ({
                        ...prev,
                        receiveEmailIds: updatedSelectedEmails.map(e => e.id).join(','),
                      }));
                    }}
                  >
                    ×
                  </button>
                </div>
              ))}
              <button
                type="button"
                className="flex items-center gap-1 bg-background border border-input px-2 py-1 rounded-md hover:bg-accent"
                onClick={() => setEmailDialogOpen(true)}
              >
                <Plus className="h-4 w-4" />
                Add
              </button>
            </div>
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

      {/* SMTP Server Selection Dialog */}
      <Dialog open={smtpDialogOpen} onOpenChange={setSmtpDialogOpen}>
        <DialogContent className="sm:max-w-[700px]">
          <DialogHeader>
            <DialogTitle>Select SMTP Server</DialogTitle>
            <DialogDescription>
              Choose an SMTP server from the list below
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {/* Selected SMTP Tag */}
            {formData.SMTPId !== 0 && (
              <div className="flex flex-wrap gap-2 mb-2">
                <div className="flex items-center gap-1 bg-muted px-2 py-1 rounded-md">
                  <span className="text-sm">
                    {smtpServers.find(server => server.id === formData.SMTPId)?.name || smtpServers.find(server => server.id === formData.SMTPId)?.host} ({smtpServers.find(server => server.id === formData.SMTPId)?.host}:{smtpServers.find(server => server.id === formData.SMTPId)?.port})
                  </span>
                  <button
                    type="button"
                    className="text-muted-foreground hover:text-foreground"
                    onClick={() => setFormData(prev => ({ ...prev, SMTPId: 0 }))}
                  >
                    ×
                  </button>
                </div>
              </div>
            )}
            <div className="border rounded-md overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Host</TableHead>
                    <TableHead>Port</TableHead>
                    <TableHead>Email</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {smtpServers.slice((smtpPage - 1) * pageSize, smtpPage * pageSize).map(server => (
                    <TableRow
                      key={server.id}
                      className={`cursor-pointer ${formData.SMTPId === server.id ? 'bg-muted' : ''}`}
                      onClick={() => setFormData(prev => ({ ...prev, SMTPId: server.id! }))}
                    >
                      <TableCell>{server.name || server.host}</TableCell>
                      <TableCell>{server.host}</TableCell>
                      <TableCell>{server.port}</TableCell>
                      <TableCell>{server.mail}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
            {/* Pagination */}
            <Pagination
              currentPage={smtpPage}
              totalItems={smtpServers.length}
              pageSize={pageSize}
              onPageChange={setSmtpPage}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="secondary" onClick={() => setSmtpDialogOpen(false)}>
              Cancel
            </Button>
            <Button type="button" onClick={() => setSmtpDialogOpen(false)}>
              Confirm
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Email Selection Dialog */}
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
        pageSize={pageSize}
      />
    </>
  );
}