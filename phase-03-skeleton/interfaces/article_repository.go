package interfaces

import "warehouse.local/core/entities"

type ArticleRepository interface {
	Save(article *entities.Article) error
	FindByID(id string) (*entities.Article, error)
	FindBySKU(sku entities.SKU) (*entities.Article, error)
	List() ([]*entities.Article, error)
	Delete(id string) error
}
