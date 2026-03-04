'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useTranslations } from 'next-intl';
import {
  LayoutDashboard,
  Mail,
  Key,
  Settings,
  LogOut,
  ChevronRight,
  Menu,
  Users,
  Clock,
  FileText,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { apiClient } from '@/lib/client-auth';
import { useRouter } from 'next/navigation';
import { LanguageSwitcher } from '@/lib/LanguageSwitcher';

interface DashboardLayoutProps {
  children: React.ReactNode;
}

interface NavigationItem {
  title: string;
  href: string;
  icon: React.ReactNode;
}

export default function DashboardLayout({ children }: DashboardLayoutProps) {
  const pathname = usePathname();
  const router = useRouter();
  const t = useTranslations('common');

  const navigationItems: NavigationItem[] = [
    {
      title: t('dashboard'),
      href: '/',
      icon: <LayoutDashboard className="h-5 w-5" />,
    },
    {
      title: t('smtpServers'),
      href: '/dashboard/smtp',
      icon: <Mail className="h-5 w-5" />,
    },
    {
      title: t('mail'),
      href: '/dashboard/mail',
      icon: <Users className="h-5 w-5" />,
    },
    {
      title: t('apiTokens'),
      href: '/dashboard/tokens',
      icon: <Key className="h-5 w-5" />,
    },
    {
      title: t('schedules'),
      href: '/dashboard/schedule',
      icon: <Clock className="h-5 w-5" />,
    },
    {
      title: t('logs'),
      href: '/dashboard/log',
      icon: <FileText className="h-5 w-5" />,
    },
    {
      title: t('settings'),
      href: '/dashboard/settings',
      icon: <Settings className="h-5 w-5" />,
    },
  ];

  const handleLogout = async () => {
    try {
      await apiClient.logout();
      router.push('/auth/login');
    } catch (err) {
      console.error('Logout error:', err);
    }
  };

  const isActive = (href: string) => {
    if (href === '/') {
      return pathname === '/';
    }
    return pathname.startsWith(href);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="flex">
        {/* Sidebar */}
        <div className="fixed inset-y-0 left-0 z-50 w-64 bg-white border-r border-gray-200">
          <div className="flex flex-col h-full">
            {/* Logo */}
            <div className="flex items-center justify-center h-16 px-4 bg-blue-600">
              <h1 className="text-xl font-bold text-white">HTTP2SMTP</h1>
            </div>

            {/* Navigation */}
            <nav className="flex-1 px-2 py-4 space-y-2">
              {navigationItems.map((item) => (
                <Link key={item.href} href={item.href}>
                  <div
                    className={`flex items-center px-4 py-2 text-sm font-medium rounded-md transition-colors ${
                      isActive(item.href)
                        ? 'bg-blue-50 text-blue-700'
                        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                    }`}
                  >
                    {item.icon}
                    <span className="ml-3">{item.title}</span>
                    {isActive(item.href) && (
                      <ChevronRight className="ml-auto h-4 w-4" />
                    )}
                  </div>
                </Link>
              ))}
            </nav>

            {/* Logout */}
            <div className="px-4 py-4 border-t border-gray-200">
              <Button
                variant="destructive"
                className="w-full gap-2"
                onClick={handleLogout}
              >
                <LogOut className="h-4 w-4" />
                {t('logout')}
              </Button>
            </div>
          </div>
        </div>

        {/* Main content */}
        <div className="flex-1 ml-64">
          {/* Header */}
          <header className="sticky top-0 z-40 bg-white border-b border-gray-200">
            <div className="px-4 sm:px-6 lg:px-8">
              <div className="flex items-center justify-between h-16">
                <div className="flex items-center">
                  {/* Mobile menu button */}
                  <div className="flex lg:hidden">
                    <button
                      type="button"
                      className="inline-flex items-center justify-center p-2 rounded-md text-gray-700 hover:text-gray-900 hover:bg-gray-100"
                    >
                      <Menu className="h-6 w-6" />
                    </button>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  {/* Language Switcher */}
                  <LanguageSwitcher />
                </div>
              </div>
            </div>
          </header>

          {/* Page content */}
          <main className="px-4 sm:px-6 lg:px-8 py-8">
            {children}
          </main>
        </div>
      </div>
    </div>
  );
}