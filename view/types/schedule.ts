export interface ScheduleHeader {
  name: string;
  value: string;
}

export interface ScheduleConfig {
  id?: number;
  name: string;
  token: string;
  tokenId?: number;
  cron: string;
  url: string;
  method: string;
  headerStr?: string;
  headers?: ScheduleHeader[];
  body?: string;
  useTemplate: boolean;
  template?: string;
  isUse: boolean;
  isSendOnlyByError?: boolean;
  createTime?: string;
  updateTime?: string;
}