package merchant

import "testing"

func hailuoTestSKU(id, suk string, price float64, recycle string) map[string]any {
	sku := map[string]any{"id": id, "suk": suk, "price": price}
	if recycle != "" {
		sku["is_recycle"] = recycle
	}
	return sku
}

func TestHailuoPickSKU(t *testing.T) {
	tests := []struct {
		name         string
		productValue any
		preferred    string
		wantPrice    float64
		wantRecycled bool
		wantFound    bool
	}{
		{
			name:         "missing productValue falls back to storeInfo",
			productValue: nil,
			wantFound:    false,
		},
		{
			name:         "empty productValue falls back to storeInfo",
			productValue: map[string]any{},
			wantFound:    false,
		},
		{
			name: "picks the only recyclable bundle spec",
			productValue: map[string]any{
				"简体中文,【塞尔达本体】二手盒装": hailuoTestSKU("61373", "简体中文,【塞尔达本体】二手盒装", 188, "0"),
				"简体中文,【塞尔达同捆】二手盒装": hailuoTestSKU("61374", "简体中文,【塞尔达同捆】二手盒装", 504, "1"),
				"简体中文,【塞尔达同捆】盒装未拆": hailuoTestSKU("61375", "简体中文,【塞尔达同捆】盒装未拆", 525, "0"),
			},
			wantPrice:    504,
			wantRecycled: true,
			wantFound:    true,
		},
		{
			name: "picks recyclable used boxed spec over accessory placeholder",
			productValue: map[string]any{
				"简体中文,二手盒装": hailuoTestSKU("57825", "简体中文,二手盒装", 164, "1"),
				"简体中文,全新绑腿": hailuoTestSKU("57829", "简体中文,全新绑腿", 9, "0"),
			},
			wantPrice:    164,
			wantRecycled: true,
			wantFound:    true,
		},
		{
			name: "preferred key selects an exact spec",
			productValue: map[string]any{
				"简体中文,【紫（二手盒装）】": hailuoTestSKU("61715", "简体中文,【紫（二手盒装）】", 193, "0"),
				"简体中文,【紫同捆】二手带盒": hailuoTestSKU("61719", "简体中文,【紫同捆】二手带盒", 464, "1"),
			},
			preferred:    "简体中文,【紫同捆】二手带盒",
			wantPrice:    464,
			wantRecycled: true,
			wantFound:    true,
		},
		{
			name: "preferred sku id selects an exact spec",
			productValue: map[string]any{
				"简体中文,【紫（二手盒装）】": hailuoTestSKU("61715", "简体中文,【紫（二手盒装）】", 193, "0"),
				"简体中文,【紫同捆】二手带盒": hailuoTestSKU("61719", "简体中文,【紫同捆】二手带盒", 464, "1"),
			},
			preferred:    "61719",
			wantPrice:    464,
			wantRecycled: true,
			wantFound:    true,
		},
		{
			name: "unknown preferred falls back to recyclable spec",
			productValue: map[string]any{
				"简体中文,【朱同捆】二手带盒": hailuoTestSKU("61877", "简体中文,【朱同捆】二手带盒", 449, "1"),
				"简体中文,【朱（二手盒装）】": hailuoTestSKU("61873", "简体中文,【朱（二手盒装）】", 179, "0"),
			},
			preferred:    "not-exist",
			wantPrice:    449,
			wantRecycled: true,
			wantFound:    true,
		},
		{
			name: "prefers used boxed among multiple recyclable specs",
			productValue: map[string]any{
				"简体中文,全新":   hailuoTestSKU("1", "简体中文,全新", 199, "1"),
				"简体中文,二手盒装": hailuoTestSKU("2", "简体中文,二手盒装", 164, "1"),
			},
			wantPrice:    164,
			wantRecycled: true,
			wantFound:    true,
		},
		{
			name: "no recyclable spec is unavailable",
			productValue: map[string]any{
				"简体中文,二手盒装": hailuoTestSKU("1", "简体中文,二手盒装", 164, "0"),
			},
			wantRecycled: false,
			wantFound:    true,
		},
		{
			name: "missing recycle flags falls back to storeInfo",
			productValue: map[string]any{
				"简体中文,二手盒装": hailuoTestSKU("1", "简体中文,二手盒装", 164, ""),
			},
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := hailuoPickSKU(tt.productValue, tt.preferred)
			if found != tt.wantFound || got.Price != tt.wantPrice || got.Recyclable != tt.wantRecycled {
				t.Fatalf("hailuoPickSKU() = (%+v, %v), want price=%v recyclable=%v found=%v",
					got, found, tt.wantPrice, tt.wantRecycled, tt.wantFound)
			}
		})
	}
}

func TestHailuoRecyclePrice(t *testing.T) {
	tests := []struct {
		name   string
		sell   float64
		diff   float64
		want   float64
		wantOK bool
	}{
		{name: "normal deduction", sell: 165, diff: 10, want: 155, wantOK: true},
		{name: "zero deduction", sell: 165, diff: 0, want: 165, wantOK: true},
		{name: "negative result is unavailable", sell: 9, diff: 30, want: -21, wantOK: false},
		{name: "zero result is unavailable", sell: 30, diff: 30, want: 0, wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := hailuoRecyclePrice(tt.sell, tt.diff)
			if got != tt.want || ok != tt.wantOK {
				t.Fatalf("hailuoRecyclePrice(%v, %v) = (%v, %v), want (%v, %v)", tt.sell, tt.diff, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
