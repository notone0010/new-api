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
import { api } from '@/lib/api'

import type { BillingExportRequest } from './types'

export function parseBillingExportFilename(disposition: string): string {
  const utf8Match = disposition.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8Match) return decodeURIComponent(utf8Match[1])
  const filenameMatch = disposition.match(/filename="?([^";]+)"?/i)
  return filenameMatch?.[1] || 'billing-export.csv'
}

export function buildBillingExportParams(request: BillingExportRequest) {
  return {
    user_id: request.userId,
    start_date: request.startDate,
    end_date: request.endDate,
    currency: request.currency,
    exchange_rate: request.exchangeRate,
  }
}

export async function downloadBillingExport(
  request: BillingExportRequest
): Promise<void> {
  const response = await api.get('/api/log/billing/export', {
    params: buildBillingExportParams(request),
    responseType: 'blob',
  })
  const disposition = String(response.headers['content-disposition'] || '')
  const filename = parseBillingExportFilename(disposition)
  const objectUrl = URL.createObjectURL(response.data)
  const anchor = document.createElement('a')
  try {
    anchor.href = objectUrl
    anchor.download = filename
    document.body.appendChild(anchor)
    anchor.click()
  } finally {
    anchor.remove()
    URL.revokeObjectURL(objectUrl)
  }
}
