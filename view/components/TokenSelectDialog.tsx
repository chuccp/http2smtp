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
import { TokenConfig } from '@/types/token';

interface TokenSelectDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tokens: TokenConfig[];
  selectedTokenId: number | null;
  onSelectedTokenIdChange: (id: number | null) => void;
  pageSize?: number;
}

export function TokenSelectDialog({
  open,
  onOpenChange,
  tokens,
  selectedTokenId,
  onSelectedTokenIdChange,
  pageSize = 10,
}: TokenSelectDialogProps) {
  const [page, setPage] = useState(1);

  // 直接在渲染时计算选中的 token，避免使用 useEffect 和 setState
  const selectedToken = tokens.find(token => token.id === selectedTokenId) || null;

  const handlePageChange = (newPage: number) => {
    setPage(newPage);
  };

  const handleRowClick = (token: TokenConfig) => {
    onSelectedTokenIdChange(token.id!);
  };

  const handleTagRemove = () => {
    onSelectedTokenIdChange(null);
  };

  const handleConfirm = () => {
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[700px]">
        <DialogHeader>
          <DialogTitle>Select Token</DialogTitle>
          <DialogDescription>
            Choose a token from the list below
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          {/* Selected Token Tag */}
          {selectedToken && (
            <div className="flex flex-wrap gap-2 mb-2">
              <div className="flex items-center gap-1 bg-muted px-2 py-1 rounded-md">
                <span className="text-sm">{selectedToken.subject || selectedToken.token?.substring(0, 12)}...</span>
                <button
                  type="button"
                  className="text-muted-foreground hover:text-foreground"
                  onClick={handleTagRemove}
                >
                  ×
                </button>
              </div>
            </div>
          )}
          <div className="border rounded-md overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Subject</TableHead>
                  <TableHead>Token</TableHead>
                  <TableHead>SMTP Server</TableHead>
                  <TableHead>Active</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tokens.slice((page - 1) * pageSize, page * pageSize).map(token => (
                  <TableRow
                    key={token.id}
                    className={`cursor-pointer ${selectedTokenId === token.id ? 'bg-muted' : ''}`}
                    onClick={() => handleRowClick(token)}
                  >
                    <TableCell>{token.subject || 'No Subject'}</TableCell>
                    <TableCell>{token.token?.substring(0, 12)}...</TableCell>
                    <TableCell>SMTP ID: {token.SMTPId}</TableCell>
                    <TableCell>{token.isUse ? 'Active' : 'Inactive'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
          {/* Pagination */}
          <Pagination
            currentPage={page}
            totalItems={tokens.length}
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
