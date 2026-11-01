/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { ThemeSwitch } from '@/components/theme-switch'
import { Skeleton } from '@/components/ui/skeleton'
import { useTheme } from '@/context/theme-provider'
import { useSystemConfig } from '@/hooks/use-system-config'
import { DEFAULT_LOGO, DEFAULT_LOGO_DARK } from '@/lib/constants'
import { cn } from '@/lib/utils'

import '@/styles/tokeninfra-home.css'

type AuthLayoutProps = {
  children: React.ReactNode
}

function BrandMark({
  logo,
  systemName,
  loading,
  className,
}: {
  logo: string
  systemName: string
  loading: boolean
  className?: string
}) {
  const { t } = useTranslation()
  const { resolvedTheme } = useTheme()
  const displayLogo = logo === DEFAULT_LOGO && resolvedTheme === 'dark'
    ? DEFAULT_LOGO_DARK
    : logo

  return (
    <Link
      to='/'
      className={cn(
        'flex items-center gap-2.5 transition-opacity hover:opacity-80',
        className
      )}
    >
      <div
        className={cn(
          'relative h-7',
          logo === DEFAULT_LOGO ? 'w-32' : 'w-7 shrink-0'
        )}
      >
        {loading ? (
          <Skeleton className='absolute inset-0 rounded-md' />
        ) : (
          <img
            src={displayLogo}
            alt={t('Logo')}
            className='h-full w-full object-contain'
          />
        )}
      </div>
      {loading ? (
        <Skeleton className='h-5 w-20' />
      ) : (
        <span
          className={cn(
            'text-lg font-semibold tracking-tight',
            logo === DEFAULT_LOGO && 'sr-only'
          )}
        >
          {systemName}
        </span>
      )}
    </Link>
  )
}

function AuthFlow() {
  return (
    <div className='ti-auth-flow' aria-hidden='true'>
      <svg viewBox='0 0 520 150' fill='none'>
        <defs>
          <linearGradient id='ti-flow-gradient' x1='0' y1='0' x2='1' y2='0'>
            <stop offset='0' stopColor='#13b58a' stopOpacity='0' />
            <stop offset='0.45' stopColor='#13b58a' />
            <stop offset='1' stopColor='#a3d6ff' stopOpacity='0' />
          </linearGradient>
        </defs>
        <path
          className='ti-flow-track'
          d='M4 26H132Q158 26 174 48L226 96'
        />
        <path className='ti-flow-track' d='M4 75H190Q222 75 238 75H246' />
        <path
          className='ti-flow-track'
          d='M4 124H132Q158 124 174 102L226 54'
        />
        <path
          className='ti-flow-track'
          d='M274 75H292Q320 75 344 48L394 26H516'
        />
        <path className='ti-flow-track' d='M274 75H516' />
        <path
          className='ti-flow-track'
          d='M274 75H292Q320 75 344 102L394 124H516'
        />
        <path
          className='ti-flow-runner'
          d='M4 26H132Q158 26 174 48L226 96'
        />
        <path
          className='ti-flow-runner two'
          d='M4 75H190Q222 75 238 75H246'
        />
        <path
          className='ti-flow-runner three'
          d='M4 124H132Q158 124 174 102L226 54'
        />
        <rect
          className='ti-flow-hub'
          x='238'
          y='45'
          width='62'
          height='60'
          rx='13'
        />
        <path
          className='ti-flow-hub-icon'
          d='M258 62h22l-12 13h12l-22 14 7-14h-9z'
        />
      </svg>
    </div>
  )
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()
  const { resolvedTheme } = useTheme()

  const advantages = [
    t('Better prices, lower costs'),
    t('Every popular model, one place'),
    t('Visible usage, clear costs'),
    t('Responsive, hands-on support'),
  ]

  return (
    <div className={cn('ti-auth', resolvedTheme)}>
      <header className='ti-auth-header'>
        <BrandMark logo={logo} systemName={systemName} loading={loading} />
        <div className='flex items-center gap-3'>
          <ThemeSwitch />
          <Link to='/' className='ti-auth-back'>
            <ArrowLeft size={15} aria-hidden='true' />
            {t('Back to home')}
          </Link>
        </div>
      </header>

      <main className='ti-auth-main'>
        <section className='ti-auth-story'>
          <h1>
            {t('Making model access universal.')}
            <br />
            <span>{t('A one-stop token platform.')}</span>
          </h1>
          <AuthFlow />
          <ul className='ti-auth-advantages'>
            {advantages.map((item) => (
              <li className='ti-auth-advantage' key={item}>
                {item}
              </li>
            ))}
          </ul>
        </section>

        <section className='ti-auth-panel'>{children}</section>
      </main>

      <footer className='ti-auth-footer'>
        © {new Date().getFullYear()} {systemName || 'TokenInfra'}
      </footer>
    </div>
  )
}
