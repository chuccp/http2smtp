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
import { MailConfig } from '@/types/mail';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { Pagination } from '@/components/Pagination';

export default function MailPage() {
  const router = useRouter();
  const [mails, setMails] = useState<MailConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingMail, setEditingMail] = useState<MailConfig | null>(null);
  const [formData, setFormData] = useState<MailConfig>({
    name: '',
    mail: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [pageNo, setPageNo] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    fetchMails();
  }, [pageNo]);

  const fetchMails = async () => {
    try {
      setLoading(true);
      const result = await apiClient.getMails(pageNo, pageSize);
      setMails(result.list);
      setTotal(result.total);
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setError('Failed to fetch mail addresses');
        console.error('Failed to fetch mail addresses:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (mail: MailConfig) => {
    setEditingMail(mail);
    setFormData({
      ...mail,
    });
    setDialogOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (confirm('Are you sure you want to delete this mail address?')) {
      try {
        await apiClient.deleteMail(id);
        fetchMails();
      } catch (err) {
        if (err instanceof Error && err.message === 'Unauthorized') {
          alert('Authentication failed, redirecting to home');
          router.push('/');
        } else {
          setError('Failed to delete mail address');
          console.error('Failed to delete mail address:', err);
        }
      }
    }
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

      setDialogOpen(false);
      setEditingMail(null);
      setFormData({
        name: '',
        mail: '',
      });
      fetchMails();
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
    setDialogOpen(false);
    setEditingMail(null);
    setFormData({
      name: '',
      mail: '',
    });
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
  };

  const handlePageChange = (page: number) => {
    setPageNo(page);
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">Mail Addresses</h1>
          <p className="text-gray-600">Manage recipient mail addresses</p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button className="gap-2">
              <Plus className="h-4 w-4" />
              Add Address
            </Button>
          </DialogTrigger>
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
      </div>

      {error && (
        <Alert variant="destructive" className="mb-6">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Mail Address List</CardTitle>
          <CardDescription>
            {total} mail addresses
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Mail Address</TableHead>
                <TableHead>Create Time</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {mails.map((mail) => (
                <TableRow key={mail.id}>
                  <TableCell className="font-medium">{mail.name || '-'}</TableCell>
                  <TableCell>{mail.mail}</TableCell>
                  <TableCell>{mail.createTime ? new Date(mail.createTime).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleEdit(mail)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(mail.id!)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {mails.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-500 mb-4">No mail addresses yet</p>
              <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                <DialogTrigger asChild>
                  <Button>Add Your First Address</Button>
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