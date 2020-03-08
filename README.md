# golangci-lint-langserver

golangci-lint-langserver is [golangci-lint](https://github.com/golangci/golangci-lint) language server.

[![asciicast](https://asciinema.org/a/308369.svg)](https://asciinema.org/a/308369)


## Installation

```console
go get github.com/nametake/golangci-lint-langserver
```

## Configuration

You need to set golangci-lint command to initializationOptions with `--out-format json`.

### Configuration for [coc.nvim](https://github.com/neoclide/coc.nvim)

coc-settings.json

```jsonc
{
  "languageserver": {
    "golangci-lint-languageserver": {
      "command": "golangci-lint-langserver",
      "filetypes": ["go"],
      "initializationOptions": {
        "command": ["golangci-lint", "run", "--enable-all", "--disable", "lll", "--out-format", "json"]
      }
    }
  }
}
```
