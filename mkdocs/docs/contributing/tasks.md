---
template: main.html
---

# Writing Iter8 Tasks

Iter8 tasks are implemented in the [`handler` Go repo](https://github.com/iter8-tools/handler). Clone this repo.

## Go version
Ensure you have [Go version 1.16+](https://golang.org/dl/).

## Running `handler` locally
```
# cd <root of the locally cloned handler repo>
go build
./handler
```

## Testing `handler` with coverage
```
make test
# uncomment the line below to show coverage percentage
# make coverage
# uncomment the line below to view coverage in a browser
# make show-coverage
# uncomment the line below to sort functions in descending order of coverage
# go tool cover -func coverage.out | sort -nr -k 3 
```

## Implementing a new task
This section is coming soon.