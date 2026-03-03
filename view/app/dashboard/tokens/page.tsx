'use client';

import { useState, useEffect } from 'react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Plus, Edit, Trash2, Copy, Check } from 'lucide-react';
import { TokenConfig } from '@/types/token';
import { SMTPConfig } from '@/types/smtp';
import { MailConfig } from '@/types/mail';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { Pagination } from '@/components/Pagination';

export default function TokenPage() {
  const router = useRouter();
  const [tokens, setTokens] = useState<TokenConfig[]>([]);
  const [smtpServers, setSmtpServers] = useState<SMTPConfig[]>([]);
  const [mails, setMails] = useState<MailConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingToken, setEditingToken] = useState<TokenConfig | null>(null);
  const [copiedId, setCopiedId] = useState<number | null>(null);
  const [formData, setFormData] = useState<TokenConfig>({
    token: '',
    subject: '',
    receiveEmailIds: '',
    SMTPId: 0,
    isUse: true,
  });
  const [submitting, setSubmitting] = useState(false);
  const [pageNo, setPageNo] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    fetchData();
  }, [pageNo]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [tokensResult, smtpData, mailsData] = await Promise.all([
        apiClient.getTokens(pageNo, pageSize),
        apiClient.getSMTPServers(1, 100),
        apiClient.getMails(1, 100),
      ]);
      setTokens(tokensResult.list);
      setTotal(tokensResult.total);
      setSmtpServers(smtpData.list);
      setMails(mailsData.list);
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setError('Failed to fetch tokens');
        console.error('Failed to fetch tokens:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  const generateToken = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < 32; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  };

  const handleEdit = (token: TokenConfig) => {
    setEditingToken(token);
    setFormData({
      ...token,
    });
    setDialogOpen(true);
  };

  const handleAdd = () => {
    setEditingToken(null);
    setFormData({
      token: generateToken(),
      subject: '',
      receiveEmailIds: '',
      SMTPId: smtpServers.length > 0 ? smtpServers[0].id! : 0,
      isUse: true,
    });
    setDialogOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (confirm('Are you sure you want to delete this token?')) {
      try {
        await apiClient.deleteToken(id);
        fetchData();
      } catch (err) {
        if (err instanceof Error && err.message === 'Unauthorized') {
          alert('Authentication failed, redirecting to home');
          router.push('/');
        } else {
          setError('Failed to delete token');
          console.error('Failed to delete token:', err);
        }
      }
    }
  };

  const handleCopy = async (token: string, id: number) => {
    try {
      await navigator.clipboard.writeText(token);
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
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

      setDialogOpen(false);
      setEditingToken(null);
      fetchData();
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
    setDialogOpen(false);
    setEditingToken(null);
  };

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

  const handlePageChange = (page: number) => {
    setPageNo(page);
  };

  const getSMTPName = (id: number) => {
    const smtp = smtpServers.find(s => s.id === id);
    return smtp ? (smtp.name || smtp.host) : '-';
  };

  const getMailNames = (ids: string) => {
    if (!ids) return '-';
    const idArray = ids.split(',').map(id => parseInt(id.trim()));
    const mailNames = idArray
      .map(id => {
        const mail = mails.find(m => m.id === id);
        return mail ? (mail.name || mail.mail) : null;
      })
      .filter(Boolean);
    return mailNames.length > 0 ? mailNames.join(', ') : '-';
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">Loading tokens...</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">API Tokens</h1>
          <p className="text-gray-600">Manage API tokens for email sending</p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button className="gap-2" onClick={handleAdd}>
              <Plus className="h-4 w-4" />
              Generate Token
            </Button>
          </DialogTrigger>
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
      </div>

      {error && (
        <Alert variant="destructive" className="mb-6">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Token List</CardTitle>
          <CardDescription>
            {total} tokens configured
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Token</TableHead>
                <TableHead>Subject</TableHead>
                <TableHead>SMTP Server</TableHead>
                <TableHead>Recipients</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Create Time</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tokens.map((token) => (
                <TableRow key={token.id}>
                  <TableCell className="font-mono text-sm">
                    <div className="flex items-center gap-2">
                      {token.token.substring(0, 12)}...
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6"
                        onClick={() => handleCopy(token.token, token.id!)}
                      >
                        {copiedId === token.id ? (
                          <Check className="h-3 w-3 text-green-500" />
                        ) : (
                          <Copy className="h-3 w-3" />
                        )}
                      </Button>
                    </div>
                  </TableCell>
                  <TableCell>{token.subject || '-'}</TableCell>
                  <TableCell>{token.SMTPStr || getSMTPName(token.SMTPId)}</TableCell>
                  <TableCell>{token.receiveEmailsStr || getMailNames(token.receiveEmailIds)}</TableCell>
                  <TableCell>
                    <span className={`px-2 py-1 rounded-full text-xs ${token.isUse ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'}`}>
                      {token.isUse ? 'Active' : 'Inactive'}
                    </span>
                  </TableCell>
                  <TableCell>{token.createTime ? new Date(token.createTime).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleEdit(token)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(token.id!)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {tokens.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-500 mb-4">No tokens configured yet</p>
              <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                <DialogTrigger asChild>
                  <Button onClick={handleAdd}>Generate Your First Token</Button>
                </DialogTrigger>
              </Dialog>
            </div>
          )}
          <Pagination pageNo={pageNo} pageSize={pageSize} total={total} onPageChange={handlePageChange} />
        </CardContent>
      </Card>
    </div>
  );
}