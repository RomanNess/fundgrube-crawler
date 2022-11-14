package crawler

import (
	"bytes"
	"fmt"
	"regexp"
	"time"
)

type queries struct {
	Queries []query `yaml:"queries"`
}

type query struct {
	Desc         string   `yaml:"desc" json:"desc,omitempty" bson:"desc"`
	NameRegex    *string  `yaml:"name_regex" json:"name_regex,omitempty" bson:"name_regex"`
	BrandRegex   *string  `yaml:"brand_regex" json:"brand_regex,omitempty" bson:"brand_regex"`
	PriceMin     *float64 `yaml:"price_min" json:"price_min,omitempty" bson:"price_min"`
	PriceMax     *float64 `yaml:"price_max" json:"price_max,omitempty" bson:"price_max"`
	DiscountMin  *int     `yaml:"discount_min" json:"discount_min,omitempty" bson:"discount_min"`
	OutletId     *int     `yaml:"outlet_id" json:"outlet_id,omitempty" bson:"outlet_id"`
	Ids          []string `yaml:"-" json:"-,omitempty" bson:"-"`
	FindInactive bool     `yaml:"find_inactive" json:"find_inactive,omitempty" bson:"find_inactive"`
}

func (q query) String() string {
	var buffer bytes.Buffer
	if q.NameRegex != nil {
		buffer.WriteString("regex: " + *q.NameRegex)
	}
	return buffer.String()
}

type postingsResponse struct {
	Postings     []posting `json:"postings"`
	Outlets      []outlet  `json:"outlets"`
	HasMorePages bool      `json:"morePostingsAvailable"`
}

/* Example json
   {
     "posting_id": "e6194b60-f031-4e25-b2c7-e8067ad9dac1",
     "posting_text": "Neuware",
     "price": "139.00",
     "price_old": "319.00",
     "shipping_cost": 0,
     "shipping_type": "shipping",
     "discount_in_percent": 56,
     "name": "AKRACING Core EXSE Schwarz/Carbon Gaming Stuhl, Carbon",
     "brand": {
       "id": 6927,
       "name": "AKRACING"
     },
     "eek": {},
     "top_level_catalog_id": "CAT_DE_SAT_786",
     "original_url": [
       "https://assets.mmsrg.com/is/166325/12975367df8e182e57044734f5165e190/c3/-/31b5554e0e7f4ad5a6a07101fd3750aa"
     ],
     "outlet": {
       "id": 60,
       "name": "Braunschweig"
     },
     "pim_id": 2681077
   }
*/
type posting struct {
	PostingId         string        `json:"posting_id" bson:"_id"`
	PriceString       string        `json:"price" bson:"-"`
	PriceOldString    string        `json:"price_old" bson:"-"`
	Price             float64       `json:"-" bson:"price"`
	PriceOld          float64       `json:"-" bson:"price_old"`
	DiscountInPercent int           `json:"discount_in_percent" bson:"discount_in_percent"`
	ShippingCost      float64       `json:"shipping_cost" bson:"shipping_cost"`
	ShippingType      string        `json:"shipping_type" bson:"shipping_type"`
	Name              string        `json:"name" bson:"name"`
	Url               []string      `json:"original_url" bson:"url"`
	Text              string        `json:"posting_text" bson:"text"`
	Outlet            postingOutlet `json:"outlet" bson:"outlet"`
	Brand             brand         `json:"brand" bson:"brand"`
	Shop              Shop          `json:"-" bson:"shop"`
	ShopUrl           string        `json:"-" bson:"shop_url"`
	PimId             int           `json:"pim_id" bson:"pim_id"`
	CreDat            *time.Time    `json:"-" bson:"cre_dat" `
	ModDat            *time.Time    `json:"-" bson:"mod_dat"`
	Active            bool          `json:"-" bson:"active"`
}

func (p posting) String() string {
	shippingInfo := ""
	if p.ShippingType == "shipping" {
		shippingInfo = fmt.Sprintf(" +%.2f€", p.ShippingCost)
	}
	uvpInfo := ""
	if p.PriceOld != 0 {
		uvpInfo = fmt.Sprintf(" (UVP %.2f€ -%d%%)", p.PriceOld, p.DiscountInPercent)
	}
	priceInfo := fmt.Sprintf("%.2f€%s%s", p.Price, shippingInfo, uvpInfo)
	return fmt.Sprintf("%s 👉%s👈 in %s [%s]\n\t📗 %s\n\t📸 %s\n\t🛒 %s", priceInfo, p.Name, p.Outlet.Name, p.PostingId, shorten(p.Text), p.Url[0], p.ShopUrl)
}

func shorten(text string) string {
	re := regexp.MustCompile("\\r?\\n")
	text = re.ReplaceAllString(text, " | ")
	if len(text) > 150 {
		text = text[:150] + "..."
	}
	return text
}

/*
   {
     "id": 67,
     "nameFull": "Saturn Neu-Isenburg",
     "name": "Neu-Isenburg",
     "isActive": true,
     "count": 76
   }
*/
type outlet struct {
	OutletId int    `json:"id"`
	Name     string `json:"name"`
	Count    int    `json:"count"`
}

type postingOutlet struct {
	OutletId int    `json:"id" bson:"id"`
	Name     string `json:"name" bson:"name"`
}

type brand struct {
	BrandId int    `json:"id" bson:"id"`
	Name    string `json:"name" bson:"name"`
}

type operation struct {
	Id          string     `bson:"_id"`
	Description string     `bson:"description"`
	Query       query      `bson:"query"`
	Timestamp   *time.Time `bson:"timestamp"`
}
