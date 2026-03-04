export interface LogFile {
  name: string;
  data: string;
  filePath?: string;
}

export interface LogEntry {
  id: number;
  name: string;
  mail: string;
  token: string;
  smtp: string;
  subject: string;
  content: string;
  files: string;
  fileArray?: LogFile[];
  createTime: string;
  updateTime: string;
  createTimeStr?: string;
  updateTimeStr?: string;
  status: number;
  statusStr: string;
  result: string;
}

export interface LogListResponse {
  code: number;
  data: {
    total: number;
    list: LogEntry[];
  };
  message: string;
}

export interface LogResponse {
  code: number;
  data: LogEntry;
  message: string;
}
