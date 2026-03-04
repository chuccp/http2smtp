'use client';

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Edit, Trash2 } from 'lucide-react';
import { MailConfig } from '@/types/mail';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { Pagination } from '@/components/Pagination';
import { MailFormDialog } from '@/components/MailFormDialog';

export default function MailPage() {
  const router = useRouter();
  const t = useTranslations('mail');
  const tCommon = useTranslations('common');
  const tAuth = useTranslations('auth');
  const [mails, setMails] = useState<MailConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingMail, setEditingMail] = useState<MailConfig | null>(null);
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
        alert(tAuth('authFailed'));
        router.push('/');
      } else {
        setError(t('fetchError'));
        console.error('Failed to fetch mail addresses:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (mail: MailConfig) => {
    setEditingMail(mail);
    setDialogOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (confirm(t('deleteConfirm'))) {
      try {
        await apiClient.deleteMail(id);
        fetchMails();
      } catch (err) {
        if (err instanceof Error && err.message === 'Unauthorized') {
          alert(tAuth('authFailed'));
          router.push('/');
        } else {
          setError(t('deleteError'));
          console.error('Failed to delete mail address:', err);
        }
      }
    }
  };

  const handleDialogClose = () => {
    setDialogOpen(false);
    setEditingMail(null);
  };

  const handleFormSuccess = () => {
    handleDialogClose();
    fetchMails();
  };

  const handlePageChange = (page: number) => {
    setPageNo(page);
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">{tCommon('loading')}</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">{t('title')}</h1>
          <p className="text-gray-600">{t('subtitle')}</p>
        </div>
        <MailFormDialog onSuccess={handleFormSuccess} />
      </div>

      {error && (
        <Alert variant="destructive" className="mb-6">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle>{t('listTitle')}</CardTitle>
          <CardDescription>
            {t('listCount', { count: total })}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{tCommon('name')}</TableHead>
                <TableHead>{t('mailAddress')}</TableHead>
                <TableHead>{tCommon('createTime')}</TableHead>
                <TableHead>{tCommon('actions')}</TableHead>
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
              <p className="text-gray-500 mb-4">{t('noAddresses')}</p>
              <MailFormDialog triggerButton={true} onSuccess={handleFormSuccess} />
            </div>
          )}
          <Pagination pageNo={pageNo} pageSize={pageSize} total={total} onPageChange={handlePageChange} />
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <MailFormDialog
        open={dialogOpen}
        onOpenChange={handleDialogClose}
        editingMail={editingMail}
        onSuccess={handleFormSuccess}
        triggerButton={false}
      />
    </div>
  );
}