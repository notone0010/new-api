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
import { ArrowRight, ArrowUpRight, LayoutGrid } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { TokenFactoryVisual, TokenInfraMark } from './token-factory-visual'
import { TokenInfraDevelopers } from './tokeninfra-developers'

import '@/styles/tokeninfra-home.css'

interface TokenInfraLandingProps {
  isAuthenticated: boolean
}

const MODEL_FILTERS = ['all', 'reasoning', 'coding', 'multi'] as const

type ModelFilter = (typeof MODEL_FILTERS)[number]

type AdvantageLink = {
  label: string
  to?: '/pricing' | '/dashboard'
  hash?: string
  mail?: string
}

type Advantage = {
  title: string
  text: string
  link: AdvantageLink
}

const MODELS = [
  {
    name: 'DeepSeek',
    mark: 'DS',
    tone: 'blue',
    categories: ['reasoning', 'coding', 'multi'],
    tags: 'Reasoning · Coding · Multimodal',
    description:
      'Reasoning and coding coverage; Flash accepts images, V4-Pro is text-only.',
  },
  {
    name: 'Kimi',
    mark: 'K',
    tone: 'mint',
    categories: ['reasoning', 'coding', 'multi'],
    tags: 'Long context · Reasoning · Coding · Multimodal',
    description:
      'Long documents, code and visual understanding; multimodal models accept images and video.',
  },
  {
    name: 'GLM',
    mark: 'GLM',
    tone: 'purple',
    categories: ['reasoning', 'coding', 'multi'],
    tags: 'Reasoning · Coding · Multimodal',
    description:
      'Language and vision models; the GLM-V series understands images and video.',
  },
  {
    name: 'Doubao',
    mark: 'DB',
    tone: 'blue',
    categories: ['reasoning', 'multi'],
    tags: 'Reasoning · Multimodal',
    description:
      'Content understanding and reasoning; multimodal models accept images and video.',
  },
  {
    name: 'Claude',
    mark: 'C',
    tone: 'peach',
    categories: ['reasoning', 'coding', 'multi'],
    tags: 'Reasoning · Coding · Long context · Multimodal',
    description:
      'Code collaboration, long-document analysis and image understanding.',
  },
] as const

export function TokenInfraLanding(props: TokenInfraLandingProps) {
  const { t } = useTranslation()
  const [filter, setFilter] = useState<ModelFilter>('all')
  const destination = props.isAuthenticated ? '/dashboard' : '/sign-up'
  const action = props.isAuthenticated ? t('Go to Dashboard') : t('Get Started')

  const advantages: Advantage[] = [
    {
      title: t('Better prices, lower costs'),
      text: t(
        'Cut model calling costs and leave more budget for experiments and bigger ambitions.'
      ),
      link: { label: t('View model pricing'), to: '/pricing' },
    },
    {
      title: t('Every popular model, one place'),
      text: t(
        'Aggregate leading models worldwide for reasoning, coding and creation—choose by task.'
      ),
      link: { label: t('Explore the model ecosystem'), hash: '#models' },
    },
    {
      title: t('Visible usage, clear costs'),
      text: t(
        'See usage and spend at a glance, so every budget decision has evidence.'
      ),
      link: { label: t('Open the usage console'), to: '/dashboard' },
    },
    {
      title: t('Responsive, hands-on support'),
      text: t(
        'From onboarding to daily use, we respond quickly and follow through.'
      ),
      link: { label: t('Contact support'), mail: 'support@tokeninfra.com' },
    },
  ]

  const filters: { key: ModelFilter; label: string }[] = [
    { key: 'all', label: t('All models') },
    { key: 'reasoning', label: t('Deep reasoning') },
    { key: 'coding', label: t('Coding') },
    { key: 'multi', label: t('Multimodal') },
  ]

  return (
    <main className='ti-home'>
      <section className='ti-hero' aria-labelledby='ti-hero-title'>
        <div className='ti-hero-grid' aria-hidden='true' />
        <div className='ti-hero-halo' aria-hidden='true' />
        <div className='ti-hero-copy'>
          <h1 id='ti-hero-title'>
            {t('At the best prices')}
            <span>{t('Enjoy global models')}</span>
          </h1>
          <div className='ti-hero-actions'>
            <Link className='ti-btn' to={destination}>
              {action}
              <ArrowUpRight size={17} aria-hidden='true' />
            </Link>
            <a className='ti-btn secondary' href='#models'>
              {t('Explore models')}
              <ArrowRight size={16} aria-hidden='true' />
            </a>
          </div>
        </div>
        <TokenFactoryVisual />
      </section>

      <section
        className='ti-section ti-advantages-section'
        id='advantages'
        aria-labelledby='ti-advantages-title'
      >
        <div className='ti-section-head'>
          <h2 id='ti-advantages-title'>{t('Why TokenInfra')}</h2>
          <p>{t('The lowest model prices, an exceptional experience')}</p>
        </div>
        <div className='ti-advantages'>
          {advantages.map((item, index) => {
            let action = (
              <a className='ti-text-link' href='#models'>
                {item.link.label}
                <ArrowRight size={13} aria-hidden='true' />
              </a>
            )
            if (item.link.to) {
              action = (
                <Link className='ti-text-link' to={item.link.to}>
                  {item.link.label}
                  <ArrowUpRight size={13} aria-hidden='true' />
                </Link>
              )
            } else if (item.link.mail) {
              action = (
                <a className='ti-text-link' href={`mailto:${item.link.mail}`}>
                  {item.link.label}
                  <ArrowUpRight size={13} aria-hidden='true' />
                </a>
              )
            }

            return (
              <article className='ti-advantage' key={item.title}>
                <span>0{index + 1}</span>
                <h3>{item.title}</h3>
                <p>{item.text}</p>
                {action}
              </article>
            )
          })}
        </div>
      </section>

      <section
        className='ti-section ti-models-section'
        id='models'
        aria-labelledby='ti-models-title'
      >
        <div className='ti-section-head'>
          <h2 id='ti-models-title'>{t('One gateway, every model')}</h2>
        </div>
        <div className='ti-model-toolbar'>
          <div className='ti-filters' role='group' aria-label={t('Model type')}>
            {filters.map((item) => (
              <button
                key={item.key}
                type='button'
                aria-pressed={filter === item.key}
                onClick={() => setFilter(item.key)}
              >
                {item.label}
              </button>
            ))}
          </div>
          <Link className='ti-text-link' to='/pricing'>
            {t('Model playground')}
            <ArrowUpRight size={14} aria-hidden='true' />
          </Link>
        </div>
        <div className='ti-model-grid'>
          {MODELS.filter(
            (model) =>
              filter === 'all' ||
              (model.categories as readonly string[]).includes(filter)
          ).map((model) => (
            <Link className='ti-model-card' to='/pricing' key={model.name}>
              <div className='ti-model-top'>
                <span className={`ti-model-icon ${model.tone}`}>
                  {model.mark}
                </span>
                <ArrowUpRight
                  className='ti-model-arrow'
                  size={16}
                  aria-hidden='true'
                />
              </div>
              <h3>{model.name}</h3>
              <span className='ti-model-tags'>{t(model.tags)}</span>
              <p>{t(model.description)}</p>
            </Link>
          ))}
          <Link className='ti-model-card' to='/pricing'>
            <div className='ti-model-top'>
              <span className='ti-model-icon more'>
                <LayoutGrid size={18} aria-hidden='true' />
              </span>
              <ArrowUpRight
                className='ti-model-arrow'
                size={16}
                aria-hidden='true'
              />
            </div>
            <h3>{t('More models')}</h3>
            <span className='ti-model-tags'>
              {t('Explore more possibilities')}
            </span>
            <p>
              {t('More models from around the world, all in the model marketplace.')}
            </p>
          </Link>
        </div>
      </section>

      <TokenInfraDevelopers />

      <section
        className='ti-section ti-enterprise'
        id='enterprise'
        aria-labelledby='ti-enterprise-title'
      >
        <div className='ti-enterprise-mark' aria-hidden='true'>
          <TokenInfraMark />
        </div>
        <h2 id='ti-enterprise-title'>
          {t('Powering customer growth')}
          <br />
          {t('A solid AI foundation')}
        </h2>
        <a className='ti-btn' href='mailto:support@tokeninfra.com'>
          {t('Talk to enterprise sales')}
          <ArrowUpRight size={17} aria-hidden='true' />
        </a>
      </section>
    </main>
  )
}
