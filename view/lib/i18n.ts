import { getRequestConfig } from 'next-intl/server';
import { cookies } from 'next/headers';

export const locales = ['en', 'zh-CN', 'zh-TW'] as const;
export type Locale = (typeof locales)[number];

export default getRequestConfig(async () => {
  // Try to get locale from cookie first, then default to zh-CN
  const cookieStore = await cookies();
  const locale = cookieStore.get('locale')?.value || 'zh-CN';

  return {
    locale,
    messages: (await import(`../locales/${locale}.json`)).default,
  };
});