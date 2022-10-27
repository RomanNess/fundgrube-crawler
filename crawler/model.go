package crawler

import (
	"fmt"
	"time"
)

type query struct {
	Keyword string
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
	PostingId   string     `json:"posting_id" bson:"_id"`
	PriceString string     `json:"price" bson:"-"`
	Price       float64    `json:"-" bson:"price"`
	Name        string     `json:"name"`
	Url         []string   `json:"original_url"`
	Text        string     `json:"posting_text"`
	Outlet      outlet     `json:"outlet"`
	Brand       brand      `json:"brand"`
	Shop        Shop       `bson:"shop"`
	ShopUrl     string     `bson:"shop_url"`
	CreDat      *time.Time `bson:"cre_dat"`
	ModDat      *time.Time `bson:"mod_dat"`
}

func (p posting) String() string {
	return fmt.Sprintf("%.2f€ - %s in %s\n\t🌄 %s\n\t🛒 %s", p.Price, p.Name, p.Outlet.Name, p.Url[0], p.ShopUrl)
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
}

type brand struct {
	BrandId int    `json:"id"`
	Name    string `json:"name"`
}

type operation struct {
	Id        string     `bson:"_id"`
	Timestamp *time.Time `bson:"timestamp"`
}
