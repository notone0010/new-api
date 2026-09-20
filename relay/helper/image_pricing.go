package helper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	hosttypes "github.com/QuantumNous/new-api/types"
)

// ResolveImagePricing reads only request parameters, never response metadata.
func ResolveImagePricing(req *dto.ImageRequest, cfg ratio_setting.ImagePricingConfig) (*hosttypes.ImageBillingDetail, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	detail := &hosttypes.ImageBillingDetail{Version: 1, BasePrice: *cfg.BasePrice, Multiplier: 1, UnitPrice: *cfg.BasePrice, Size: req.Size, Count: 1, CountSource: "request", QuotaPerUnit: common.QuotaPerUnit}
	if req.N != nil {
		if *req.N == 0 || *req.N > dto.MaxImageN {
			return nil, fmt.Errorf("n must be between 1 and %d", dto.MaxImageN)
		}
		detail.Count = int(*req.N)
	}
	if len(cfg.PixelTiers) == 0 {
		return detail, nil
	}
	parts := strings.Split(req.Size, "x")
	if req.Size == "" && len(req.Extra["width"]) > 0 && len(req.Extra["height"]) > 0 {
		var width, height int64
		if common.Unmarshal(req.Extra["width"], &width) != nil || common.Unmarshal(req.Extra["height"], &height) != nil {
			return nil, fmt.Errorf("pixel billing requires integer request width and height")
		}
		parts = []string{strconv.FormatInt(width, 10), strconv.FormatInt(height, 10)}
		detail.Size = strings.Join(parts, "x")
	}
	if len(parts) != 2 {
		return nil, fmt.Errorf("pixel billing requires explicit request size WIDTHxHEIGHT or integer width and height; auto, resolution labels and missing dimensions are unsupported")
	}
	w, ew := strconv.ParseInt(parts[0], 10, 64)
	h, eh := strconv.ParseInt(parts[1], 10, 64)
	if ew != nil || eh != nil || w <= 0 || h <= 0 || w > ratio_setting.MaxImageBillingPixels/h {
		return nil, fmt.Errorf("pixel billing requires positive integer dimensions with width*height <= %d", ratio_setting.MaxImageBillingPixels)
	}
	detail.Width, detail.Height, detail.Pixels = w, h, w*h
	for i, tier := range cfg.PixelTiers {
		if tier.MaxPixels == nil || detail.Pixels <= *tier.MaxPixels {
			detail.Tier, detail.MaxPixels, detail.Multiplier = i+1, tier.MaxPixels, tier.Multiplier
			detail.UnitPrice = detail.BasePrice * tier.Multiplier
			return detail, nil
		}
		detail.MinPixels = *tier.MaxPixels
	}
	return nil, fmt.Errorf("no pixel billing tier matched")
}
