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
import { Plus, Edit, Trash2 } from 'lucide-react';
import { SMTPConfig } from '@/types/smtp';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { Pagination } from '@/components/Pagination';

export default function SMTPPage() {
  const router = useRouter();
  const [smtpServers, setSmtpServers] = useState<SMTPConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
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
  const [pageNo, setPageNo] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    fetchSMTPServers();
  }, [pageNo]);

  const fetchSMTPServers = async () => {
    try {
      setLoading(true);
      const result = await apiClient.getSMTPServers(pageNo, pageSize);
      setSmtpServers(result.list);
      setTotal(result.total);
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setError('Failed to fetch SMTP servers');
        console.error('Failed to fetch SMTP servers:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (server: SMTPConfig) => {
    setEditingServer(server);
    setFormData({
      ...server,
    });
    setDialogOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (confirm('Are you sure you want to delete this SMTP server?')) {
      try {
        await apiClient.deleteSMTPServer(id);
        fetchSMTPServers();
      } catch (err) {
        if (err instanceof Error && err.message === 'Unauthorized') {
          alert('Authentication failed, redirecting to home');
          router.push('/');
        } else {
          setError('Failed to delete SMTP server');
          console.error('Failed to delete SMTP server:', err);
        }
      }
    }
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

      setDialogOpen(false);
      setEditingServer(null);
      setFormData({
        name: '',
        host: '',
        port: 587,
        mail: '',
        username: '',
        password: '',
      });
      fetchSMTPServers();
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
    setDialogOpen(false);
    setEditingServer(null);
    setFormData({
      name: '',
      host: '',
      port: 587,
      mail: '',
      username: '',
      password: '',
    });
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'number' ? parseInt(value) || 0 : value,
    }));
  };

  const handlePageChange = (page: number) => {
    setPageNo(page);
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">Loading SMTP servers...</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">SMTP Servers</h1>
          <p className="text-gray-600">Manage your email delivery servers</p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button className="gap-2">
              <Plus className="h-4 w-4" />
              Add Server
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[550px]">
            <DialogHeader>
              <DialogTitle>{editingServer ? 'Edit SMTP Server' : 'Add SMTP Server'}</DialogTitle>
              <DialogDescription>
                {editingServer ? 'Modify the settings of an existing SMTP server.' : 'Configure a new SMTP server for email delivery.'}
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleSubmit} className="space-y-4 py-4">
              {error && (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
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
              <DialogFooter>
                <Button type="button" variant="secondary" onClick={handleCancel} disabled={submitting}>
                  Cancel
                </Button>
                <Button type="submit" disabled={submitting}>
                  {submitting ? 'Saving...' : 'Save Server'}
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
          <CardTitle>SMTP Server List</CardTitle>
          <CardDescription>
            {total} servers configured
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Host</TableHead>
                <TableHead>Port</TableHead>
                <TableHead>From Address</TableHead>
                <TableHead>Username</TableHead>
                <TableHead>Create Time</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {smtpServers.map((server) => (
                <TableRow key={server.id}>
                  <TableCell className="font-medium">{server.name || server.username}</TableCell>
                  <TableCell>{server.host}</TableCell>
                  <TableCell>{server.port}</TableCell>
                  <TableCell>{server.mail}</TableCell>
                  <TableCell>{server.username}</TableCell>
                  <TableCell>{server.createTime ? new Date(server.createTime).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleEdit(server)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(server.id!)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {smtpServers.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-500 mb-4">No SMTP servers configured yet</p>
              <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                <DialogTrigger asChild>
                  <Button>Add Your First Server</Button>
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