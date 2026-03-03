export interface TokenConfig {
  id?: number;
  token: string;
  name?: string;
  subject: string;
  receiveEmailIds: string;
  receiveEmails?: { id: number; name: string; mail: string }[];
  receiveEmailsStr?: string;
  SMTPId: number;
  SMTP?: { id: number; name: string; host: string };
  SMTPStr?: string;
  isUse: boolean;
  createTime?: string;
  updateTime?: string;
}

export interface TokenListResponse {
  code: number;
  data: {
    total: number;
    list: TokenConfig[];
  };
  message: string;
}

export interface TokenResponse {
  code: number;
  data: TokenConfig;
  message: string;
}