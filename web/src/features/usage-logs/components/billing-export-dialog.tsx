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
import { zodResolver } from '@hookform/resolvers/zod'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { z } from 'zod'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { searchUsers } from '@/features/users/api'
import dayjs from '@/lib/dayjs'
import { useSystemConfigStore } from '@/stores/system-config-store'

import { downloadBillingExport } from '../billing-export-api'

const billingExportSchema = z
  .object({
    userId: z.number().int().positive(),
    startDate: z.string().min(1),
    endDate: z.string().min(1),
    currency: z.enum(['USD', 'CNY']),
    exchangeRate: z.string().refine((value) => {
      const rate = Number(value)
      return Number.isFinite(rate) && rate > 0
    }),
  })
  .refine((value) => value.endDate >= value.startDate, {
    path: ['endDate'],
    message: 'End date must not be before start date',
  })

type BillingExportForm = z.infer<typeof billingExportSchema>

interface BillingExportDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function BillingExportDialog(props: BillingExportDialogProps) {
  const { t } = useTranslation()
  const defaultRate = useSystemConfigStore(
    (state) => state.config.currency.usdExchangeRate
  )
  const [keyword, setKeyword] = useState('')
  const form = useForm<BillingExportForm>({
    resolver: zodResolver(billingExportSchema),
    defaultValues: {
      userId: 0,
      startDate: dayjs().subtract(1, 'month').startOf('month').format('YYYY-MM-DD'),
      endDate: dayjs().subtract(1, 'month').endOf('month').format('YYYY-MM-DD'),
      currency: 'CNY',
      exchangeRate: String(defaultRate || 1),
    },
  })
  const currency = form.watch('currency')
  const usersQuery = useQuery({
    queryKey: ['billing-export-users', keyword],
    queryFn: () =>
      searchUsers({ keyword, p: 1, page_size: 50 }),
    enabled: props.open,
  })

  useEffect(() => {
    if (currency === 'USD') {
      form.setValue('exchangeRate', '1', { shouldValidate: true })
    }
  }, [currency, form])

  const submit = form.handleSubmit(async (values) => {
    try {
      await downloadBillingExport({
        userId: values.userId,
        startDate: values.startDate,
        endDate: values.endDate,
        currency: values.currency,
        exchangeRate: values.exchangeRate,
      })
      toast.success(t('Billing export downloaded'))
      props.onOpenChange(false)
    } catch {
      toast.error(t('Failed to export billing data'))
    }
  })
  const users = usersQuery.data?.data?.items || []

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>{t('Export bill')}</DialogTitle>
          <DialogDescription>
            {t('Export settled usage grouped by API token and model.')}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={submit}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor='billing-customer-search'>
                {t('Customer')}
              </FieldLabel>
              <Input
                id='billing-customer-search'
                value={keyword}
                onChange={(event) => setKeyword(event.target.value)}
                placeholder={t('Search customers')}
              />
              <Controller
                control={form.control}
                name='userId'
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <Select
                      value={field.value > 0 ? String(field.value) : null}
                      onValueChange={(value) => field.onChange(Number(value))}
                    >
                      <SelectTrigger className='w-full' aria-invalid={fieldState.invalid}>
                        <SelectValue placeholder={t('Select customer')} />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          {users.map((user) => (
                            <SelectItem key={user.id} value={String(user.id)}>
                              {user.username} (ID: {user.id})
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                    {fieldState.error && <FieldError>{t('Customer is required')}</FieldError>}
                  </Field>
                )}
              />
            </Field>
            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              <Field data-invalid={Boolean(form.formState.errors.startDate)}>
                <FieldLabel htmlFor='billing-start-date'>{t('Start date')}</FieldLabel>
                <Input id='billing-start-date' type='date' {...form.register('startDate')} />
              </Field>
              <Field data-invalid={Boolean(form.formState.errors.endDate)}>
                <FieldLabel htmlFor='billing-end-date'>{t('End date')}</FieldLabel>
                <Input id='billing-end-date' type='date' {...form.register('endDate')} />
                {form.formState.errors.endDate && (
                  <FieldError>{t('End date must not be before start date')}</FieldError>
                )}
              </Field>
            </div>
            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              <Controller
                control={form.control}
                name='currency'
                render={({ field }) => (
                  <Field>
                    <FieldLabel>{t('Settlement currency')}</FieldLabel>
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger className='w-full'><SelectValue /></SelectTrigger>
                      <SelectContent><SelectGroup>
                        <SelectItem value='CNY'>CNY</SelectItem>
                        <SelectItem value='USD'>USD</SelectItem>
                      </SelectGroup></SelectContent>
                    </Select>
                  </Field>
                )}
              />
              <Field data-disabled={currency === 'USD'}>
                <FieldLabel htmlFor='billing-exchange-rate'>{t('Exchange rate')}</FieldLabel>
                <Input
                  id='billing-exchange-rate'
                  inputMode='decimal'
                  disabled={currency === 'USD'}
                  aria-invalid={Boolean(form.formState.errors.exchangeRate)}
                  {...form.register('exchangeRate')}
                />
              </Field>
            </div>
            <FieldDescription>
              {t('Dates use the server timezone. The selected end date is included. Billing data comes from retained consumption logs.')}
            </FieldDescription>
          </FieldGroup>
          <DialogFooter className='mt-5'>
            <Button type='button' variant='outline' onClick={() => props.onOpenChange(false)}>
              {t('Cancel')}
            </Button>
            <Button type='submit' disabled={form.formState.isSubmitting}>
              {form.formState.isSubmitting && <Spinner data-icon='inline-start' />}
              {form.formState.isSubmitting ? t('Exporting...') : t('Export CSV')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
