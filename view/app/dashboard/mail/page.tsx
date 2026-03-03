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

  useEffect(() => {
    fetchMails();
  }, []);

  const fetchMails = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getMails();
      setMails(data);
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('权限验证失败，将跳转到首页');
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
    if (confirm('确定要删除此邮件地址吗？')) {
      try {
        await apiClient.deleteMail(id);
        setMails(mails.filter(m => m.id !== id));
      } catch (err) {
        if (err instanceof Error && err.message === 'Unauthorized') {
          alert('权限验证失败，将跳转到首页');
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
        setError('请输入邮件地址');
        return;
      }

      // Simple email validation
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formData.mail)) {
        setError('请输入有效的邮件地址');
        return;
      }

      if (editingMail) {
        await apiClient.updateMail(formData);
        setMails(mails.map(m =>
          m.id === editingMail.id ? formData : m
        ));
      } else {
        await apiClient.createMail(formData);
        fetchMails(); // Refresh list to get the new item with ID
      }

      setDialogOpen(false);
      setEditingMail(null);
      setFormData({
        name: '',
        mail: '',
      });
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('权限验证失败，将跳转到首页');
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

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">加载中...</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">邮件地址管理</h1>
          <p className="text-gray-600">管理常用的收件人邮件地址</p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button className="gap-2">
              <Plus className="h-4 w-4" />
              添加地址
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[450px]">
            <DialogHeader>
              <DialogTitle>{editingMail ? '编辑邮件地址' : '添加邮件地址'}</DialogTitle>
              <DialogDescription>
                {editingMail ? '修改收件人邮件地址信息' : '添加新的收件人邮件地址'}
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleSubmit} className="space-y-4 py-4">
              {error && (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}
              <div className="space-y-2">
                <Label htmlFor="name">名称</Label>
                <Input
                  id="name"
                  name="name"
                  value={formData.name}
                  onChange={handleInputChange}
                  placeholder="例如: 张三、客服"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="mail">邮件地址 *</Label>
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
                  取消
                </Button>
                <Button type="submit" disabled={submitting}>
                  {submitting ? '保存中...' : '保存'}
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
          <CardTitle>邮件地址列表</CardTitle>
          <CardDescription>
            共 {mails.length} 个邮件地址
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>邮件地址</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead>操作</TableHead>
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
              <p className="text-gray-500 mb-4">暂无邮件地址</p>
              <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                <DialogTrigger asChild>
                  <Button>添加第一个邮件地址</Button>
                </DialogTrigger>
              </Dialog>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}