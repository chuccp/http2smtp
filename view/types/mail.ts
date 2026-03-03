export interface MailConfig {
  id?: number;
  name: string;
  mail: string;
  createTime?: string;
  updateTime?: string;
}

export interface MailListResponse {
  code: number;
  data: {
    total: number;
    list: MailConfig[];
  };
  message: string;
}

export interface MailResponse {
  code: number;
  data: MailConfig;
  message: string;
}