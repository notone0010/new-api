# Admin Billing Export Design

## Goal

Add an administrator-only CSV export that produces a customer settlement statement from finalized consumption logs. The export aggregates usage by customer, API token, and model so it remains readable for high-volume customers while preserving the token-level cost-center boundary.

## Scope

The first release provides on-demand CSV generation only. It does not create persistent statements, invoice records, payment states, manual adjustments, tax calculations, or immutable billing snapshots.

Historical exports depend on retained consumption logs. Deployments must retain logs for at least 400 days. When ClickHouse is used, `LOG_SQL_CLICKHOUSE_TTL_DAYS` must therefore be unset, zero, or at least 400. Manual log cleanup remains destructive and must warn that deleted periods cannot be regenerated.

## Reference Statement

The reference CSV `/Users/lvsongke/Downloads/9月9日 xx账单.csv` contains 2,673 detail rows. It separates input, output, and cache-hit usage into billing units and exposes usage, unit price, currency, and original charge. The new export adopts its auditable column-oriented presentation but deliberately aggregates each customer/API-token/model combination into one row. This avoids an impractical request-per-row export and avoids inventing component prices when this project's dynamic, tiered, per-call, image, audio, or task billing cannot be decomposed reliably.

## User Experience

An administrator sees an **Export bill** action on the common usage-log page. The action is unavailable in the self-service user view.

The action opens a dialog containing:

- Customer: required single selection, stored and submitted by user ID.
- Billing period: required start and end dates.
- Currency: `CNY` or `USD`.
- Exchange rate: fixed to `1` and read-only for USD; initialized from the system USD-to-CNY exchange rate and editable for CNY.

The dialog explains that dates use the server timezone and that the selected end date is included. The browser sends calendar dates unchanged; the server converts them to a `[start, day-after-end)` Unix interval in its local timezone. The submit button shows a loading state while the browser downloads the response.

Successful responses download a UTF-8 CSV with BOM using a sanitized filename such as `billing-customer-2026-09-01-2026-10-01-CNY.csv`. An empty result produces a visible error and no file. API errors use the project's standard error handling.

## Authorization and API

Add an admin-authenticated endpoint under the existing log routes:

```http
GET /api/log/billing/export
```

Query parameters:

- `user_id`: positive integer, required.
- `start_date`: `YYYY-MM-DD` calendar date, inclusive in the server timezone.
- `end_date`: `YYYY-MM-DD` calendar date, inclusive and not before the start date.
- `currency`: `USD` or `CNY`.
- `exchange_rate`: decimal; must equal `1` for USD and must be finite and greater than zero for CNY.

Validation failures return HTTP 400. An unknown user returns HTTP 404. A period with no billable consumption returns HTTP 404 with a user-facing message. Database or streaming failures return HTTP 500.

The response uses `text/csv; charset=utf-8` and an RFC-compatible `Content-Disposition` attachment filename. CSV fields are written with Go's standard CSV writer; spreadsheet formula injection is prevented by prefixing text cells beginning with `=`, `+`, `-`, or `@` with an apostrophe.

Every successful export records an administrator operation-audit log containing the target user ID, period, currency, exchange rate, exported row count, and total quota. It never records the CSV body.

## Authoritative Data and Aggregation

The export reads `LOG_DB.logs`, not `quota_data`. Only finalized consumption rows with `type = LogTypeConsume` are included. The exact predicates are:

```text
user_id = selected customer
type = LogTypeConsume
created_at >= start_timestamp
created_at < end_timestamp
```

Rows are grouped by these stable billing dimensions:

```text
user_id, username, token_id, token_name, model_name
```

For each group the query computes:

- request count: `count(*)`
- prompt tokens: `sum(prompt_tokens)`
- completion tokens: `sum(completion_tokens)`
- settled quota: `sum(quota)`

Cache-hit tokens are stored inside the log's `other` JSON rather than a dedicated cross-database column. Because extracting JSON in SQL would require incompatible SQLite, MySQL, PostgreSQL, and ClickHouse expressions, the first release does not present cache tokens as a separately summed numeric column. The CSV labels prompt and completion tokens according to their stored log semantics and treats settled quota as the financial truth. This avoids a report whose cache total is correct only for some providers.

The query must work against SQLite, MySQL, PostgreSQL, and ClickHouse log databases. It uses GORM aggregation and the shared quoted `group` column constant where relevant. Results are ordered by token ID, model name, and username to ensure deterministic output.

Deleted tokens are not joined as a prerequisite. Historical `token_id` and `token_name` values already captured on each log remain the source. If the stored name is empty, the display name is `Deleted token (<id>)`.

## Financial Calculations

`quota` is the actual post-settlement customer charge and is never recomputed from current model prices, ratios, or billing expressions.

The export captures the current positive `QuotaPerUnit` once at request start:

```text
usd_amount = settled_quota / quota_per_unit
settlement_amount = usd_amount * exchange_rate
```

Calculations use decimal arithmetic. Raw quota remains an integer. `QuotaPerUnit`, exchange rate, USD amount, and settlement amount are emitted as decimal strings without binary floating-point artifacts. Display amounts are rounded to six decimal places while the raw quota and conversion inputs remain available for reconciliation.

The export does not split a settled charge into input/output/cache monetary components. This is essential for correctness because the project supports fixed per-call pricing, group multipliers, conditional request multipliers, tool surcharges, image/audio/task charges, and expression-based tiered pricing. The sum of the CSV row quotas must exactly equal the sum of matching consumption-log quotas.

## CSV Columns

The output contains one header row followed by one row per aggregated token/model group:

1. Billing period
2. Customer ID
3. Customer name
4. API Token ID
5. API Token name
6. Model
7. Billing method (`Actual settled usage`)
8. Request count
9. Input tokens
10. Output tokens
11. Total tokens
12. Settled quota
13. Quota per USD
14. Original amount (USD)
15. Settlement currency
16. Exchange rate
17. Settlement amount

The first release uses stable English headers so exported files remain machine-readable across administrator UI languages. Frontend dialog text and messages are localized through the existing i18n system.

## Components and Responsibilities

### Backend model

A focused billing-export model unit owns the filter and result types and the cross-database aggregation query. It returns numeric facts only and does not know about HTTP, CSV formatting, currency labels, or audit events.

### Backend controller

A focused controller unit validates the request, resolves the customer, captures `QuotaPerUnit`, performs decimal conversion, safely writes CSV, and records the successful audit event. Keeping it separate from the already broad log controller prevents unrelated log-list behavior from becoming coupled to file export.

### Router

The endpoint is registered before the parameterized or catch-all log routes and uses `middleware.AdminAuth()`.

### Frontend

The usage-log header action owns opening the dialog. A dedicated billing-export dialog owns form state and validation. The API module requests the response as a blob, obtains the server filename when available, and triggers a browser download without navigating away.

## Error Handling and Safety

- Reject invalid IDs, dates, currencies, exchange rates, reversed ranges, and non-positive `QuotaPerUnit` before querying or writing a response.
- Set response headers only after validation and successful aggregation, so JSON errors remain readable by the frontend.
- Treat all database-derived text as untrusted CSV content and neutralize spreadsheet formulas.
- Do not export request IDs, IP addresses, channel credentials, `other`, admin-only diagnostics, or upstream identifiers.
- Abort cleanly if the request context is canceled.
- Avoid a request-count row limit because aggregation bounds output cardinality to token/model combinations rather than raw logs.
- Warn administrators in the existing log-cleanup flow that cleaned periods cannot be exported again.

## Verification

Backend tests protect these observable contracts:

- grouping distinguishes API tokens and models while combining requests in the same group;
- only the selected user, consumption type, and half-open period are included;
- the summed exported quota exactly reconciles to source consumption logs;
- empty token names use the deleted-token label;
- USD and CNY conversions use captured inputs and deterministic decimal rounding;
- invalid input is rejected before response headers are committed;
- CSV has a BOM, correct headers, safe quoting, and spreadsheet-formula neutralization;
- the route requires administrator authentication;
- aggregation works with the project's SQLite test fixture, with query construction remaining portable to MySQL, PostgreSQL, and ClickHouse.

Frontend tests protect these user-visible contracts:

- the action appears only in administrator scope;
- customer, period, currency, and exchange rate validation is enforced;
- USD forces exchange rate `1`, while CNY permits a positive override;
- submitting downloads the returned blob with the response filename;
- empty-data and API failures show an error without creating a download.

Run affected Go tests, the frontend Vitest files, frontend type checking, lint for changed files, and a production frontend build.

## Future Extension

A later release may create immutable billing statements with generation, confirmation, adjustment, invoicing, and payment states. That snapshot must preserve its source period, `QuotaPerUnit`, currency, exchange rate, totals, and item rows so it remains downloadable after raw logs expire. The API and aggregation result introduced here should be reusable by that statement-generation service.
