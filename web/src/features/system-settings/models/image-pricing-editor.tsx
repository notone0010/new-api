import { useRef } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

import type { ImagePriceConfig } from './image-pricing'

export function ImagePricingEditor(props: {
  value: string
  onChange: (value: string) => void
}) {
  const { t } = useTranslation()
  const rowIds = useRef<string[]>([])
  const nextRowId = useRef(0)
  let config: ImagePriceConfig = { base_price: '' }
  let legacy = false
  try {
    const parsed = JSON.parse(props.value || '{}') as ImagePriceConfig
    legacy = Array.isArray(parsed)
    if (!legacy && parsed && typeof parsed === 'object') config = parsed
  } catch {
    /* Invalid drafts stay invalid until edited. */
  }
  const tiers = Array.isArray(config.pixel_tiers) ? config.pixel_tiers : []
  while (rowIds.current.length < tiers.length) {
    rowIds.current.push(String(nextRowId.current++))
  }
  const update = (next: ImagePriceConfig) =>
    props.onChange(JSON.stringify(next))
  if (legacy) {
    return (
      <div className='space-y-3'>
        <p>
          {t(
            'Legacy image prices are preserved. Replace them to configure a base price and pixel tiers.'
          )}
        </p>
        <Button type='button' onClick={() => update({ base_price: '' })}>
          {t('Replace image pricing')}
        </Button>
      </div>
    )
  }
  return (
    <div className='space-y-5'>
      <section className='space-y-3 rounded-lg border p-4'>
        <h4 className='font-medium'>{t('Image price')}</h4>
        <label className='block space-y-2'>
          <span>{t('Base price (USD/image)')}</span>
          <Input
            type='number'
            min='0'
            step='any'
            value={config.base_price ?? ''}
            onChange={(e) =>
              update({
                ...config,
                base_price: e.target.value === '' ? '' : Number(e.target.value),
              })
            }
          />
        </label>
      </section>
      <section className='space-y-3 rounded-lg border p-4'>
        <h4 className='font-medium'>{t('Billing rules')}</h4>
        {!tiers.length ? (
          <Button
            type='button'
            variant='outline'
            onClick={() =>
              update({
                ...config,
                pixel_tiers: [
                  { max_pixels: 1048576, multiplier: 1 },
                  { max_pixels: null, multiplier: 1 },
                ],
              })
            }
          >
            {t('Add billing rule')}
          </Button>
        ) : (
          <>
            <div className='flex items-center justify-between gap-3'>
              <label className='flex items-center gap-2'>
                <span>{t('Rule type')}</span>
                <select
                  aria-label={t('Rule type')}
                  value='pixels'
                  className='bg-background rounded-md border p-2'
                  onChange={() => {}}
                >
                  <option value='pixels'>{t('Pixels')}</option>
                  <option value='other' disabled>
                    {t('Other (coming soon)')}
                  </option>
                </select>
              </label>
              <Button
                type='button'
                variant='ghost'
                onClick={() => update({ ...config, pixel_tiers: [] })}
              >
                {t('Remove rule')}
              </Button>
            </div>
            <p className='text-muted-foreground text-sm'>
              {t(
                'Pixels = request width × height. 1K = 1,048,576 pixels. Each image uses one tier; upper bounds are inclusive.'
              )}
            </p>
            {tiers.map((tier, index) => (
              <div
                key={rowIds.current[index]}
                className='grid gap-3 rounded-md border p-3 sm:grid-cols-2'
              >
                <label className='space-y-2'>
                  <span>
                    {t('Pixel upper bound')} · {index + 1}
                  </span>
                  {tier.max_pixels === null ? (
                    <p className='py-2'>{t('Above the previous tier')}</p>
                  ) : (
                    <Input
                      aria-label={`${t('Pixel upper bound')} ${index + 1}`}
                      type='number'
                      min='1'
                      step='1'
                      value={tier.max_pixels}
                      onChange={(e) =>
                        update({
                          ...config,
                          pixel_tiers: tiers.map((row, i) =>
                            i === index
                              ? {
                                  ...row,
                                  max_pixels:
                                    e.target.value === ''
                                      ? ''
                                      : Number(e.target.value),
                                }
                              : row
                          ),
                        })
                      }
                    />
                  )}
                </label>
                <label className='space-y-2'>
                  <span>
                    {t('Price multiplier')} · {index + 1}
                  </span>
                  <Input
                    aria-label={`${t('Price multiplier')} ${index + 1}`}
                    type='number'
                    min='0'
                    step='any'
                    value={tier.multiplier}
                    onChange={(e) =>
                      update({
                        ...config,
                        pixel_tiers: tiers.map((row, i) =>
                          i === index
                            ? {
                                ...row,
                                multiplier:
                                  e.target.value === ''
                                    ? ''
                                    : Number(e.target.value),
                              }
                            : row
                        ),
                      })
                    }
                  />
                </label>
                <p className='text-muted-foreground text-sm'>
                  {t('Price per image')}: $
                  {Number(config.base_price) * Number(tier.multiplier)}
                </p>
                {tier.max_pixels !== null && (
                  <Button
                    type='button'
                    variant='ghost'
                    onClick={() => {
                      rowIds.current.splice(index, 1)
                      update({
                        ...config,
                        pixel_tiers: tiers.filter((_, i) => i !== index),
                      })
                    }}
                  >
                    {t('Remove tier')}
                  </Button>
                )}
              </div>
            ))}
            <Button
              type='button'
              variant='outline'
              onClick={() => {
                const last = tiers.at(-1)
                if (!last) return
                rowIds.current.splice(
                  tiers.length - 1,
                  0,
                  String(nextRowId.current++)
                )
                update({
                  ...config,
                  pixel_tiers: [
                    ...tiers.slice(0, -1),
                    { max_pixels: '', multiplier: 1 },
                    last,
                  ],
                })
              }}
            >
              {t('Add tier')}
            </Button>
            <p className='text-muted-foreground text-sm'>
              {t(
                '0.8 charges 80% of the base price; 1 is full price. Values above 1 increase the price.'
              )}
            </p>
          </>
        )}
      </section>
    </div>
  )
}
