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
import { Edit, Trash2, Send } from 'lucide-react';
import { SMTPConfig } from '@/types/smtp';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { Pagination } from '@/components/Pagination';
import { SMTPFormDialog } from '@/components/SMTPFormDialog';
import { SendMailDialog } from '@/components/SendMailDialog';
import { MailConfig } from '@/types/mail';

export default function SMTPPage() {
  const router = useRouter();
  const t = useTranslations('smtp');
  const tCommon = useTranslations('common');
  const tAuth = useTranslations('auth');
  const [smtpServers, setSmtpServers] = useState<SMTPConfig[]>([]);
  const [mails, setMails] = useState<MailConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingServer, setEditingServer] = useState<SMTPConfig | null>(null);
  const [sendMailDialogOpen, setSendMailDialogOpen] = useState(false);
  const [selectedSMTPId, setSelectedSMTPId] = useState<number | null>(null);
  const [pageNo, setPageNo] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    fetchSMTPServers();
  }, [pageNo]);

  const fetchSMTPServers = async () => {
    try {
      setLoading(true);
      const [smtpResult, mailResult] = await Promise.all([
        apiClient.getSMTPServers(pageNo, pageSize),
        apiClient.getMails(1, 100),
      ]);
      setSmtpServers(smtpResult.list);
      setTotal(smtpResult.total);
      setMails(mailResult.list);
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert(tAuth('authFailed'));
        router.push('/');
      } else {
        setError(t('fetchError'));
        console.error('Failed to fetch SMTP servers:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (server: SMTPConfig) => {
    setEditingServer(server);
    setDialogOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (confirm(t('deleteConfirm'))) {
      try {
        await apiClient.deleteSMTPServer(id);
        fetchSMTPServers();
      } catch (err) {
        if (err instanceof Error && err.message === 'Unauthorized') {
          alert(tAuth('authFailed'));
          router.push('/');
        } else {
          setError(t('deleteError'));
          console.error('Failed to delete SMTP server:', err);
        }
      }
    }
  };

  const handleDialogClose = () => {
    setDialogOpen(false);
    setEditingServer(null);
  };

  const handleFormSuccess = () => {
    handleDialogClose();
    fetchSMTPServers();
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
        <SMTPFormDialog onSuccess={handleFormSuccess} />
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
                <TableHead>{t('serverName')}</TableHead>
                <TableHead>{t('host')}</TableHead>
                <TableHead>{t('port')}</TableHead>
                <TableHead>{t('fromAddress')}</TableHead>
                <TableHead>{t('username')}</TableHead>
                <TableHead>{tCommon('createTime')}</TableHead>
                <TableHead>{tCommon('actions')}</TableHead>
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
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => {
                          setSelectedSMTPId(server.id!);
                          setSendMailDialogOpen(true);
                        }}
                      >
                        <Send className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {smtpServers.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-500 mb-4">{t('noServers')}</p>
              <SMTPFormDialog triggerButton={true} onSuccess={handleFormSuccess} />
            </div>
          )}
          <Pagination pageNo={pageNo} pageSize={pageSize} total={total} onPageChange={handlePageChange} />
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <SMTPFormDialog
        open={dialogOpen}
        onOpenChange={handleDialogClose}
        editingServer={editingServer}
        onSuccess={handleFormSuccess}
        triggerButton={false}
      />

      {/* Send Mail Dialog */}
      <SendMailDialog
        open={sendMailDialogOpen}
        onOpenChange={setSendMailDialogOpen}
        smtpServers={smtpServers.filter(server => server.id === selectedSMTPId)}
        mails={mails}
        onSuccess={() => {
          setSendMailDialogOpen(false);
          setSelectedSMTPId(null);
        }}
        triggerButton={false}
      />
    </div>
  );
}