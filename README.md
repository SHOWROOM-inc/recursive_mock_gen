# recursive_mock_gen

It is the tool to generate mocks recursive from the specified directory in parallel.
And skip not changed interfaces by using a previously done cache.

## Installation

1. Install [mockgen](https://github.com/uber-go/mock) to generate mocks.
2. Install this tool. 

Please do bellow commands.
```shell
go install go.uber.org/mock/mockgen@latest
go install github.com/SHOWROOM-inc/recursive_mock_gen/cmd/recursive_mock_gen@latest
```

## Usage
```shell
# Generate mocks by scan go files from current directory.
recursive_mock_gen --output testing/mocks

# In case Specify parameters.
recursive_mock_gen --input . --output testing/mocks --cache .mock_cache.json --max-parallels 100 
```

- --input: root directory path to source files (default is ".", current directory)
- --output: directory path to output mock files (required)
- --cache: path to the cache file (default is ".mockgen-cache.json")
- --max-parallels: the number of parallels (default is 100)
