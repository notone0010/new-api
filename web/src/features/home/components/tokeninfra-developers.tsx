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
import { ArrowUpRight } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { CODE_SAMPLES, type CodeLanguage } from '../lib/code-samples'

export function TokenInfraDevelopers() {
  const { t } = useTranslation()
  const [language, setLanguage] = useState<CodeLanguage>('Python')
  const [copyState, setCopyState] = useState<'idle' | 'copied' | 'error'>(
    'idle'
  )

  const steps = [
    {
      title: t('Create an account, get an API key'),
      text: t('Manage your access keys in the console.'),
    },
    {
      title: t('Configure the endpoint and model'),
      text: t('Set the base URL and choose a model ID.'),
    },
    {
      title: t('Start calling and track usage'),
      text: t('From the first response to continuous iteration.'),
    },
  ]

  async function copyExample() {
    try {
      await navigator.clipboard.writeText(CODE_SAMPLES[language])
      setCopyState('copied')
    } catch {
      setCopyState('error')
    }
  }

  return (
    <section
      className='ti-section ti-developer-section'
      id='developers'
      aria-labelledby='ti-developer-title'
    >
      <div className='ti-developer-copy'>
        <h2 id='ti-developer-title'>
          {t('Easy to start')}
          <br />
          {t('Make your first call today')}
        </h2>
        <ol className='ti-steps'>
          {steps.map((step, index) => (
            <li key={step.title}>
              <span>0{index + 1}</span>
              <div>
                <b>{step.title}</b>
                <p>{step.text}</p>
              </div>
            </li>
          ))}
        </ol>
        <Link to='/docs' className='ti-text-link'>
          {t('Read the integration docs')}
          <ArrowUpRight size={14} aria-hidden='true' />
        </Link>
      </div>

      <div className='ti-code-window'>
        <div className='ti-code-top'>
          <span className='ti-code-dots' aria-hidden='true'>
            <i />
            <i />
            <i />
          </span>
          <span>YOUR FIRST REQUEST</span>
          <button type='button' className='ti-copy' onClick={copyExample}>
            {t('Copy code')}
          </button>
        </div>
        <div className='ti-code-tabs' role='tablist' aria-label={t('Language')}>
          {(Object.keys(CODE_SAMPLES) as CodeLanguage[]).map((item) => (
            <button
              key={item}
              type='button'
              role='tab'
              aria-selected={item === language}
              onClick={() => setLanguage(item)}
            >
              {item}
            </button>
          ))}
        </div>
        <pre tabIndex={0} aria-label={t('API example')}>
          <code>{CODE_SAMPLES[language]}</code>
        </pre>
        <div className='ti-code-foot'>
          <span className='ti-dot' />
          <span>OpenAI-compatible API</span>
          <span role='status'>
            {copyState === 'copied' && t('Copied')}
            {copyState === 'error' &&
              t('Copy failed. Select and copy the code manually.')}
            {copyState === 'idle' && '/v1/chat/completions'}
          </span>
        </div>
      </div>
    </section>
  )
}
