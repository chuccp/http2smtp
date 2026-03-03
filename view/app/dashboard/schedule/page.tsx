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
import { Edit, Trash2, Send } from 'lucide-react';
import { ScheduleConfig } from '@/types/schedule';
import { TokenConfig } from '@/types/token';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { Pagination } from '@/components/Pagination';
import { ScheduleFormDialog } from '@/components/ScheduleFormDialog';

export default function SchedulePage() {
  const router = useRouter();
  const [schedules, setSchedules] = useState<ScheduleConfig[]>([]);
  const [tokens, setTokens] = useState<TokenConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingSchedule, setEditingSchedule] = useState<ScheduleConfig | null>(null);
  const [pageNo, setPageNo] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    fetchSchedules();
  }, [pageNo]);

  const fetchSchedules = async () => {
    try {
      setLoading(true);
      const result = await apiClient.getSchedules(pageNo, pageSize);
      setSchedules(result.list);
      setTotal(result.total);
    } catch (err) {
      if (err instanceof Error && err.message === 'Unauthorized') {
        alert('Authentication failed, redirecting to home');
        router.push('/');
      } else {
        setError('Failed to fetch schedules');
        console.error('Failed to fetch schedules:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  // Fetch tokens only when dialog opens
  const fetchTokens = useCallback(async () => {
    if (tokens.length === 0) {
      try {
        const result = await apiClient.getTokens(1, 100);
        setTokens(result.list);
      } catch (err) {
        console.error('Failed to fetch tokens:', err);
      }
    }
  }, [tokens.length]);

  const handleEdit = async (schedule: ScheduleConfig) => {
    setEditingSchedule(schedule);
    await fetchTokens();
    setDialogOpen(true);
  };

  const handleAdd = async () => {
    setEditingSchedule(null);
    await fetchTokens();
    setDialogOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (confirm('Are you sure you want to delete this schedule?')) {
      try {
        await apiClient.deleteSchedule(id);
        fetchSchedules();
      } catch (err) {
        if (err instanceof Error && err.message === 'Unauthorized') {
          alert('Authentication failed, redirecting to home');
          router.push('/');
        } else {
          setError('Failed to delete schedule');
          console.error('Failed to delete schedule:', err);
        }
      }
    }
  };

  const handleSendMail = async (schedule: ScheduleConfig) => {
    if (confirm('Send email now using this schedule configuration?')) {
      try {
        await apiClient.sendMailBySchedule(schedule);
        alert('Email sent successfully');
      } catch (err) {
        if (err instanceof Error && err.message === 'Unauthorized') {
          alert('Authentication failed, redirecting to home');
          router.push('/');
        } else {
          setError('Failed to send email');
          console.error('Failed to send email:', err);
        }
      }
    }
  };

  const handleDialogClose = () => {
    setDialogOpen(false);
    setEditingSchedule(null);
  };

  const handleFormSuccess = () => {
    handleDialogClose();
    fetchSchedules();
  };

  const handlePageChange = (page: number) => {
    setPageNo(page);
  };

  const getTokenName = (tokenStr: string) => {
    const token = tokens.find(t => t.token === tokenStr);
    return token ? (token.subject || token.token?.substring(0, 12) + '...') : tokenStr?.substring(0, 12) + '...';
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">Loading schedules...</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">Scheduled Tasks</h1>
          <p className="text-gray-600">Manage scheduled email tasks with cron expressions</p>
        </div>
        <Button className="gap-2" onClick={handleAdd}>
          Add Schedule
        </Button>
      </div>

      {error && (
        <Alert variant="destructive" className="mb-6">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Schedule List</CardTitle>
          <CardDescription>
            {total} scheduled tasks configured
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Token</TableHead>
                <TableHead>Cron</TableHead>
                <TableHead>URL</TableHead>
                <TableHead>Method</TableHead>
                <TableHead>Template</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Create Time</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {schedules.map((schedule) => (
                <TableRow key={schedule.id}>
                  <TableCell className="font-medium">{schedule.name || '-'}</TableCell>
                  <TableCell className="font-mono text-sm">{getTokenName(schedule.token)}</TableCell>
                  <TableCell className="font-mono text-sm">{schedule.cron}</TableCell>
                  <TableCell className="max-w-[200px] truncate" title={schedule.url}>{schedule.url || '-'}</TableCell>
                  <TableCell>{schedule.method || 'GET'}</TableCell>
                  <TableCell>{schedule.useTemplate ? 'Yes' : 'No'}</TableCell>
                  <TableCell>
                    <span className={`px-2 py-1 rounded-full text-xs ${schedule.isUse ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'}`}>
                      {schedule.isUse ? 'Active' : 'Inactive'}
                    </span>
                  </TableCell>
                  <TableCell>{schedule.createTime ? new Date(schedule.createTime).toLocaleString() : '-'}</TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleSendMail(schedule)}
                        title="Send email now"
                      >
                        <Send className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleEdit(schedule)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(schedule.id!)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {schedules.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-500 mb-4">No scheduled tasks configured yet</p>
              <Button onClick={handleAdd}>Add Your First Schedule</Button>
            </div>
          )}
          <Pagination pageNo={pageNo} pageSize={pageSize} total={total} onPageChange={handlePageChange} />
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <ScheduleFormDialog
        open={dialogOpen}
        onOpenChange={handleDialogClose}
        editingSchedule={editingSchedule}
        tokens={tokens}
        onSuccess={handleFormSuccess}
        triggerButton={false}
      />
    </div>
  );
}