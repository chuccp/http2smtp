import { LoginUser, SystemInfo, LoginResponse } from '@/types/auth';
import { SMTPConfig } from '@/types/smtp';
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
  async getSMTPServers(): Promise<SMTPConfig[]> {
    const response = await fetch(`${this.manageBaseUrl}/smtp`, {
      credentials: 'include',
    });
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    const responseData = await response.json();
    return responseData.data || responseData;
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
    const responseData = await response.json();
    return responseData.data || responseData;
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
    const responseData = await response.json();
    return responseData.data || responseData;
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
}

// Create default client
export const apiClient = new ApiClient();
