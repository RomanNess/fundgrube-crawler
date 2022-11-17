# fundgrube-crawler [WIP]

Quick and dirty Crawler to find deals in Fundgrube by MediaMarkt and Saturn because their shop systems are so horrible.

Implemented in Golang and using MongoDB as persistence since I usually don't use these. :)

## Use Case

- Cross compiled and deployed on a raspberry pi.
- MongoDB Atlas free tier used for persistence.
- Is currently started as a script that's configured with env vars.
  Will fetch all Postings in a Category and search for matching Deals.
- Alerts via email when new Deals are found.

## Environmental Variables

| name                            | desc                                                   | default                     |
|---------------------------------|--------------------------------------------------------|-----------------------------|
| `MONGODB_URI`                   | -                                                      | `mongodb://localhost:27017` |
| `MONGODB_USERNAME`              | -                                                      | `root`                      |
| `MONGODB_PASSWORD`              | -                                                      | `example`                   |
| `MONGODB_DB`                    | -                                                      | `fundgrube`                 |
| `MONGODB_COLLECTION_POSTINGS`   | -                                                      | `postings`                  |
| `MONGODB_COLLECTION_OPERATIONS` | -                                                      | `operations`                |
| `FIND_ALL`                      | ignore last run and search in all postings             | `false`                     |
| `LIMIT_OUTLETS`                 | only fetch 5 first outlets (for development)           | `false`                     |
| `LOG_TO_FILE`                   | log to /tmp/fundgrube.txt instead of stdout            | `false`                     |
| `MOCKED_POSTINGS`               | mock response from api                                 | `false`                     |
| `SKIP_CRAWLING`                 | skip fetching postings from api                        | `false`                     |
| `FAST_CRAWLING`                 | stop crawling api when no new postings on current page | `false`                     |
| `LOG_LEVEL`                     | levels: trace, debug, info, warn, error, fatal, panic  | `info`                      |

## API peculiarities

- There is only a `/api/postings` endpoint known to me, but it also returns a list of `outlets` and `brands` in the
  response.
- Requests with a `limit >= 100` always return the first page.
- Requests with an `offset > 990` return `422 Unprocessable Entity`, so you need to iterate over `brands` or `outlets`
  to see all `postings`.

## Shell script

https://github.com/RomanNess/fundgrube-crawler/issues/1 inspired me to quickly hack my initial idea that solves the same
use case with a `bash` script.
Simply run `SEARCH_REGEX="sony walkman" ./_sh/fundgrube-crawler.sh` to search for matching postings.
The script creates a TSV file with previous results in `/tmp` and will therefore only discover new postings.

## MongoDB migrations
[`cmd/fundgrube-migrate/main.go`](cmd/fundgrube-migrate/main.go) contains poor man's db migrations and a tooling to clean up obsolete db entries.
* Provide a `filterString` and `updateString` to perform a migration on the `postings` collection with `MIGRATE=true make migrate`.
* Provide a `filterString` to delete entries from the `postings` collection with `CLEANUP=true make migrate`.

If the env var is not provided a dry run with the `filterString` is performed in both cases.