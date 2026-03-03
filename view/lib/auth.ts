'use server';

import { SystemInfo, LoginResponse } from '@/types/auth';
import { ApiClient } from './client-auth';

export const USER_TOKEN_COOKIE = 'user_token';

// ==========================================
// 端口说明：
// - 管理端口 (12566): Web UI 管理界面 + 管理 API (需要登录认证)
// - API 端口 (12567): 公共邮件发送 API (使用 Token 认证)
// ==========================================

// Server actions - only async functions allowed

// Server action for system info
export async function getSystemInfo(): Promise<SystemInfo> {
  const client = new ApiClient();
  return client.getSystemInfo();
}

// Server action for login
export async function login(username: string, password: string): Promise<LoginResponse> {
  const client = new ApiClient();
  return client.login(username, password);
}

// Server action for logout
export async function logout(): Promise<LoginResponse> {
  const client = new ApiClient();
  return client.logout();
}
