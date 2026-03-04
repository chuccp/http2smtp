'use client';

import { useState, useEffect, useCallback } from 'react';
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
import { Edit, Trash2, Copy, Check, Send } from 'lucide-react';
import { TokenConfig } from '@/types/token';
import { SMTPConfig } from '@/types/smtp';
import { MailConfig } from '@/types/mail';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { Pagination } from '@/components/Pagination';
import { TokenFormDialog } from '@/components/TokenFormDialog';
import { SendMailByTokenDialog } from '@/components/SendMailByTokenDialog';

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
  const [sendMailDialogOpen, setSendMailDialogOpen] = useState(false);
  const [selectedToken, setSelectedToken] = useState<TokenConfig | null>(null);
  const [pageNo, setPageNo] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    fetchTokens();
  }, [pageNo]);

  const fetchTokens = async () => {
    try {
      setLoading(true);
      const result = await apiClient.getTokens(pageNo, pageSize);
      setTokens(result.list);
      setTotal(result.total);
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

  // Fetch auxiliary data only when dialog opens
  const fetchAuxiliaryData = useCallback(async () => {
    if (smtpServers.length === 0 || mails.length === 0) {
      try {
        const [smtpData, mailsData] = await Promise.all([
          apiClient.getSMTPServers(1, 100),
          apiClient.getMails(1, 100),
        ]);
        setSmtpServers(smtpData.list);
        setMails(mailsData.list);
      } catch (err) {
        console.error('Failed to fetch auxiliary data:', err);
      }
    }
  }, [smtpServers.length, mails.length]);

  const handleEdit = async (token: TokenConfig) => {
    setEditingToken(token);
    await fetchAuxiliaryData();
    setDialogOpen(true);
  };

  const handleAdd = async () => {
    setEditingToken(null);
    await fetchAuxiliaryData();
    setDialogOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (confirm('Are you sure you want to delete this token?')) {
      try {
        await apiClient.deleteToken(id);
        fetchTokens();
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

  const handleDialogClose = () => {
    setDialogOpen(false);
    setEditingToken(null);
  };

  const handleFormSuccess = () => {
    handleDialogClose();
    fetchTokens();
  };

  const handlePageChange = (page: number) => {
    setPageNo(page);
  };

  const handleSendMail = (token: TokenConfig) => {
    setSelectedToken(token);
    setSendMailDialogOpen(true);
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
        <Button className="gap-2" onClick={handleAdd}>
          Generate Token
        </Button>
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
          <div className="overflow-x-auto">
            <Table className="min-w-full">
              <TableHeader>
                <TableRow>
                  <TableHead className="w-1/6">Token</TableHead>
                  <TableHead className="w-1/6">Subject</TableHead>
                  <TableHead className="w-1/6">SMTP Server</TableHead>
                  <TableHead className="w-1/4">Recipients</TableHead>
                  <TableHead className="w-1/12">Status</TableHead>
                  <TableHead className="w-1/8">Create Time</TableHead>
                  <TableHead className="w-1/12">Actions</TableHead>
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
                    <TableCell className="whitespace-pre-wrap break-words max-w-[200px]">
                      {token.receiveEmailsStr || getMailNames(token.receiveEmailIds)}
                    </TableCell>
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
                          onClick={() => handleSendMail(token)}
                          title="Send Mail"
                        >
                          <Send className="h-4 w-4" />
                        </Button>
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
          </div>
          {tokens.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-500 mb-4">No tokens configured yet</p>
              <Button onClick={handleAdd}>Generate Your First Token</Button>
            </div>
          )}
          <Pagination pageNo={pageNo} pageSize={pageSize} total={total} onPageChange={handlePageChange} />
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <TokenFormDialog
        open={dialogOpen}
        onOpenChange={handleDialogClose}
        editingToken={editingToken}
        smtpServers={smtpServers}
        mails={mails}
        onSuccess={handleFormSuccess}
        triggerButton={false}
      />

      {selectedToken && (
        <SendMailByTokenDialog
          open={sendMailDialogOpen}
          onOpenChange={setSendMailDialogOpen}
          token={selectedToken.token}
          tokenSubject={selectedToken.subject || ''}
          mails={mails}
          onSuccess={() => {
            setSendMailDialogOpen(false);
            setSelectedToken(null);
          }}
        />
      )}
    </div>
  );
}