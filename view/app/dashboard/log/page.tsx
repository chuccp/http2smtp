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
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { X, Search, Download, FileText } from 'lucide-react';
import { LogEntry, LogFile } from '@/types/log';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { Pagination } from '@/components/Pagination';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

export default function LogPage() {
  const router = useRouter();
  const t = useTranslations('log');
  const tCommon = useTranslations('common');
  const tAuth = useTranslations('auth');
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [pageNo, setPageNo] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [searchKey, setSearchKey] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<number[]>([]);
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null);
  const [alertHide, setAlertHide] = useState(false);

  useEffect(() => {
    fetchLogs();
  }, [pageNo]);

  useEffect(() => {
    const hide = localStorage.getItem('log_alert_hide');
    if (hide) {
      setAlertHide(true);
    }
  }, []);

  const fetchLogs = async () => {
    try {
      setLoading(true);
      const result = await apiClient.getLogs(pageNo, pageSize, searchKey);
      setLogs(result.list);
      setTotal(result.total);
      setLoading(false);
    } catch (err) {
      setLoading(false);
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert(tAuth('authFailed'));
        router.push('/');
      } else {
        setError(t('fetchError'));
        console.error('Failed to fetch logs:', err);
      }
    }
  };

  const handleSearch = () => {
    setPageNo(1);
    fetchLogs();
  };

  const handleRowClick = (record: LogEntry) => {
    setSelectedRowKeys([record.id]);
    setSelectedLog(record);
    setDetailDialogOpen(true);
  };

  const handleDownload = (file: LogFile) => {
    if (file.data) {
      const link = document.createElement('a');
      link.href = `data:application/octet-stream;base64,${file.data}`;
      link.download = file.name;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    }
  };

  const getStatusColor = (statusStr: string) => {
    if (statusStr.toLowerCase().includes('success')) {
      return 'text-green-600';
    } else if (statusStr.toLowerCase().includes('error') || statusStr.toLowerCase().includes('fail')) {
      return 'text-red-600';
    } else if (statusStr.toLowerCase().includes('warn')) {
      return 'text-yellow-600';
    }
    return 'text-gray-600';
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
      </div>

      {!alertHide && (
        <Alert className="mb-6 bg-blue-50 border-blue-200">
          <AlertDescription className="flex items-center justify-between">
            <span>{t('alertMessage')}</span>
            <button
              onClick={() => {
                setAlertHide(true);
                localStorage.setItem('log_alert_hide', 'true');
              }}
              className="text-blue-500 hover:text-blue-700"
            >
              <X className="h-4 w-4" />
            </button>
          </AlertDescription>
        </Alert>
      )}

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
          {/* Search Bar */}
          <div className="flex gap-4 mb-6">
            <div className="flex-1">
              <Input
                placeholder={t('searchPlaceholder')}
                value={searchKey}
                onChange={(e) => setSearchKey(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
              />
            </div>
            <Button onClick={handleSearch}>
              <Search className="h-4 w-4 mr-2" />
              {t('search')}
            </Button>
          </div>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[100px]">{t('token')}</TableHead>
                <TableHead>{t('subject')}</TableHead>
                <TableHead className="w-[300px]">{t('content')}</TableHead>
                <TableHead className="w-[100px]">{tCommon('status')}</TableHead>
                <TableHead className="w-[150px]">{t('result')}</TableHead>
                <TableHead className="w-[180px]">{tCommon('createTime')}</TableHead>
                <TableHead className="w-[100px]">{tCommon('actions')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((log) => (
                <TableRow
                  key={log.id}
                  className={`cursor-pointer ${
                    selectedRowKeys.includes(log.id) ? 'bg-blue-50' : ''
                  }`}
                  onClick={() => handleRowClick(log)}
                >
                  <TableCell className="font-medium">{log.token}</TableCell>
                  <TableCell>
                    <HighlightText text={log.subject || '-'} searchKey={searchKey} />
                  </TableCell>
                  <TableCell>
                    <div className="max-h-20 overflow-y-auto text-sm">
                      <HighlightText text={log.content || '-'} searchKey={searchKey} />
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className={`font-medium ${getStatusColor(log.statusStr)}`}>
                      {log.statusStr}
                    </span>
                  </TableCell>
                  <TableCell className="max-w-[150px] truncate">{log.result}</TableCell>
                  <TableCell>
                    {log.createTime ? new Date(log.createTime).toLocaleString() : '-'}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleRowClick(log);
                        }}
                        title={t('viewDetails')}
                      >
                        <FileText className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {logs.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-500">{t('noLogs')}</p>
            </div>
          )}
          
          <div className="flex justify-end mt-4">
            <Pagination
              pageNo={pageNo}
              pageSize={pageSize}
              total={total}
              onPageChange={(page) => setPageNo(page)}
            />
          </div>
        </CardContent>
      </Card>

      {/* Detail Dialog */}
      <Dialog open={detailDialogOpen} onOpenChange={setDetailDialogOpen}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{t('details')}</DialogTitle>
            <DialogDescription>
              {selectedLog?.createTime ? new Date(selectedLog.createTime).toLocaleString() : ''}
            </DialogDescription>
          </DialogHeader>
          
          {!selectedLog ? (
            <div className="py-8 text-center text-gray-500">
              {t('pleaseSelectOne')}
            </div>
          ) : (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label className="text-sm font-medium">{t('token')}</Label>
                  <div className="p-3 border rounded-md bg-gray-50">
                    {selectedLog.token || '-'}
                  </div>
                </div>
                <div className="space-y-2">
                  <Label className="text-sm font-medium">{t('smtpServer')}</Label>
                  <div className="p-3 border rounded-md bg-gray-50">
                    {selectedLog.smtp || '-'}
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <Label className="text-sm font-medium">{t('recipients')}</Label>
                <div className="p-3 border rounded-md bg-gray-50">
                  {selectedLog.mail || '-'}
                </div>
              </div>

              <div className="space-y-2">
                <Label className="text-sm font-medium">{t('subject')}</Label>
                <div className="p-3 border rounded-md bg-gray-50">
                  {selectedLog.subject || '-'}
                </div>
              </div>

              <div className="space-y-2">
                <Label className="text-sm font-medium">{t('content')}</Label>
                <div className="p-3 border rounded-md bg-gray-50 whitespace-pre-wrap max-h-60 overflow-y-auto">
                  {selectedLog.content || '-'}
                </div>
              </div>

              <div className="space-y-2">
                <Label className="text-sm font-medium">{t('files')}</Label>
                <div className="border rounded-md">
                  {selectedLog.fileArray && selectedLog.fileArray.length > 0 ? (
                    <ul className="divide-y">
                      {selectedLog.fileArray.map((file, index) => (
                        <li key={index} className="p-3 flex items-center justify-between hover:bg-gray-50">
                          <span className="text-sm">{file.name}</span>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDownload(file)}
                          >
                            <Download className="h-4 w-4 mr-2" />
                            {t('download')}
                          </Button>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <div className="p-3 text-gray-500">{t('noFiles')}</div>
                  )}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label className="text-sm font-medium">{tCommon('status')}</Label>
                  <div className="p-3 border rounded-md bg-gray-50">
                    <span className={`font-medium ${getStatusColor(selectedLog.statusStr)}`}>
                      {selectedLog.statusStr}
                    </span>
                  </div>
                </div>
                <div className="space-y-2">
                  <Label className="text-sm font-medium">{t('result')}</Label>
                  <div className="p-3 border rounded-md bg-gray-50">
                    {selectedLog.result || '-'}
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4 pt-4 border-t">
                <div className="space-y-2">
                  <Label className="text-sm font-medium">{tCommon('createTime')}</Label>
                  <div className="p-3 border rounded-md bg-gray-50">
                    {selectedLog.createTime ? new Date(selectedLog.createTime).toLocaleString() : '-'}
                  </div>
                </div>
                <div className="space-y-2">
                  <Label className="text-sm font-medium">{tCommon('updateTime')}</Label>
                  <div className="p-3 border rounded-md bg-gray-50">
                    {selectedLog.updateTime ? new Date(selectedLog.updateTime).toLocaleString() : '-'}
                  </div>
                </div>
              </div>
            </div>
          )}
          
          <div className="flex justify-end mt-4">
            <Button onClick={() => setDetailDialogOpen(false)}>
              <X className="h-4 w-4 mr-2" />
              {tCommon('close')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// Helper component to highlight search text
function HighlightText({ text, searchKey }: { text: string; searchKey: string }) {
  if (!searchKey || !text) {
    return <span>{text}</span>;
  }

  const parts = text.split(new RegExp(`(${searchKey})`, 'gi'));
  return (
    <span>
      {parts.map((part, i) =>
        part.toLowerCase() === searchKey.toLowerCase() ? (
          <mark key={i} className="bg-yellow-200 px-0.5">
            {part}
          </mark>
        ) : (
          <span key={i}>{part}</span>
        )
      )}
    </span>
  );
}
