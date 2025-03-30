package book

import "akira/internal/entity"

var _ entity.BookRepository = (*MemoRepository)(nil)

type MemoRepository struct {
	books map[string]*entity.Book
}

func NewMemoRepository() *MemoRepository {
	return &MemoRepository{
		books: make(map[string]*entity.Book),
	}
}

func (r *MemoRepository) CreateBook(book *entity.Book) error {
	r.books[book.ID] = book
	return nil
}

func (r *MemoRepository) CreateCollectionBook(collectionID string, book *entity.Book) error {
	// This method is not implemented in the memo repository.
	// In a real implementation, you would associate the book with the collection.
	return nil
}

func (r *MemoRepository) FindBookBySlug(userID, slug string) (*entity.Book, error) {
	if book, ok := r.books[slug]; ok {
		return book, nil
	}
	return nil, entity.ErrNotFound
}
