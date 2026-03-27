# Neovim Setup with vim.lsp.config

This guide covers setting up the EditorConfig LSP server using Neovim's built-in `vim.lsp.config` (Neovim 0.10+).

## Requirements

- Neovim 0.10 or later
- Built `ecfg-lsp` binary (see [Build](#build))

## Build

```bash
# Build the LSP server binary
go build -o ecfg-lsp ./cmd/ecfg-lsp

# Or use the Makefile
make build
```

Place the `ecfg-lsp` binary in your PATH, or use the full path in the config below.

## Verify Binary Works

Before configuring Neovim, verify the binary works:

```bash
# Test the server responds to initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":123,"rootUri":"file:///tmp","capabilities":{}}}' | ./ecfg-lsp

# Should return a response with capabilities
```

## Configuration

Add to your Neovim config (`~/.config/nvim/init.lua`):

```lua
-- Register the EditorConfig LSP server
vim.lsp.config['ecfg'] = {
  -- Command to run the server
  -- Replace with full path if not in PATH
  cmd = { '/home/kevinnitro/projects/ecfg/ecfg-lsp' },

  -- Filetypes this server handles
  filetypes = { 'editorconfig' },

  -- Root directory detection
  root_dir = function(fname)
    return vim.fs.root(fname, { '.editorconfig' })
  end,
}

-- Start the server when you open an .editorconfig file
vim.lsp.enable('ecfg')
```

## Starting the Server

Manually start the server:

```vim
:LspStart ecfg
```

Or just open an `.editorconfig` file — it should auto-start.

## Verify in Neovim

Check server status:

```vim
:LspInfo
```

You should see something like:

```
  ecfg: active (attached)
```

## Troubleshooting

### Server starts but no features work

1. **Enable LSP logging in Neovim:**
   ```vim
   :lua vim.lsp.set_log_level('debug')
   ```
   Then check the log:
   ```vim
   :edit ~/.local/share/nvim/lsp.log
   ```

2. **Check server is actually receiving requests:**
   The binary now logs to stderr. Run in terminal to see:
   ```bash
   ./ecfg-lsp 2>&1
   ```
   Then send LSP messages manually.

3. **Check if Neovim is sending didOpen:**
   Add this to your config to confirm document is opened:
   ```lua
   -- In init.lua, after lsp.config
   vim.api.nvim_create_autocmd('LspAttach', {
     callback = function(args)
       print('LSP Attached! Client: ' .. vim.inspect(args.data.client))
       print('Buffer: ' .. args.buf)
       
       -- Manually trigger diagnostics
       vim.lsp.buf_request(0, 'textDocument/didOpen', {
         textDocument = {
           uri = vim.uri_from_fname(vim.api.nvim_buf_get_name(0)),
           version = 1,
           text = table.concat(vim.api.nvim_buf_get_lines(0, 0, -1, false), '\n')
         }
       }, function(err, result)
         print('didOpen response:', err, result)
       end)
     end
   })
   ```

4. **Manual test with the binary:**
   Test the server manually to verify it works:
   ```bash
   # Start server
   ./ecfg-lsp
   
   # In another terminal, use Python to send LSP messages
   ```
   See the test script at `test_lsp.py` for a working example.

### Check capabilities are advertised

In Neovim, after starting the server:
```vim
:LspInfo
```

Should show capabilities including `hover` and `completion`.

### Server logs to stderr

The binary now logs all incoming requests to stderr. To see them:
```bash
# In one terminal, run server
./ecfg-lsp 2>&1

# In another, send test LSP messages
```