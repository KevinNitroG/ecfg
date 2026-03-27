# ecfg - EditorConfig Language Server

An LSP (Language Server Protocol) implementation for EditorConfig files, designed for Neovim and other LSP-compatible editors.

## Features

- **Diagnostics**: Real-time validation of `.editorconfig` files
- **Hover**: Inline documentation for EditorConfig properties
- **Completion**: Context-aware autocomplete for property names and values

## Installation

### Pre-built Binaries

Download the latest release for your platform from the releases page.

### Build from Source

```bash
go build -o ecfg-lsp ./cmd/ecfg-lsp
```

## Neovim Setup with lspconfig

Add to your Neovim configuration (Lua):

```lua
-- ~/.config/nvim/lua/plugins/lspconfig.lua
require('lspconfig').ecfg.setup({
  cmd = {'ecfg-lsp'},
  filetypes = {'editorconfig'},
  root_dir = function(fname)
    return require('lspconfig').util.find_file('.editorconfig', fname:match('(.*)/') .. '/') or vim.loop.cwd()
  end,
})
```

Or if using packer:

```lua
use({
  'neovim/nvim-lspconfig',
  config = function()
    require('lspconfig').ecfg.setup({
      cmd = {'ecfg-lsp'},
      filetypes = {'editorconfig'},
      root_dir = function(fname)
        return require('lspconfig').util.find_file('.editorconfig', fname:match('(.*)/') .. '/') or vim.loop.cwd()
      end,
    })
  end
})
```

## Capabilities

| Feature | Description |
|---------|-------------|
| **Hover** | Shows property documentation and valid values |
| **Completion** | Context-aware property and value autocomplete |
| **Diagnostics** | Validation errors, warnings, and info |

## Supported Properties

The server validates all standard EditorConfig properties including:

- `indent_style` (tab/space)
- `indent_size` (number/tab)
- `end_of_line` (lf/crlf/cr)
- `insert_final_newline` (boolean)
- `trim_trailing_whitespace` (boolean)
- `root` (boolean)
- `charset` (latin1/utf-8/utf-16be/utf-16le/utf-8-bom)
- `max_line_length` (number/off)

## Troubleshooting

### Server Not Starting

1. Ensure `ecfg-lsp` is in your PATH, or use full path in `cmd`:
   ```lua
   cmd = {'/path/to/ecfg-lsp'}
   ```

2. Check server status:
   ```vim
   :LspInfo
   ```

### Debugging

Add log level configuration to see verbose output:

```lua
require('lspconfig').ecfg.setup({
  cmd = {'ecfg-lsp'},
  filetypes = {'editorconfig'},
  settings = {
    ecfg = {
      trace = 'verbose'
    }
  }
})
```

## License

MIT