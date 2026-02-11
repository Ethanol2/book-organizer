import type { BookFiles } from "./book"

export type Download = {
    id: string,
    created_at: string,
    files: BookFiles
}

export function getDownloadName(download: Download): string {
    if (download == null || download.files.root == "") {
        return "";
    }
    var items = download.files.root.split("/");
    return items[items.length - 1] ?? "";
}

export function getTimeAdded(download: Download): string {
    if (download == null) {
        return "";
    }

    var createTime = Date.parse(download.created_at);
    var since = (Date.now() - createTime) / 1000;

    // if it's been over a day
    if (since > 86400) {
        return "Added on " + (download.created_at.split("T")[0] ?? "");
    }

    var min = (since % 3600) / 60;
    var hours = Math.floor(since / 3600)

    return `Added ${hours == 0 ? "" : Math.round(hours) + " hours"} ${min == 0 ? "" : Math.round(min) + " minutes"} ago`
}