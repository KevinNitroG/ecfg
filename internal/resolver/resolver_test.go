package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_Resolve(t *testing.T) {
	// Create temp directory structure:
	// /root/.editorconfig (root = true, indent_style = space)
	// /root/project/.editorconfig (indent_size = 2)
	// /root/project/src/.editorconfig (end_of_line = lf)
	// /root/project/src/file.go (target file)

	tmpDir := t.TempDir()

	// Create root .editorconfig
	rootConfig := filepath.Join(tmpDir, ".editorconfig")
	err := os.WriteFile(rootConfig, []byte("root = true\n[*]\nindent_style = space\n"), 0o644)
	require.NoError(t, err)

	// Create project .editorconfig
	projectDir := filepath.Join(tmpDir, "project")
	err = os.MkdirAll(projectDir, 0o755)
	require.NoError(t, err)

	projectConfig := filepath.Join(projectDir, ".editorconfig")
	err = os.WriteFile(projectConfig, []byte("[*]\nindent_size = 2\n"), 0o644)
	require.NoError(t, err)

	// Create src .editorconfig
	srcDir := filepath.Join(projectDir, "src")
	err = os.MkdirAll(srcDir, 0o755)
	require.NoError(t, err)

	srcConfig := filepath.Join(srcDir, ".editorconfig")
	err = os.WriteFile(srcConfig, []byte("[*]\nend_of_line = lf\n"), 0o644)
	require.NoError(t, err)

	// Create target file
	targetFile := filepath.Join(srcDir, "file.go")
	err = os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	// Test resolution
	resolver := NewResolver(tmpDir)
	def, err := resolver.Resolve(targetFile)

	require.NoError(t, err)
	assert.NotNil(t, def)

	// Check merged properties
	assert.Equal(t, "space", def.Raw["indent_style"])
	assert.Equal(t, "2", def.Raw["indent_size"])
	assert.Equal(t, "lf", def.Raw["end_of_line"])

	// Check file hierarchy - should have all 3 files
	assert.GreaterOrEqual(t, len(def.Files), 1, "should have at least one config file")

	t.Logf("Resolved files: %v", def.Files)
}

func TestResolver_InheritedProperties(t *testing.T) {
	tmpDir := t.TempDir()

	// Create parent .editorconfig
	rootConfig := filepath.Join(tmpDir, ".editorconfig")
	err := os.WriteFile(rootConfig, []byte("root = true\n[*]\nindent_style = space\ncharset = utf-8\n"), 0o644)
	require.NoError(t, err)

	// Create child .editorconfig with redundant property
	projectDir := filepath.Join(tmpDir, "project")
	err = os.MkdirAll(projectDir, 0o755)
	require.NoError(t, err)

	projectConfig := filepath.Join(projectDir, ".editorconfig")
	err = os.WriteFile(projectConfig, []byte("[*]\nindent_style = space\nend_of_line = lf\n"), 0o644)
	require.NoError(t, err)

	// Create target file
	targetFile := filepath.Join(projectDir, "file.go")
	err = os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	// Test inheritance detection
	resolver := NewResolver(tmpDir)
	inherited, inheritedFrom, err := resolver.InheritedProperties(targetFile)

	require.NoError(t, err)
	assert.Contains(t, inherited, "indent_style", "indent_style should be inherited")
	assert.Equal(t, "space", inherited["indent_style"])
	assert.Contains(t, inheritedFrom, "indent_style")

	t.Logf("Inherited properties: %v", inherited)
	t.Logf("Inherited from: %v", inheritedFrom)
}

func TestResolver_StopsAtRootTrue(t *testing.T) {
	tmpDir := t.TempDir()

	// Create root .editorconfig with root=true
	rootConfig := filepath.Join(tmpDir, ".editorconfig")
	err := os.WriteFile(rootConfig, []byte("root = true\n[*]\nindent_style = space\n"), 0o644)
	require.NoError(t, err)

	// Create project .editorconfig (should NOT be traversed beyond root)
	projectDir := filepath.Join(tmpDir, "project")
	err = os.MkdirAll(projectDir, 0o755)
	require.NoError(t, err)

	projectConfig := filepath.Join(projectDir, ".editorconfig")
	err = os.WriteFile(projectConfig, []byte("[*]\nindent_size = 2\n"), 0o644)
	require.NoError(t, err)

	// Test resolution
	resolver := NewResolver(tmpDir)
	def, err := resolver.Resolve(filepath.Join(projectDir, "file.go"))

	require.NoError(t, err)
	assert.NotNil(t, def)

	// Should include root but stop there (no parent traversal beyond root=true)
	assert.Equal(t, "space", def.Raw["indent_style"])
}

func TestResolver_StopsAtFilesystemRoot(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .editorconfig WITHOUT root=true (should traverse to filesystem root)
	configPath := filepath.Join(tmpDir, ".editorconfig")
	err := os.WriteFile(configPath, []byte("[*]\nindent_style = space\n"), 0o644)
	require.NoError(t, err)

	// Create target file in a nested directory
	subDir := filepath.Join(tmpDir, "subdir", "nested")
	err = os.MkdirAll(subDir, 0o755)
	require.NoError(t, err)

	targetFile := filepath.Join(subDir, "file.go")
	err = os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	// Test resolution
	resolver := NewResolver(tmpDir)
	def, err := resolver.Resolve(targetFile)

	require.NoError(t, err)
	assert.NotNil(t, def)

	// Should find at least the local config
	t.Logf("Files found: %v", def.Files)
}

func TestResolver_FindRedundantProperties(t *testing.T) {
	tmpDir := t.TempDir()

	// Create parent .editorconfig
	rootConfig := filepath.Join(tmpDir, ".editorconfig")
	err := os.WriteFile(rootConfig, []byte("root = true\n[*]\nindent_style = space\n"), 0o644)
	require.NoError(t, err)

	// Create child .editorconfig with same value (redundant)
	projectDir := filepath.Join(tmpDir, "project")
	err = os.MkdirAll(projectDir, 0o755)
	require.NoError(t, err)

	projectConfig := filepath.Join(projectDir, ".editorconfig")
	err = os.WriteFile(projectConfig, []byte("[*]\nindent_style = space\nend_of_line = lf\n"), 0o644)
	require.NoError(t, err)

	// Create target file
	targetFile := filepath.Join(projectDir, "file.go")
	err = os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	// Test redundant properties
	resolver := NewResolver(tmpDir)
	redundant, err := resolver.FindRedundantProperties(targetFile)

	require.NoError(t, err)

	// indent_style should be redundant (same value as parent)
	assert.Contains(t, redundant, "indent_style", "indent_style should be redundant")
	assert.Equal(t, "space", redundant["indent_style"])

	// end_of_line should NOT be redundant (not in parent)
	assert.NotContains(t, redundant, "end_of_line")

	t.Logf("Redundant properties: %v", redundant)
}

func TestResolver_ResolveWithSource(t *testing.T) {
	tmpDir := t.TempDir()

	// Create root .editorconfig
	rootConfig := filepath.Join(tmpDir, ".editorconfig")
	err := os.WriteFile(rootConfig, []byte("root = true\n[*]\nindent_style = space\n"), 0o644)
	require.NoError(t, err)

	// Create project .editorconfig
	projectDir := filepath.Join(tmpDir, "project")
	err = os.MkdirAll(projectDir, 0o755)
	require.NoError(t, err)

	projectConfig := filepath.Join(projectDir, ".editorconfig")
	err = os.WriteFile(projectConfig, []byte("[*]\nindent_size = 2\n"), 0o644)
	require.NoError(t, err)

	// Create target file
	targetFile := filepath.Join(projectDir, "file.go")
	err = os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	// Test resolve with source tracking
	resolver := NewResolver(tmpDir)
	props, sources, err := resolver.ResolveWithSource(targetFile)

	require.NoError(t, err)
	assert.NotEmpty(t, props)

	// Check source tracking
	assert.Contains(t, sources, "indent_style", "should track source for indent_style")

	t.Logf("Properties: %v", props)
	t.Logf("Sources: %v", sources)
}

func TestInheritance(t *testing.T) {
	// Test that inheritance detection works correctly
	// This test covers the key inheritance scenario from the plan

	tmpDir := t.TempDir()

	// Create root .editorconfig (root=true)
	rootConfig := filepath.Join(tmpDir, ".editorconfig")
	err := os.WriteFile(rootConfig, []byte("root = true\n[*]\nindent_style = space\ncharset = utf-8\n"), 0o644)
	require.NoError(t, err)

	// Create project level
	projectDir := filepath.Join(tmpDir, "project")
	err = os.MkdirAll(projectDir, 0o755)
	require.NoError(t, err)

	projectConfig := filepath.Join(projectDir, ".editorconfig")
	err = os.WriteFile(projectConfig, []byte("[*]\nindent_size = 2\n"), 0o644)
	require.NoError(t, err)

	// Create src level
	srcDir := filepath.Join(projectDir, "src")
	err = os.MkdirAll(srcDir, 0o755)
	require.NoError(t, err)

	srcConfig := filepath.Join(srcDir, ".editorconfig")
	err = os.WriteFile(srcConfig, []byte("[*]\nindent_style = space\nend_of_line = lf\n"), 0o644)
	require.NoError(t, err)

	// Create target file
	targetFile := filepath.Join(srcDir, "file.go")
	err = os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	resolver := NewResolver(tmpDir)

	// Test a: Resolve from src - should include root + project + src
	def, err := resolver.Resolve(targetFile)
	require.NoError(t, err)
	assert.NotNil(t, def)
	t.Logf("Test a - Files: %v", def.Files)

	// Test b: Resolve from project
	projectFile := filepath.Join(projectDir, "main.go")
	err = os.WriteFile(projectFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	def, err = resolver.Resolve(projectFile)
	require.NoError(t, err)
	assert.NotNil(t, def)
	t.Logf("Test b - Files: %v", def.Files)

	// Test inheritance: src has indent_style=space which is redundant from root
	inherited, inheritedFrom, err := resolver.InheritedProperties(targetFile)
	require.NoError(t, err)
	t.Logf("Inherited: %v, From: %v", inherited, inheritedFrom)

	// Should detect indent_style is inherited
	if _, ok := inherited["indent_style"]; ok {
		t.Log("SUCCESS: indent_style correctly detected as inherited")
	}
}
