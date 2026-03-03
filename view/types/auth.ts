export interface LoginUser {
  username: string;
  response: string;
  nonce: string;
}

export interface SessionUser {
  username: string;
  // Add other user fields as needed
}

export interface SystemInfo {
  hasInit: boolean;
  hasLogin: boolean;
  isDocker: boolean;
}

export interface LoginResponse {
  code: number;
  data: string;
  message: string;
}

export interface ManageConfig {
  port?: number;
  webPath?: string;
  username?: string;
  password?: string;
  confirmPassword?: string;
}

export interface ApiConfig {
  port?: number;
}

export interface SqliteConfig {
  filename?: string;
}

export interface MysqlConfig {
  host?: string;
  port?: number;
  dbname?: string;
  username?: string;
  password?: string;
  charset?: string;
}

export interface ScheduleConfig {
  start?: boolean;
}

export interface SetInfo {
  hasInit?: boolean;
  dbType?: string;
  cachePath?: string;
  sqlite?: SqliteConfig;
  mysql?: MysqlConfig;
  manage?: ManageConfig;
  api?: ApiConfig;
  isDocker?: boolean;
  schedule?: ScheduleConfig;
}
