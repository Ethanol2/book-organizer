import type { BookFiles } from "./book"

export type Download = {
    id: string,
    created_at: string,
    files: BookFiles
}