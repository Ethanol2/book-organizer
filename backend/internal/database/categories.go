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

func (c Client) AddCategory(categoryType CategoryType, name string) (Category, error) {

	query := fmt.Sprintf(`
	INSERT INTO %s
		(id, name)
	VALUES
		(NULL, ?)
	ON CONFLICT(name)
		DO NOTHING
	RETURNING id
	`, categoryType)

	var id int
	err := c.handler.QueryRow(query, name).Scan(&id)
	if err != nil {
		return Category{}, err
	}

	log.Println("Added \"", name, "\" to", categoryType)

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

	row, err := c.handler.Query(query, id)
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
	err := c.handler.QueryRow(query, id).Scan(&cat.Id, &cat.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Category{}, nil
		}
		return Category{}, err
	}

	defer log.Println("Retrieved \"", cat.Name, "\" from", categoryType)

	return cat, err
}

func (c Client) GetAllOfCategory(categoryType CategoryType) ([]Category, error) {

	rows, err := c.handler.Query(fmt.Sprintf("SELECT * FROM %s", categoryType))
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

// The Type field of the category must not be nil
func (c Client) associateBookAndCategoryType(bookId string, category Category, rank int) error {

	if category.Id == nil {
		if ok, err := category.GetID(c.handler); err != nil {
			return err
		} else if !ok {
			cat, err := c.AddCategory(category.Type, category.Name)
			if err != nil {
				return err
			}
			cat.Index = category.Index
			category = cat
		}
	}

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

	_, err := c.handler.Exec(query, args...)
	if err != nil {
		fmt.Println(query)
		return err
	}

	if category.Type == Series {
		_, err := c.handler.Exec("UPDATE books_series SET series_index = ? WHERE book_id = ? AND series_id = ?", category.Index, bookId, category.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Client) GetCategoryTypesAssociatedWithBook(bookId string, categoryType CategoryType) ([]Category, error) {

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

	rows, err := c.handler.Query(query, bookId)
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

	return cats, nil
}

func CategoryToStrSlice(items []Category) []string {
	names := []string{}
	for i := range items {
		names = append(names, items[i].Name)
	}

	return names
}

func StrToCategorySlice(items []string) []Category {
	cats := []Category{}
	for i := range items {
		cats = append(cats, Category{
			Name: items[i],
		})
	}
	return cats
}

func stringToCategoryType(str string) CategoryType {
	switch str {
	case "authors", "author":
		return Authors
	case "narrators", "narrator":
		return Narrators
	case "genres", "genre":
		return Genres
	case "series":
		return Series
	default:
		return NoType
	}
}

func (c Client) CleanupCategories() error {

	cleanup := func(catType CategoryType) error {

		query := fmt.Sprintf(
			"DELETE FROM %s WHERE NOT EXISTS (SELECT 1 FROM books_%s WHERE %s.id = books_%s.%s_id)",
			catType, catType, catType, catType, categorySingular[catType])

		//log.Println(query)

		results, err := c.handler.Exec(query)
		if err != nil {
			return err
		}

		count, err := results.RowsAffected()
		if err != nil {
			log.Println("Failed to retrieve deleted rows from the category cleanup")
			return nil
		}

		if count > 0 {
			log.Println("Deleted", count, "rows from", catType)
		}
		return nil
	}

	err := cleanup(Authors)
	if err != nil {
		return err
	}
	err = cleanup(Genres)
	if err != nil {
		return err
	}
	err = cleanup(Series)
	if err != nil {
		return err
	}
	err = cleanup(Narrators)
	if err != nil {
		return err
	}

	return nil
}

func (cat *Category) GetID(handler Handler) (bool, error) {

	if cat.Type == NoType {
		return false, fmt.Errorf("category must have a type to find id")
	}
	if cat.Name == "" {
		return false, fmt.Errorf("category must have a value to find id")
	}

	query := fmt.Sprintf(`
	SELECT id FROM %s WHERE name = ?
	`, cat.Type)

	err := handler.QueryRow(query, cat.Name).Scan(&cat.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	// defer log.Println("Retrieved \"", cat.Name, "\" from", cat.Type)

	return true, nil
}
