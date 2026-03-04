import { LoginUser, SystemInfo, LoginResponse } from '@/types/auth';
import { SMTPConfig } from '@/types/smtp';
import { MailConfig } from '@/types/mail';
import { SetInfo } from '@/types/settings';
import { TokenConfig } from '@/types/token';
import { ScheduleConfig } from '@/types/schedule';
import { LogEntry } from '@/types/log';
import { md5, generateRandomString } from './auth-utils';

export const USER_TOKEN_COOKIE = 'user_token';

// ==========================================
// 端口说明：
// - 管理端口 (12566): Web UI 管理界面 + 管理 API (需要登录认证)
// - API 端口 (12567): 公共邮件发送 API (使用 Token 认证)
// ==========================================

// API client for client-side usage
export class ApiClient {
  private manageBaseUrl: string;
  private sendMailBaseUrl: string;

  constructor(
    manageBaseUrl: string = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:12566',
    sendMailBaseUrl: string = process.env.NEXT_PUBLIC_SENDMAIL_API_URL || 'http://localhost:12567'
  ) {
    this.manageBaseUrl = manageBaseUrl;
    this.sendMailBaseUrl = sendMailBaseUrl;
  }

  // SMTP Server Management
  async getSMTPServers(pageNo: number = 1, pageSize: number = 10): Promise<{ list: SMTPConfig[]; total: number }> {
    const response = await fetch(`${this.manageBaseUrl}/smtp?pageNo=${pageNo}&pageSize=${pageSize}`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    // Backend returns { code, data: { total, list } } format
    const data = responseData.data || responseData;
    // Handle PageAble format: { total, list }
    if (data && data.list) {
      return { list: data.list, total: data.total || 0 };
    }
    // Fallback for direct array response
    return { list: Array.isArray(data) ? data : [], total: Array.isArray(data) ? data.length : 0 };
  }

  async getSMTPServer(id: number): Promise<SMTPConfig> {
    const response = await fetch(`${this.manageBaseUrl}/smtp/${id}`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    return responseData.data || responseData;
  }

  async createSMTPServer(server: SMTPConfig): Promise<SMTPConfig> {
    const response = await fetch(`${this.manageBaseUrl}/smtp`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(server),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to create SMTP server');
    }
    // Backend may return "ok" as plain text or JSON
    const text = await response.text();
    if (text === 'ok' || text === '"ok"') {
      return server;
    }
    try {
      const responseData = JSON.parse(text);
      return responseData.data || responseData;
    } catch {
      return server;
    }
  }

  async updateSMTPServer(server: SMTPConfig): Promise<SMTPConfig> {
    const response = await fetch(`${this.manageBaseUrl}/smtp`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(server),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to update SMTP server');
    }
    // Backend may return "ok" as plain text or JSON
    const text = await response.text();
    if (text === 'ok' || text === '"ok"') {
      return server;
    }
    try {
      const responseData = JSON.parse(text);
      return responseData.data || responseData;
    } catch {
      return server;
    }
  }

  async deleteSMTPServer(id: number): Promise<void> {
    const response = await fetch(`${this.manageBaseUrl}/smtp/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
  }

  async testSMTPServer(server: SMTPConfig): Promise<{ code: number; data: string; message: string }> {
    const response = await fetch(`${this.manageBaseUrl}/test`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(server),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    return responseData;
  }

  async sendMailBySMTP(sendMail: any): Promise<{ code: number; data: string; message: string }> {
    const response = await fetch(`${this.manageBaseUrl}/sendMailBySMTP`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(sendMail),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    return responseData;
  }

  // 获取管理 API 地址 (端口 12566)
  getManageBaseUrl(): string {
    return this.manageBaseUrl;
  }

  // 获取邮件发送 API 地址 (端口 12567)
  getSendMailBaseUrl(): string {
    return this.sendMailBaseUrl;
  }

  async getSystemInfo(): Promise<SystemInfo> {
    const response = await fetch(`${this.manageBaseUrl}/set`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    // Handle wrapped response format: { code: 200, data: {...}, msg: "ok" }
    if (responseData.data && responseData.data.hasInit !== undefined) {
      return responseData.data;
    }
    // Fallback for direct response format
    if (responseData.hasInit !== undefined) {
      return responseData;
    }
    // Default fallback
    return { hasInit: false, hasLogin: false, isDocker: false };
  }

  async login(username: string, password: string): Promise<LoginResponse> {
    // First get system info to check if initialized
    const systemInfo = await this.getSystemInfo();

    if (!systemInfo.hasInit) {
      return { code: 400, data: '', message: 'System not initialized' };
    }

    // Generate nonce
    const nonce = generateRandomString(16);

    // Calculate signature - this should match the backend's expected signature
    // Setup saves: manage.Password = MD5(plain_password)
    // Backend uses: key = MD5(manage.Password + manage.Username) = MD5(MD5(plain) + username)
    // Then sign = MD5(key + nonce)
    const key = md5(md5(password) + username);
    const responseSignature = md5(key + nonce);

    const loginData: LoginUser = {
      username,
      response: responseSignature,
      nonce,
    };

    const response = await fetch(`${this.manageBaseUrl}/signIn`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(loginData),
      credentials: 'include',
    });

    if (response.ok) {
      return { code: 200, data: 'success', message: 'Login successful' };
    } else {
      return { code: response.status, data: '', message: 'Invalid credentials' };
    }
  }

  async logout(): Promise<LoginResponse> {
    const response = await fetch(`${this.manageBaseUrl}/logout`, {
      method: 'POST',
      credentials: 'include',
    });

    if (response.ok) {
      return { code: 200, data: 'success', message: 'Logout successful' };
    } else {
      return { code: response.status, data: '', message: 'Logout failed' };
    }
  }

  // Mail Address Management
  async getMails(pageNo: number = 1, pageSize: number = 10): Promise<{ list: MailConfig[]; total: number }> {
    const response = await fetch(`${this.manageBaseUrl}/mail?pageNo=${pageNo}&pageSize=${pageSize}`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    const data = responseData.data || responseData;
    if (data && data.list) {
      return { list: data.list, total: data.total || 0 };
    }
    return { list: Array.isArray(data) ? data : [], total: Array.isArray(data) ? data.length : 0 };
  }

  async getMail(id: number): Promise<MailConfig> {
    const response = await fetch(`${this.manageBaseUrl}/mail/${id}`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    return responseData.data || responseData;
  }

  async createMail(mail: MailConfig): Promise<MailConfig> {
    const response = await fetch(`${this.manageBaseUrl}/mail`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(mail),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to create mail');
    }
    const text = await response.text();
    if (text === 'ok' || text === '"ok"') {
      return mail;
    }
    try {
      const responseData = JSON.parse(text);
      return responseData.data || responseData;
    } catch {
      return mail;
    }
  }

  async updateMail(mail: MailConfig): Promise<MailConfig> {
    const response = await fetch(`${this.manageBaseUrl}/mail`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(mail),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to update mail');
    }
    const text = await response.text();
    if (text === 'ok' || text === '"ok"') {
      return mail;
    }
    try {
      const responseData = JSON.parse(text);
      return responseData.data || responseData;
    } catch {
      return mail;
    }
  }

  async deleteMail(id: number): Promise<void> {
    const response = await fetch(`${this.manageBaseUrl}/mail/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
  }

  // Settings Management
  async getSettings(): Promise<SetInfo> {
    const response = await fetch(`${this.manageBaseUrl}/readSet`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    return responseData.data || responseData;
  }

  async getDefaultSettings(): Promise<SetInfo> {
    const response = await fetch(`${this.manageBaseUrl}/defaultSet`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    return responseData.data || responseData;
  }

  async updateSettings(settings: SetInfo): Promise<void> {
    const response = await fetch(`${this.manageBaseUrl}/reSet`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(settings),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to update settings');
    }
  }

  async restart(): Promise<void> {
    const response = await fetch(`${this.manageBaseUrl}/reStart`, {
      method: 'POST',
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to restart');
    }
  }

  // Token Management
  async getTokens(pageNo: number = 1, pageSize: number = 10): Promise<{ list: TokenConfig[]; total: number }> {
    const response = await fetch(`${this.manageBaseUrl}/token?pageNo=${pageNo}&pageSize=${pageSize}`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    const data = responseData.data || responseData;
    if (data && data.list) {
      return { list: data.list, total: data.total || 0 };
    }
    return { list: Array.isArray(data) ? data : [], total: Array.isArray(data) ? data.length : 0 };
  }

  async getToken(id: number): Promise<TokenConfig> {
    const response = await fetch(`${this.manageBaseUrl}/token/${id}`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    return responseData.data || responseData;
  }

  async createToken(token: TokenConfig): Promise<TokenConfig> {
    const response = await fetch(`${this.manageBaseUrl}/token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(token),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to create token');
    }
    const text = await response.text();
    if (text === 'ok' || text === '"ok"') {
      return token;
    }
    try {
      const responseData = JSON.parse(text);
      return responseData.data || responseData;
    } catch {
      return token;
    }
  }

  async updateToken(token: TokenConfig): Promise<TokenConfig> {
    const response = await fetch(`${this.manageBaseUrl}/token`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(token),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to update token');
    }
    const text = await response.text();
    if (text === 'ok' || text === '"ok"') {
      return token;
    }
    try {
      const responseData = JSON.parse(text);
      return responseData.data || responseData;
    } catch {
      return token;
    }
  }

  async deleteToken(id: number): Promise<void> {
    const response = await fetch(`${this.manageBaseUrl}/token/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
  }

  // Schedule Management
  async getSchedules(pageNo: number = 1, pageSize: number = 10): Promise<{ list: ScheduleConfig[]; total: number }> {
    const response = await fetch(`${this.manageBaseUrl}/schedule?pageNo=${pageNo}&pageSize=${pageSize}`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    const data = responseData.data || responseData;
    if (data && data.list) {
      return { list: data.list, total: data.total || 0 };
    }
    return { list: Array.isArray(data) ? data : [], total: Array.isArray(data) ? data.length : 0 };
  }

  async getSchedule(id: number): Promise<ScheduleConfig> {
    const response = await fetch(`${this.manageBaseUrl}/schedule/${id}`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    return responseData.data || responseData;
  }

  async createSchedule(schedule: ScheduleConfig): Promise<ScheduleConfig> {
    const response = await fetch(`${this.manageBaseUrl}/schedule`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(schedule),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to create schedule');
    }
    const text = await response.text();
    if (text === 'ok' || text === '"ok"') {
      return schedule;
    }
    try {
      const responseData = JSON.parse(text);
      return responseData.data || responseData;
    } catch {
      return schedule;
    }
  }

  async updateSchedule(schedule: ScheduleConfig): Promise<ScheduleConfig> {
    const response = await fetch(`${this.manageBaseUrl}/schedule`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(schedule),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to update schedule');
    }
    const text = await response.text();
    if (text === 'ok' || text === '"ok"') {
      return schedule;
    }
    try {
      const responseData = JSON.parse(text);
      return responseData.data || responseData;
    } catch {
      return schedule;
    }
  }

  async deleteSchedule(id: number): Promise<void> {
    const response = await fetch(`${this.manageBaseUrl}/schedule/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
  }

  async sendMailBySchedule(schedule: ScheduleConfig): Promise<void> {
    const response = await fetch(`${this.manageBaseUrl}/sendMailBySchedule`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(schedule),
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      throw new Error('Failed to send mail');
    }
  }

  async sendMailByToken(token: string, subject: string, content: string, recipients?: string[]): Promise<{ code: number; data: string; message: string }> {
    const response = await fetch(`${this.sendMailBaseUrl}/sendMail`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        token,
        subject,
        content,
        recipients,
      }),
    });
    const responseData = await response.json();
    if (response.ok) {
      return { code: 200, data: 'ok', message: 'Success' };
    }
    return { code: response.status, data: 'error', message: responseData };
  }

  // Log Management
  async getLogs(pageNo: number = 1, pageSize: number = 10, searchKey: string = ''): Promise<{ list: LogEntry[]; total: number }> {
    const response = await fetch(`${this.manageBaseUrl}/log?pageNo=${pageNo}&pageSize=${pageSize}&searchKey=${encodeURIComponent(searchKey)}`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    const data = responseData.data || responseData;
    if (data && data.list) {
      return { list: data.list, total: data.total || 0 };
    }
    return { list: Array.isArray(data) ? data : [], total: Array.isArray(data) ? data.length : 0 };
  }

  async getLog(id: number): Promise<LogEntry> {
    const response = await fetch(`${this.manageBaseUrl}/log/${id}`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    return responseData.data || responseData;
  }
}

// Create default client
export const apiClient = new ApiClient();
