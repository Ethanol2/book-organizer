package database

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) Client {
	dbPath := "/tmp/test_book_organizer.db"
	os.Remove(dbPath) // Clean up any previous test db
	client, err := NewClient(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	return client
}

func TestAddBook(t *testing.T) {
	client := setupTestDB(t)
	defer client.db.Close()

	title := "Test Book"
	params := BookParams{
		Title: &title,
	}

	book, err := client.AddBook(params)
	if err != nil {
		t.Fatalf("AddBook failed: %v", err)
	}

	if book.Id == nil {
		t.Error("Book ID not set")
	}

	if book.Title != title {
		t.Errorf("Expected title %s, got %s", title, book.Title)
	}
}

func TestGetBook(t *testing.T) {
	client := setupTestDB(t)
	defer client.db.Close()

	title := "Test Book"
	params := BookParams{
		Title: &title,
	}

	addedBook, err := client.AddBook(params)
	if err != nil {
		t.Fatalf("AddBook failed: %v", err)
	}

	retrievedBook, err := client.GetBook(*addedBook.Id)
	if err != nil {
		t.Fatalf("GetBook failed: %v", err)
	}

	if retrievedBook.Title != title {
		t.Errorf("Expected title %s, got %s", title, retrievedBook.Title)
	}

	if *retrievedBook.Id != *addedBook.Id {
		t.Error("Retrieved book ID does not match added book ID")
	}
}

func TestUpdateBook(t *testing.T) {
	client := setupTestDB(t)
	defer client.db.Close()

	title := "Test Book"
	year := 2020
	params := BookParams{
		Title: &title,
		Year:  &year,
	}

	addedBook, err := client.AddBook(params)
	if err != nil {
		t.Fatalf("AddBook failed: %v", err)
	}

	newYear := 2021
	updateParams := BookParams{
		Year: &newYear,
	}

	updatedBook, needsFileUpdate, err := client.UpdateBook(*addedBook.Id, updateParams)
	if err != nil {
		t.Fatalf("UpdateBook failed: %v", err)
	}

	if needsFileUpdate {
		t.Error("File update should not be needed for year change")
	}

	if updatedBook.Year == nil || *updatedBook.Year != newYear {
		t.Errorf("Expected updated year %d, got %v", newYear, *updatedBook.Year)
	}
}

func TestDeleteBook(t *testing.T) {
	client := setupTestDB(t)
	defer client.db.Close()

	title := "Test Book"
	params := BookParams{
		Title: &title,
	}

	addedBook, err := client.AddBook(params)
	if err != nil {
		t.Fatalf("AddBook failed: %v", err)
	}

	err = client.DeleteBook(*addedBook.Id)
	if err != nil {
		t.Fatalf("DeleteBook failed: %v", err)
	}

	_, err = client.GetBook(*addedBook.Id)
	if err == nil {
		t.Error("GetBook should fail after deletion")
	}
}
