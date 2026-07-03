package events

import "time"

type ArticleCreated struct {
	baseEvent
	articleID  string
	sku        string
	name       string
	priceCents int64
	currency   string
}

func NewArticleCreated(articleID, sku, name string, priceCents int64, currency string, at time.Time) ArticleCreated {
	return ArticleCreated{
		baseEvent:  baseEvent{occurredAt: occurredAt(at)},
		articleID:  articleID,
		sku:        sku,
		name:       name,
		priceCents: priceCents,
		currency:   currency,
	}
}

func (ArticleCreated) EventName() string   { return "ArticleCreated" }
func (e ArticleCreated) ArticleID() string { return e.articleID }
func (e ArticleCreated) SKU() string       { return e.sku }
func (e ArticleCreated) Name() string      { return e.name }
func (e ArticleCreated) PriceCents() int64 { return e.priceCents }
func (e ArticleCreated) Currency() string  { return e.currency }

type ArticlePriceChanged struct {
	baseEvent
	articleID     string
	oldPriceCents int64
	newPriceCents int64
	currency      string
}

func NewArticlePriceChanged(articleID string, oldPriceCents, newPriceCents int64, currency string, at time.Time) ArticlePriceChanged {
	return ArticlePriceChanged{
		baseEvent:     baseEvent{occurredAt: occurredAt(at)},
		articleID:     articleID,
		oldPriceCents: oldPriceCents,
		newPriceCents: newPriceCents,
		currency:      currency,
	}
}

func (ArticlePriceChanged) EventName() string      { return "ArticlePriceChanged" }
func (e ArticlePriceChanged) ArticleID() string    { return e.articleID }
func (e ArticlePriceChanged) OldPriceCents() int64 { return e.oldPriceCents }
func (e ArticlePriceChanged) NewPriceCents() int64 { return e.newPriceCents }
func (e ArticlePriceChanged) Currency() string     { return e.currency }
