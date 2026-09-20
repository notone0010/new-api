package ratio_setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImagePriceMatchesMostSpecificRule(t *testing.T) {
	require.NoError(t, UpdateImagePriceByJSONString(`{
		"test-image": [
			{"price": 1, "size": "1024x1024"},
			{"price": 2, "size": "1024x1024", "resolution": "2k"},
			{"price": 3}
		]
	}`))
	t.Cleanup(func() { require.NoError(t, UpdateImagePriceByJSONString(`{}`)) })

	price, ok := GetImagePrice("test-image", "1024x1024", "2k", "")
	require.True(t, ok)
	require.Equal(t, 2.0, price)

	price, ok = GetImagePrice("test-image", "1024x1024", "", "")
	require.True(t, ok)
	require.Equal(t, 1.0, price)
}

func TestImagePriceRejectsInvalidRule(t *testing.T) {
	err := UpdateImagePriceByJSONString(`{"test-image":[{"price":-1}]}`)
	require.Error(t, err)
}
