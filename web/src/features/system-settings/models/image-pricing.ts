export type ImagePriceConfig = {
  base_price: number | ''
  pixel_tiers?: Array<{
    max_pixels: number | null | ''
    multiplier: number | ''
  }>
}

export function validateImagePricing(value: string): boolean {
  try {
    const cfg = JSON.parse(value) as ImagePriceConfig
    if (
      !cfg ||
      Array.isArray(cfg) ||
      typeof cfg.base_price !== 'number' ||
      !Number.isFinite(cfg.base_price) ||
      cfg.base_price < 0
    ) {
      return false
    }
    if (cfg.pixel_tiers === undefined) return true
    if (!Array.isArray(cfg.pixel_tiers)) return false
    let previous = 0
    return cfg.pixel_tiers.every((tier, index, tiers) => {
      if (
        !tier ||
        typeof tier.multiplier !== 'number' ||
        !Number.isFinite(tier.multiplier) ||
        tier.multiplier <= 0 ||
        !Number.isFinite(Number(cfg.base_price) * tier.multiplier)
      ) {
        return false
      }
      if (index === tiers.length - 1) return tier.max_pixels === null
      if (
        typeof tier.max_pixels !== 'number' ||
        !Number.isSafeInteger(tier.max_pixels) ||
        tier.max_pixels <= previous
      ) {
        return false
      }
      previous = tier.max_pixels
      return true
    })
  } catch {
    return false
  }
}
