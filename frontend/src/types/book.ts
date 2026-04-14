export type Book = {
    id: string
    title: string
    subtitle: string
    description: string
    year: number
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

export type BookParams = {
  title?: string | null
  subtitle?: string | null
  description?: string | null
  year?: number | null
  isbn?: string | null
  publisher?: string | null

  series?: Series[] | null
  authors?: Author[] | null
  genres?: Genre[] | null
  narrators?: Narrator[] | null

  cover?: string | null
  key?: string | null
}

export type BookSummary = {
    id: string
    title: string
    subtitle: string
    cover?: string | null
    authors: Author[]
}

export type BookFiles = {
    root?: string | null,
    cover?: string | null,
    audio_files?: string[] | null,
    text_files?: string[] | null
};

export type Category = {
    id: string | null
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

export function getAuthorsList(authors: Author[]): string {
    if (authors.length == 0) {
        return ""
    }

    let list = ""
    authors.forEach(author => {
        list += ', ' + author.name
    });
    return list.slice(2, list.length)
}

export function getCategoriesString(categories: Category[]): string {
    if (categories.length === 0) {
        return ''
    }

    return categories.map(category => category.name).join(', ')
}

export function getCategoriesArray(categories: string): Category[] {
    if (categories.trim() === "") {
        return []
    }

    return categories.split(',').map(c => ({ id: null, name: c.trim() }))
}

export function getSeriesString(series: Series[]): string {
    if (series.length === 0) {
        return ''
    }

    return series.map(s => `${s.name}` + (s.index ? ` #${s.index}` : '')).join(', ')
}
export function getSeriesArray(series: string): Series[] {
    if (series.trim() === "") {
        return []
    }

    const seriesArray: Series[] = []
    const seriesItems = series.split(',')
    seriesItems.forEach(item => {
        const itemSplit = item.trim().split('#')
        var name: string;
        var index: string;
        if (itemSplit.length == 0) {
            return
        } else if (itemSplit.length == 1) {
            name = itemSplit[0] === null || itemSplit[0] === undefined ? '' : itemSplit[0].trim()
            index = ''
        } else {
            name = itemSplit[0] === null || itemSplit[0] === undefined ? '' : itemSplit[0].trim()
            index = itemSplit[1] === null || itemSplit[1] === undefined ? '' : itemSplit[1].trim()
        }

        if (name === '') {
            return
        }
        seriesArray.push({ id: null, name, index })
    })
    return seriesArray
}