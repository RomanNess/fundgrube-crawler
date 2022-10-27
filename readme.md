# fundgrube-crawler [WIP]

Quick and dirty Crawler to find deals in Fundgrube by MediaMarkt and Saturn because their shop systems are so horrible.

Implemented in Golang and using MongoDB as persistence since I usually don't use these. :)

# Environmental Variables
| name                            | desc                                         | default                     |
|---------------------------------|----------------------------------------------|-----------------------------|
| `MONGODB_URI`                   | -                                            | `mongodb://localhost:27017` |
| `MONGODB_USERNAME`              | -                                            | `root`                      |
| `MONGODB_PASSWORD`              | -                                            | `example`                   |
| `MONGODB_DB`                    | -                                            | `fundgrube`                 |
| `MONGODB_COLLECTION_POSTINGS`   | -                                            | `postings`                  |
| `MONGODB_COLLECTION_OPERATIONS` | -                                            | `operations`                |
| `SKIP_CRAWLING`                 | skip fetching postings from api              | `false`                     |
| `MOCKED_POSTINGS`               | mock response from api                       | `false`                     |
| `LIMIT_OUTLETS`                 | only fetch 5 first outlets (for development) | `false`                     |
| `LOG_TO_FILE`                   | log to /tmp/fundgrube.txt instead of stdout  | `false`                     |
| `SEARCH_KEYWORD`                | keyword used to search for deals             | `example`                   |
