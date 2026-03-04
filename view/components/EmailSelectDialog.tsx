'use client';

import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Pagination } from '@/components/ui/pagination';
import { MailConfig } from '@/types/mail';

interface EmailSelectDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  emails: MailConfig[];
  selectedEmailIds: number[];
  onSelectedEmailIdsChange: (ids: number[]) => void;
  pageSize?: number;
}

export function EmailSelectDialog({
  open,
  onOpenChange,
  emails,
  selectedEmailIds,
  onSelectedEmailIdsChange,
  pageSize = 10,
}: EmailSelectDialogProps) {
  const [page, setPage] = useState(1);

  // 直接在渲染时计算选中的邮件，避免使用 useEffect 和 setState
  const selectedEmails = emails.filter(email => selectedEmailIds.includes(email.id!));

  const handlePageChange = (newPage: number) => {
    setPage(newPage);
  };

  const handleRowClick = (mail: MailConfig) => {
    if (selectedEmailIds.includes(mail.id!)) {
      const updatedSelectedEmailIds = selectedEmailIds.filter(id => id !== mail.id);
      onSelectedEmailIdsChange(updatedSelectedEmailIds);
    } else {
      const updatedSelectedEmailIds = [...selectedEmailIds, mail.id!];
      onSelectedEmailIdsChange(updatedSelectedEmailIds);
    }
  };

  const handleTagRemove = (email: MailConfig) => {
    const updatedSelectedEmailIds = selectedEmailIds.filter(id => id !== email.id);
    onSelectedEmailIdsChange(updatedSelectedEmailIds);
  };

  const handleConfirm = () => {
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[700px]">
        <DialogHeader>
          <DialogTitle>Select Recipient Emails</DialogTitle>
          <DialogDescription>
            Choose one or more recipient emails from the list below
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          {/* Selected Emails Tags */}
          <div className="flex flex-wrap gap-2 mb-2">
            {selectedEmails.map(email => (
              <div key={email.id} className="flex items-center gap-1 bg-muted px-2 py-1 rounded-md">
                <span className="text-sm">{email.name || email.mail}</span>
                <button
                  type="button"
                  className="text-muted-foreground hover:text-foreground"
                  onClick={() => handleTagRemove(email)}
                >
                  ×
                </button>
              </div>
            ))}
          </div>
          <div className="border rounded-md overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Email</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {emails.slice((page - 1) * pageSize, page * pageSize).map(mail => (
                  <TableRow
                    key={mail.id}
                    className={`cursor-pointer ${selectedEmailIds.includes(mail.id!) ? 'bg-muted' : ''}`}
                    onClick={() => handleRowClick(mail)}
                  >
                    <TableCell>{mail.name || mail.mail}</TableCell>
                    <TableCell>{mail.mail}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
          {/* Pagination */}
          <Pagination
            currentPage={page}
            totalItems={emails.length}
            pageSize={pageSize}
            onPageChange={handlePageChange}
          />
        </div>
        <DialogFooter>
          <Button type="button" variant="secondary" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="button" onClick={handleConfirm}>
            Confirm
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}