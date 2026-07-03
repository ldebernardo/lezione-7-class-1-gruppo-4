package entities_test

import (
	"errors"
	"testing"
	"time"

	"warehouse.local/core/entities"
	"warehouse.local/core/events"
)

func validArticleParts(t *testing.T) (entities.SKU, entities.Money, entities.WarehouseLocation) {
	t.Helper()
	sku, err := entities.NewSKU("HW-001")
	if err != nil {
		t.Fatalf("sku: %v", err)
	}
	price, err := entities.NewMoney(1299, "EUR")
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	location, err := entities.NewWarehouseLocation("IT-MIL1")
	if err != nil {
		t.Fatalf("location: %v", err)
	}
	return sku, price, location
}

func newValidArticle(t *testing.T) (*entities.Article, entities.WarehouseLocation) {
	t.Helper()
	sku, price, location := validArticleParts(t)
	article, err := entities.NewArticle("article-1", sku, "NAS Lite", "Network storage", price, time.Unix(100, 0))
	if err != nil {
		t.Fatalf("article: %v", err)
	}
	return article, location
}

func TestArticleFactoryGuardsNameAndPrice(t *testing.T) {
	t.Parallel()
	sku, price, _ := validArticleParts(t)

	if _, err := entities.NewArticle("", sku, "NAS Lite", "", price, time.Time{}); !errors.Is(err, entities.ErrInvalidArticleID) {
		t.Fatalf("expected invalid id, got %v", err)
	}
	if _, err := entities.NewArticle("article-1", sku, " ", "", price, time.Time{}); !errors.Is(err, entities.ErrInvalidName) {
		t.Fatalf("expected invalid name, got %v", err)
	}
	zero, err := entities.NewMoney(0, "EUR")
	if err != nil {
		t.Fatalf("zero money should be valid: %v", err)
	}
	if _, err := entities.NewArticle("article-1", sku, "NAS Lite", "", zero, time.Time{}); !errors.Is(err, entities.ErrInvalidPrice) {
		t.Fatalf("expected invalid article price, got %v", err)
	}
}

func TestArticleChangePriceKeepsCurrencyStable(t *testing.T) {
	t.Parallel()
	article, _ := newValidArticle(t)
	article.PullEvents()

	usd, err := entities.NewMoney(1400, "USD")
	if err != nil {
		t.Fatalf("usd money: %v", err)
	}
	if err := article.ChangePrice(usd, time.Unix(101, 0)); !errors.Is(err, entities.ErrCurrencyChange) {
		t.Fatalf("expected currency change rejection, got %v", err)
	}

	eur, err := entities.NewMoney(1499, "EUR")
	if err != nil {
		t.Fatalf("eur money: %v", err)
	}
	if err := article.ChangePrice(eur, time.Unix(102, 0)); err != nil {
		t.Fatalf("change price: %v", err)
	}
	if got := article.Price().AmountCents(); got != 1499 {
		t.Fatalf("expected price 1499, got %d", got)
	}
	events := article.PendingEvents()
	if len(events) != 1 || events[0].EventName() != "ArticlePriceChanged" {
		t.Fatalf("expected ArticlePriceChanged event, got %#v", events)
	}
}

func TestArticleRecordsEventsAndDrainsWithoutPublishing(t *testing.T) {
	t.Parallel()
	article, location := newValidArticle(t)

	pending := article.PendingEvents()
	if len(pending) != 1 || pending[0].EventName() != "ArticleCreated" {
		t.Fatalf("expected ArticleCreated pending event, got %#v", pending)
	}
	if _, ok := pending[0].(events.ArticleCreated); !ok {
		t.Fatalf("expected concrete ArticleCreated event, got %T", pending[0])
	}
	drained := article.PullEvents()
	if len(drained) != 1 {
		t.Fatalf("expected one drained event, got %d", len(drained))
	}
	if remaining := article.PendingEvents(); len(remaining) != 0 {
		t.Fatalf("expected empty pending events after drain, got %d", len(remaining))
	}

	if err := article.AddInventoryLevel("level-1", location, 10, 0); err != nil {
		t.Fatalf("add inventory: %v", err)
	}
	if err := article.AdjustInventory(location, 5, "manual-count", time.Unix(103, 0)); err != nil {
		t.Fatalf("adjust inventory: %v", err)
	}
	if got := article.PendingEvents(); len(got) != 1 || got[0].EventName() != "InventoryAdjusted" {
		t.Fatalf("expected InventoryAdjusted pending event, got %#v", got)
	}
}

func TestInventoryStaysSaneAndOnlyArticleMutatesIt(t *testing.T) {
	t.Parallel()
	article, location := newValidArticle(t)
	if err := article.AddInventoryLevel("level-1", location, 10, 0); err != nil {
		t.Fatalf("add inventory: %v", err)
	}

	if err := article.AdjustInventory(location, -11, "bad-count", time.Time{}); !errors.Is(err, entities.ErrInvalidQuantity) {
		t.Fatalf("expected negative quantity rejection, got %v", err)
	}
	if err := article.ReserveStock(location, 11, "reservation-1", "order-1", time.Time{}); !errors.Is(err, entities.ErrReservedExceedsStock) {
		t.Fatalf("expected over-reservation rejection, got %v", err)
	}
	if err := article.ReserveStock(location, 4, "reservation-1", "order-1", time.Time{}); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if err := article.AdjustInventory(location, -7, "bad-count", time.Time{}); !errors.Is(err, entities.ErrReservedExceedsStock) {
		t.Fatalf("expected reserved exceeds stock rejection, got %v", err)
	}
	if err := article.ReleaseStock(location, 5, "reservation-1", "cancelled", time.Time{}); !errors.Is(err, entities.ErrInvalidReserved) {
		t.Fatalf("expected over-release rejection, got %v", err)
	}

	level, ok := article.InventoryLevel(location)
	if !ok {
		t.Fatal("expected inventory level")
	}
	if level.Quantity() != 10 || level.Reserved() != 4 || level.Available() != 6 {
		t.Fatalf("unexpected level: quantity=%d reserved=%d available=%d", level.Quantity(), level.Reserved(), level.Available())
	}
}

func TestInventoryStressKeepsStockSane(t *testing.T) {
	article, location := newValidArticle(t)
	if err := article.AddInventoryLevel("level-1", location, 0, 0); err != nil {
		t.Fatalf("add inventory: %v", err)
	}
	article.PullEvents()

	const adjustments = 500
	done := make(chan error, adjustments)
	for i := 0; i < adjustments; i++ {
		go func() {
			done <- article.AdjustInventory(location, 1, "stress", time.Time{})
		}()
	}
	for i := 0; i < adjustments; i++ {
		if err := <-done; err != nil {
			t.Fatalf("stress adjustment failed: %v", err)
		}
	}
	level, ok := article.InventoryLevel(location)
	if !ok {
		t.Fatal("expected inventory level")
	}
	if level.Quantity() != adjustments {
		t.Fatalf("expected quantity %d, got %d", adjustments, level.Quantity())
	}
	if len(article.PendingEvents()) != adjustments {
		t.Fatalf("expected %d pending events, got %d", adjustments, len(article.PendingEvents()))
	}
}
