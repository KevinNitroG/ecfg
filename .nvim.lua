vim.filetype.add({
	extension = {
		editorconfig = "editorconfig",
	},
})

vim.lsp.config.ecfg = {
	cmd = { vim.fn.expand("%:p:h") .. "/ecfg-lsp" },
	filetypes = { "editorconfig" },
	root_markers = { ".editorconfig" },
}

vim.lsp.enable("ecfg")
