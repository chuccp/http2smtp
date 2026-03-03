'use client';

import { Button } from '@/components/ui/button';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PaginationProps {
  pageNo: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
}

export function Pagination({ pageNo, pageSize, total, onPageChange }: PaginationProps) {
  const totalPages = Math.ceil(total / pageSize);

  if (totalPages <= 1) return null;

  const start = (pageNo - 1) * pageSize + 1;
  const end = Math.min(pageNo * pageSize, total);

  return (
    <div className="flex items-center justify-between px-2 py-4">
      <div className="text-sm text-gray-500">
        Showing {start} to {end} of {total} results
      </div>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(pageNo - 1)}
          disabled={pageNo <= 1}
        >
          <ChevronLeft className="h-4 w-4" />
          Previous
        </Button>
        <div className="flex items-center gap-1">
          {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
            let pageNum;
            if (totalPages <= 5) {
              pageNum = i + 1;
            } else if (pageNo <= 3) {
              pageNum = i + 1;
            } else if (pageNo >= totalPages - 2) {
              pageNum = totalPages - 4 + i;
            } else {
              pageNum = pageNo - 2 + i;
            }

            return (
              <Button
                key={pageNum}
                variant={pageNum === pageNo ? 'default' : 'outline'}
                size="sm"
                className="h-8 w-8 p-0"
                onClick={() => onPageChange(pageNum)}
              >
                {pageNum}
              </Button>
            );
          })}
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(pageNo + 1)}
          disabled={pageNo >= totalPages}
        >
          Next
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}