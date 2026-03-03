export interface SMTPConfig {
  id?: number;
  host: string;
  port: number;
  mail: string;
  username: string;
  password?: string;
  name?: string;
  createTime?: string;
  updateTime?: string;
}

export interface SMTPListResponse {
  code: number;
  data: SMTPConfig[];
  message: string;
}

export interface SMTPResponse {
  code: number;
  data: SMTPConfig;
  message: string;
}