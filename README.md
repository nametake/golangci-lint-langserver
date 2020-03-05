# golangci-lint-langserver

golangci-lint-langserver is [golangci-lint](https://github.com/golangci/golangci-lint) language server.

![demo](https://raw.github.com/wiki/nametake/golangci-lint-langserver/img/demo.gif)

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
