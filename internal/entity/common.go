package entity

type Rowscan interface {
	// Scan *sql.Row|Rows.Scan
	Scan(dest ...any) error
}
