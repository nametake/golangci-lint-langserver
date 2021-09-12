# golangci-lint-langserver

golangci-lint-langserver is [golangci-lint](https://github.com/golangci/golangci-lint) language server.

[![asciicast](https://asciinema.org/a/308369.svg)](https://asciinema.org/a/308369)


## Installation

```console
go get github.com/nametake/golangci-lint-langserver
```

## Options

```console
  -debug
        output debug log
  -nolintername
        don't show a linter name in message
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

### Configuration for [vim-lsp](https://github.com/prabirshrestha/vim-lsp)

```vim
augroup vim_lsp_golangci_lint_langserver
  au!
  autocmd User lsp_setup call lsp#register_server({
      \ 'name': 'golangci-lint-langserver',
      \ 'cmd': {server_info->['golangci-lint-langserver']},
      \ 'initialization_options': {'command': ['golangci-lint', 'run', '--enable-all', '--disable', 'lll', '--out-format', 'json']},
      \ 'whitelist': ['go'],
      \ })
augroup END
```

[vim-lsp-settings](https://github.com/mattn/vim-lsp-settings) provide installer for golangci-lint-langserver.

### Configuration for [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig)

**Requires [Neovim HEAD/nightly](https://github.com/neovim/neovim/releases/tag/nightly) (v0.5 prerelease).**

```lua
local lspconfig = require 'lspconfig'
local configs = require 'lspconfig/configs'

if not lspconfig.golangcilsp then
 	configs.golangcilsp = {
		default_config = {
			cmd = {'golangci-lint-langserver'},
			root_dir = lspconfig.util.root_pattern('.git', 'go.mod'),
			init_options = {
					command = { "golangci-lint", "run", "--enable-all", "--disable", "lll", "--out-format", "json" };
			}
		};
	}
end
lspconfig.golangcilsp.setup {
	filetypes = {'go'}
}
```
