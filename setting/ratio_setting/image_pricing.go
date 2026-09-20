package ratio_setting

import (
	"fmt"
	"math"

	"github.com/QuantumNous/new-api/common"
)

// Pixel tiers use exclusive lower bounds and inclusive upper bounds. The last
// tier has no upper bound. Prices are USD per image; ratios may exceed one.
type ImagePixelTier struct {
	MaxPixels  *int64  `json:"max_pixels"`
	Multiplier float64 `json:"multiplier"`
}

type ImagePricingConfig struct {
	BasePrice  *float64         `json:"base_price"`
	PixelTiers []ImagePixelTier `json:"pixel_tiers,omitempty"`
}

const MaxImageBillingPixels int64 = 9007199254740991

func (cfg ImagePricingConfig) Validate() error {
	if cfg.BasePrice == nil || math.IsNaN(*cfg.BasePrice) || math.IsInf(*cfg.BasePrice, 0) || *cfg.BasePrice < 0 {
		return fmt.Errorf("base_price must be a non-negative finite number")
	}
	var previous int64
	for i, tier := range cfg.PixelTiers {
		if tier.Multiplier <= 0 || math.IsNaN(tier.Multiplier) || math.IsInf(tier.Multiplier, 0) || math.IsInf(*cfg.BasePrice*tier.Multiplier, 0) {
			return fmt.Errorf("pixel tier %d must have a positive finite multiplier and price", i+1)
		}
		if i == len(cfg.PixelTiers)-1 {
			if tier.MaxPixels != nil {
				return fmt.Errorf("last pixel tier must have an unlimited upper bound")
			}
			continue
		}
		if tier.MaxPixels == nil || *tier.MaxPixels <= previous || *tier.MaxPixels > MaxImageBillingPixels {
			return fmt.Errorf("pixel tier upper bounds must be positive, strictly increasing safe integers")
		}
		previous = *tier.MaxPixels
	}
	return nil
}

func GetImagePricingConfig(model string) (ImagePricingConfig, bool) {
	raw, ok := imagePriceMap.Get(FormatMatchingModelName(model))
	if !ok || common.GetJsonType(raw) != "object" {
		return ImagePricingConfig{}, false
	}
	var cfg ImagePricingConfig
	if common.Unmarshal(raw, &cfg) != nil {
		return cfg, false
	}
	return cfg, true
}
