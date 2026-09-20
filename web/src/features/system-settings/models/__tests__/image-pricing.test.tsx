import { fireEvent, render, screen } from '@testing-library/react'
import { useState } from 'react'
import { afterEach, describe, expect, test, vi } from 'vitest'

import { validateImagePricing } from '../image-pricing'
import { ImagePricingEditor } from '../image-pricing-editor'
import {
  buildModelSnapshots,
  getModeLabel,
  isBasePricingUnset,
} from '../model-pricing-snapshots'

function EditorFixture() {
  const [value, setValue] = useState('{"base_price":0.1}')
  return (
    <>
      <ImagePricingEditor value={value} onChange={setValue} />
      <output aria-label='configuration'>{value}</output>
    </>
  )
}

describe('image pixel pricing', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  test('adds a billing rule when randomUUID is unavailable', () => {
    vi.stubGlobal('crypto', {})
    render(<EditorFixture />)

    expect(() =>
      fireEvent.click(screen.getByRole('button', { name: 'Add billing rule' }))
    ).not.toThrow()
    expect(
      screen.getByRole('option', { name: 'Other (coming soon)' })
    ).toBeDisabled()
  })

  test('adds one pixel rule, edits its multiplier and removes the rule without losing the base price', () => {
    render(<EditorFixture />)
    fireEvent.click(screen.getByRole('button', { name: 'Add billing rule' }))
    expect(
      screen.queryByRole('button', { name: 'Add billing rule' })
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('option', { name: 'Other (coming soon)' })
    ).toBeDisabled()
    fireEvent.change(
      screen.getByRole('spinbutton', { name: 'Price multiplier 1' }),
      { target: { value: '0.5' } }
    )
    expect(
      JSON.parse(screen.getByLabelText('configuration').textContent || '')
    ).toEqual({
      base_price: 0.1,
      pixel_tiers: [
        { max_pixels: 1048576, multiplier: 0.5 },
        { max_pixels: null, multiplier: 1 },
      ],
    })
    fireEvent.click(screen.getByRole('button', { name: 'Remove rule' }))
    expect(
      JSON.parse(screen.getByLabelText('configuration').textContent || '')
    ).toEqual({ base_price: 0.1, pixel_tiers: [] })
  })

  test('rejects missing prices, overlapping limits, and zero multipliers but accepts free base prices', () => {
    expect(validateImagePricing('{"base_price":0}')).toBe(true)
    expect(validateImagePricing('{"base_price":""}')).toBe(false)
    expect(
      validateImagePricing(
        '{"base_price":1,"pixel_tiers":[{"max_pixels":100,"multiplier":1},{"max_pixels":99,"multiplier":1},{"max_pixels":null,"multiplier":1}]}'
      )
    ).toBe(false)
    expect(
      validateImagePricing(
        '{"base_price":1,"pixel_tiers":[{"max_pixels":null,"multiplier":0}]}'
      )
    ).toBe(false)
    expect(
      validateImagePricing(
        '{"base_price":1,"pixel_tiers":[{"max_pixels":100,"multiplier":0.5},{"max_pixels":null,"multiplier":1.5}]}'
      )
    ).toBe(true)
  })

  test('image-only models appear in the pricing list and are not classified as unpriced', () => {
    const rows = buildModelSnapshots({
      modelPrice: '{}',
      modelRatio: '{}',
      cacheRatio: '{}',
      createCacheRatio: '{}',
      completionRatio: '{}',
      imageRatio: '{}',
      audioRatio: '{}',
      audioCompletionRatio: '{}',
      billingMode: '{}',
      billingExpr: '{}',
      imagePrice: '{"image-model":{"base_price":0.1}}',
    })
    expect(rows).toHaveLength(1)
    expect(rows[0].billingMode).toBe('per-image')
    expect(getModeLabel(rows[0].billingMode)).toBe('Per-image')
    expect(isBasePricingUnset(rows[0])).toBe(false)
  })
})
