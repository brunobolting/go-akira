package collection

import "akira/internal/entity"

var _ entity.CollectionRepository = (*MemoRepository)(nil)

type MemoRepository struct {
	collections map[string]*entity.Collection
}

func NewMemoRepository() *MemoRepository {
	return &MemoRepository{
		collections: make(map[string]*entity.Collection),
	}
}

func (r *MemoRepository) CreateCollection(collection *entity.Collection) error {
	r.collections[collection.ID] = collection
	return nil
}

func (r *MemoRepository) FindCollectionBySlug(userID, slug string) (*entity.Collection, error) {
	if collection, ok := r.collections[slug]; ok {
		return collection, nil
	}
	return nil, entity.ErrNotFound
}
