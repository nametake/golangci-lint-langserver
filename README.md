# golangci-lint-langserver

golangci-lint-langserver is [golangci-lint](https://github.com/golangci/golangci-lint) language server.

[![asciicast](https://asciinema.org/a/308369.svg)](https://asciinema.org/a/308369)


## Installation

Install [golangci-lint](https://golangci-lint.run).

```console
go install github.com/nametake/golangci-lint-langserver@latest
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
        "command": ["golangci-lint", "run", "--output.json.path", "stdout", "--show-stats=false", "--issues-exit-code=1"]
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
      \ 'initialization_options': {'command': ['golangci-lint', 'run', '--output.json.path', 'stdout', '--show-stats=false', '--issues-exit-code=1']},
      \ 'whitelist': ['go'],
      \ })
augroup END
```

[vim-lsp-settings](https://github.com/mattn/vim-lsp-settings) provide installer for golangci-lint-langserver.

### Configuration for [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig)

**Requires Neovim [v0.6.1](https://github.com/neovim/neovim/releases/tag/v0.6.1) or [nightly](https://github.com/neovim/neovim/releases/tag/nightly).**

```lua
local lspconfig = require 'lspconfig'
local configs = require 'lspconfig/configs'

if not configs.golangcilsp then
 	configs.golangcilsp = {
		default_config = {
			cmd = {'golangci-lint-langserver'},
			root_dir = lspconfig.util.root_pattern('.git', 'go.mod'),
			init_options = {
					command = { "golangci-lint", "run", "--output.json.path", "stdout", "--show-stats=false", "--issues-exit-code=1" };
			};
		}
	}
end
lspconfig.golangci_lint_ls.setup {
	filetypes = {'go','gomod'}
}
```

### Configuration for [lsp-mode](https://github.com/emacs-lsp/lsp-mode) (Emacs)

Support for golangci-lint-langserver is
[built-in](https://github.com/emacs-lsp/lsp-mode/blob/master/clients/lsp-golangci-lint.el)
to lsp-mode since late 2023. When the `golangci-lint-langserver` executable is
found, it is automatically started for Go buffers as an add-on server along with
the `gopls` language server.

### Configuration for [helix](https://helix-editor.com/)

You can use `.golangci.yaml` in the project root directory to enable other [linters](https://golangci-lint.run/usage/linters/)

```toml
[[language]]
name = "go"
auto-format = true
language-servers = [ "gopls", "golangci-lint-lsp" ]

[language-server.golangci-lint-lsp]
command = "golangci-lint-langserver"

[language-server.golangci-lint-lsp.config]
command = ["golangci-lint", "run", "--output.json.path", "stdout", "--show-stats=false", "--issues-exit-code=1"]
```

## golangci-lint Version Compatibility

- For golangci-lint v2+: Use `--output.json.path stdout --show-stats=false` parameters
- For golangci-lint v1: Use `--out-format json` parameter
