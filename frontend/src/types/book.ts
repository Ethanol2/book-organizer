export type Book = {
    id: string
    title: string
    subtitle: string
    description: string
    year: string
    isbn: string
    asin: string
    tags: string[]
    publisher: string
    created_at: string
    updated_at: string
    series: Series[]
    authors: Author[]
    genres: Genre[]
    narrators: Narrator[]
    files: BookFiles
};

export type BookFiles = {
    root?: string | null,
    cover?: string | null,
    audio_files?: string[] | null,
    text_files?: string[] | null
};

export type Category = {
    id: string
    name: string
}

export type Series = Category & { index: string }
export type Author = Category
export type Narrator = Category
export type Genre = Category

export function getBookCoverSrc(book: Book): string {
    if (book.files != null && book.files.cover != null) {
        return book.files.cover
    }
    return ""
}

export function getAuthorsList(book: Book): string {
    if (book.authors.length == 0) {
        return ""
    }

    let list = ""
    book.authors.forEach(author => {
        list += ', ' + author.name
    });
    return list.slice(2, list.length)
}