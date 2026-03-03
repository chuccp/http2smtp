export interface SetInfo {
  hasInit: boolean;
  dbType: string;
  cachePath: string;
  sqlite?: SqliteConfig;
  mysql?: MysqlConfig;
  manage?: ManageConfig;
  api?: ApiConfig;
  isDocker: boolean;
}

export interface SqliteConfig {
  filename: string;
}

export interface MysqlConfig {
  host: string;
  port: number;
  dbname: string;
  username: string;
  password: string;
  charset: string;
}

export interface ManageConfig {
  port: number;
  webPath: string;
  username: string;
  password?: string;
  confirmPassword?: string;
}

export interface ApiConfig {
  port: number;
}

export interface SystemInfo {
  hasInit: boolean;
  hasLogin: boolean;
  isDocker: boolean;
}