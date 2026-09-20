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
import { describe, expect, it } from 'vitest'

import {
  buildBillingExportParams,
  parseBillingExportFilename,
} from '../../billing-export-api'

describe('billing export filename', () => {
  it('prefers and decodes the UTF-8 content-disposition filename', () => {
    expect(
      parseBillingExportFilename(
        "attachment; filename=bill.csv; filename*=UTF-8''billing-%E5%AE%A2%E6%88%B7.csv"
      )
    ).toBe('billing-客户.csv')
  })

  it('falls back safely when the response omits a filename', () => {
    expect(parseBillingExportFilename('')).toBe('billing-export.csv')
  })
})

describe('billing export parameters', () => {
  it('sends server-local calendar dates and the selected settlement options', () => {
    expect(
      buildBillingExportParams({
        userId: 42,
        startDate: '2026-09-01',
        endDate: '2026-09-30',
        currency: 'CNY',
        exchangeRate: '7.2',
      })
    ).toEqual({
      user_id: 42,
      start_date: '2026-09-01',
      end_date: '2026-09-30',
      currency: 'CNY',
      exchange_rate: '7.2',
    })
  })
})
