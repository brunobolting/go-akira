package entity

import "time"

type ContentReview struct {
	ID        string
	VolumeID  string
	Author    string
	Title     string
	Content   string
	Rating    float64
	Date      time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Book struct {
	ID          string
	Name        string
	Edition     string
	Description string
	Slug        string
	CoverImage  string
	PageCount   int
	Volume      *int
	Rating      float64
	Reviews     []ContentReview
	Publisher   string
	Author      []string
	UserID      string
	ISBN        string
	Tags        []string
	Metadata    map[string]string
	Language    string
	LastSync    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewBook(
	userID, name, edition, description, slug, coverImage string,
	pageCount int, volume *int, rating float64,
	publisher string, author []string, isbn string,
	tags []string, metadata map[string]string, language string,
) *Book {
	return &Book{
		UserID:      userID,
		Name:        name,
		Edition:     edition,
		Description: description,
		Slug:        slug,
		CoverImage:  coverImage,
		PageCount:   pageCount,
		Volume:      volume,
		Rating:      rating,
		Publisher:   publisher,
		Author:      author,
		ISBN:        isbn,
		Tags:        tags,
		Metadata:    metadata,
		Language:    language,
	}
}

type CreateBookRequest struct {
	Name        string
	Edition     string
	Description string
	Slug        string
	CoverImage  string
	PageCount   int
	Volume      *int
	Rating      float64
	Publisher   string
	Author      []string
	ISBN        string
	Tags        []string
	Metadata    map[string]string
	Language    string
	ColletionID *string
}

func (r *CreateBookRequest) Validate() error {
	var e RequestError
	if r.Name == "" {
		e = e.Add("name", ErrBookNameInvalid.Error())
	}
	if len(r.Name) > 255 {
		e = e.Add("name", ErrBookNameTooLong.Error())
	}
	return e
}

type BookService interface {
	CreateBook(userID string, req CreateBookRequest) (*Book, error)
	CreateCollectionBook(userID, collectionID string, req CreateBookRequest) (*Book, error)
}

type BookRepository interface {
	CreateBook(book *Book) error
	CreateCollectionBook(collectionID string, book *Book) error
	FindBookBySlug(userID, slug string) (*Book, error)
}
