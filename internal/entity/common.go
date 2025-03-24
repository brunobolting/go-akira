package entity

import "encoding/json"

type RequestError map[string][]string

func (e RequestError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (e RequestError) Add(field, message string) RequestError {
	if e == nil {
		e = RequestError{}
	}
	if _, ok := e[field]; !ok {
		e[field] = []string{}
	}
	e[field] = append(e[field], message)
	return e
}

func (e RequestError) HasError() bool {
	return len(e) > 0
}

type Rowscan interface {
	// Scan *sql.Row|Rows.Scan
	Scan(dest ...any) error
}
