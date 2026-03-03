import crypto from 'crypto';

// MD5 hashing function matching backend
export function md5(str: string): string {
  return crypto.createHash('md5').update(str).digest('hex');
}

// Generate random string matching backend's GenerateRandomStringByAlphanumeric
export function generateRandomString(length: number): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}