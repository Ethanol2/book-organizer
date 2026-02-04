package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type CategoryType string

const (
	NoType    CategoryType = ""
	Series    CategoryType = "series"
	Authors   CategoryType = "authors"
	Narrators CategoryType = "narrators"
	Genres    CategoryType = "genres"
)

type Category struct {
	Id    *int         `json:"id,omitempty"`
	Name  string       `json:"name"`
	Index *string      `json:"index,omitempty"`
	Type  CategoryType `json:"-"`
}

var categorySingular = map[CategoryType]string{
	Series:    "series",
	Genres:    "genre",
	Authors:   "author",
	Narrators: "narrator",
}

func (c Client) AddCategory(tx *sql.Tx, categoryType CategoryType, name string) (Category, error) {

	indyTx := tx == nil
	if indyTx {
		tx, _ = c.db.Begin()
		defer tx.Rollback()
	}

	query := fmt.Sprintf(`
	INSERT INTO %s
		(id, name)
	VALUES
		(NULL, ?)
	`, categoryType)

	result, err := tx.Exec(query, name)
	if err != nil {
		return Category{}, err
	}

	log.Println("Added \"", name, "\" to", categoryType)

	id64, err := result.LastInsertId()
	id := int(id64)
	if err != nil {
		return Category{}, err
	}

	if indyTx {
		err = tx.Commit()
		if err != nil {
			return Category{}, err
		}
	}

	return Category{
		Id:   &id,
		Name: name,
		Type: categoryType,
	}, nil
}

func (c Client) DeleteCategory(category Category) error {
	return c.DeleteCategoryWithID(category.Type, *category.Id)
}

func (c Client) DeleteCategoryWithID(categoryType CategoryType, id int) error {

	query := fmt.Sprintf(`
	SELECT name FROM %s WHERE id = ?;
	REMOVE FROM %s WHERE id = ?;
	`, categoryType, categoryType)

	row, err := c.db.Query(query, id)
	if err != nil {
		return err
	}

	var value string
	err = row.Scan(&value)
	if err != nil {
		return err
	}

	defer log.Println("Removed \"", value, "\" from", categoryType)

	return nil
}

func (c Client) GetCategory(tx *sql.Tx, categoryType CategoryType, id int) (Category, error) {

	indyTx := tx == nil
	if indyTx {
		tx, _ = c.db.Begin()
		defer tx.Rollback()
	}

	query := fmt.Sprintf(`
	SELECT * FROM %s WHERE id = ?
	`, categoryType)

	cat := Category{Type: categoryType}
	err := tx.QueryRow(query, id).Scan(&cat.Id, &cat.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Category{}, nil
		}
		return Category{}, err
	}

	if indyTx {
		err = tx.Commit()
		if err != nil {
			return Category{}, err
		}
	}

	defer log.Println("Retrieved \"", cat.Name, "\" from", categoryType)

	return cat, err
}

func (c Client) GetCategoryByValue(tx *sql.Tx, categoryType CategoryType, name string) (Category, error) {

	indyTx := tx == nil
	if indyTx {
		tx, _ = c.db.Begin()
		defer tx.Rollback()
	}

	query := fmt.Sprintf(`
	SELECT * FROM %s WHERE name = ?
	`, categoryType)

	cat := Category{Type: categoryType}
	err := c.db.QueryRow(query, name).Scan(&cat.Id, &cat.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Category{}, nil
		}
		return Category{}, err
	}

	defer log.Println("Retrieved \"", cat.Name, "\" from", categoryType)

	if indyTx {
		err = tx.Commit()
		if err != nil {
			return Category{}, err
		}
	}

	return cat, err
}

func (c Client) GetAllOfCategory(categoryType CategoryType) ([]Category, error) {

	rows, err := c.db.Query(fmt.Sprintf("SELECT * FROM %s", categoryType))
	if err != nil {
		return []Category{}, err
	}

	categories := []Category{}
	for rows.Next() {
		cat := Category{Type: categoryType}

		err = rows.Scan(&cat.Id, &cat.Name)
		if err != nil {
			return []Category{}, err
		}

		categories = append(categories, cat)
	}

	return categories, nil
}

func (c Client) associateBookAndCategoryType(tx *sql.Tx, bookId string, category Category, rank int) error {

	insertLine := ""
	var args []any
	if category.Type == Series {
		insertLine = fmt.Sprintf("INSERT INTO books_%s (book_id, %s_id, series_index, rank)\nSELECT b.id, c.id, ?, ?", category.Type, categorySingular[category.Type])
		args = []any{category.Index, rank, bookId, category.Id}
	} else {
		insertLine = fmt.Sprintf("INSERT INTO books_%s (book_id, %s_id, rank)\nSELECT b.id, c.id, ?", category.Type, categorySingular[category.Type])
		args = []any{rank, bookId, category.Id}
	}

	query := fmt.Sprintf(`
	%s
	FROM books AS b
	CROSS JOIN %s AS c
	WHERE b.id = ? AND c.id = ?
	`, insertLine, category.Type)

	_, err := tx.Exec(query, args...)
	if err != nil {
		fmt.Println(query)
		return err
	}

	if category.Type == Series {
		_, err := tx.Exec("UPDATE books_series SET series_index = ? WHERE book_id = ? AND series_id = ?", category.Index, bookId, category.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Client) GetCategoryTypesAssociatedWithBook(tx *sql.Tx, bookId string, categoryType CategoryType) ([]Category, error) {

	indyTx := tx == nil
	if indyTx {
		tx, _ = c.db.Begin()
		defer tx.Rollback()
	}

	selectLine := ""
	if categoryType == Series {
		selectLine = fmt.Sprintf("SELECT cat.*, jn.series_index FROM %s AS cat", categoryType)
	} else {
		selectLine = fmt.Sprintf("SELECT cat.* FROM %s AS cat", categoryType)
	}

	query := fmt.Sprintf(`
	%s
	INNER JOIN books_%s AS jn 
		ON cat.id = jn.%s_id
	WHERE jn.book_id = ?
	ORDER BY jn.rank ASC;
	`, selectLine, categoryType, categorySingular[categoryType])

	rows, err := tx.Query(query, bookId)
	if err != nil {
		//fmt.Println(query)
		return []Category{}, err
	}

	cats := []Category{}
	for rows.Next() {
		cat := Category{Type: categoryType}

		if categoryType == Series {
			err = rows.Scan(&cat.Id, &cat.Name, &cat.Index)
		} else {
			err = rows.Scan(&cat.Id, &cat.Name)
		}

		if err != nil {
			return []Category{}, err
		}
		cats = append(cats, cat)
	}

	if indyTx {
		err = tx.Commit()
		if err != nil {
			return []Category{}, err
		}
	}

	return cats, nil
}
