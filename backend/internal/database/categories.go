package database

import (
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
	Id    *int         `json:"id"`
	Value string       `json:"value"`
	Index *int         `json:"index,omitempty"`
	Type  CategoryType `json:"-"`
}

var categorySingular = map[CategoryType]string{
	Series:    "series",
	Genres:    "genre",
	Authors:   "author",
	Narrators: "narrator",
}

func (c Client) AddCategory(categoryType CategoryType, value string) (Category, error) {

	query := fmt.Sprintf(`
	INSERT INTO %s
		(id, name)
	VALUES
		(NULL, ?)
	`, categoryType)

	result, err := c.db.Exec(query, value)
	if err != nil {
		return Category{}, err
	}

	log.Println("Added \"", value, "\" to", categoryType)

	id, err := result.LastInsertId()
	if err != nil {
		return Category{}, err
	}

	return c.GetCategory(categoryType, int(id))
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

func (c Client) GetCategory(categoryType CategoryType, id int) (Category, error) {

	query := fmt.Sprintf(`
	SELECT * FROM %s WHERE id = ?
	`, categoryType)

	cat := Category{Type: categoryType}
	err := c.db.QueryRow(query, id).Scan(&cat.Id, &cat.Value)
	if err != nil {
		return Category{}, err
	}

	defer log.Println("Retrieved \"", cat.Value, "\" from", categoryType)

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

		err = rows.Scan(&cat.Id, &cat.Value)
		if err != nil {
			return []Category{}, err
		}

		categories = append(categories, cat)
	}

	return categories, nil
}

func (c Client) associateBookAndCategoryType(bookId string, category Category) error {

	query := fmt.Sprintf(`
	INSERT INTO books_%s (book_id, %s_id)
	SELECT b.id, c.id,
	FROM books b
	CROSS JOIN %s c
	WHERE b.id = ? AND c.id = ?
	`, category.Type, categorySingular[category.Type], category.Type)

	_, err := c.db.Exec(query, bookId, category.Id)
	if err != nil {
		return err
	}

	if category.Type == Series {
		_, err := c.db.Exec("UPDATE books_series SET series_index = ? WHERE book_id = ? AND series_id = ?", category.Index, bookId, category.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Client) getCategoryTypesAssociatedWithBook(bookId string, categoryType CategoryType) ([]Category, error) {

	selectLine := ""
	if categoryType == Series {
		selectLine = fmt.Sprintf("SELECT cat.*, books_%s.series_index FROM %s AS cat", categoryType, categoryType)
	} else {
		selectLine = fmt.Sprintf("SELECT cat.* FROM %s AS cat", categoryType)
	}

	query := fmt.Sprintf(`
	%s
	INNER JOIN books_%s AS join ON cat.id = join.%s_id
	INNER JOIN books ON join.books_id = ?;
	`, selectLine, categoryType, categorySingular[categoryType])

	rows, err := c.db.Query(query, bookId)
	if err != nil {
		return []Category{}, err
	}

	cats := []Category{}
	for rows.Next() {
		cat := Category{Type: categoryType}
		err = rows.Scan(&cat.Id, &cat.Value)
		if err != nil {
			return []Category{}, err
		}
		cats = append(cats, cat)
	}

	return cats, err
}
