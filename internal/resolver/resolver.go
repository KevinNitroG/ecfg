// Package resolver provides file system resolution for EditorConfig inheritance.
//
// This package wraps editorconfig-core-go to provide:
// - Resolution of EditorConfig properties for a given file path
// - Detection of the hierarchy of .editorconfig files
// - Identification of inherited (redundant) properties from parent files
package resolver

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/editorconfig/editorconfig-core-go/v2"
)

// Definition represents a resolved EditorConfig definition with its source files.
type Definition struct {
	// Raw contains the merged properties from all .editorconfig files in the hierarchy.
	Raw map[string]string

	// Files lists the .editorconfig files in the hierarchy (from root to nearest).
	// This is used to track where properties come from.
	Files []string
}

// Resolver handles EditorConfig resolution for file paths within a workspace.
type Resolver struct {
	rootDir string
}

// NewResolver creates a new Resolver for the given workspace root directory.
// The rootDir should be the root of the project/workspace.
func NewResolver(rootDir string) *Resolver {
	return &Resolver{
		rootDir: filepath.Clean(rootDir),
	}
}

// Resolve resolves the EditorConfig definition for a given file path.
// It walks up the directory tree to find all .editorconfig files,
// merges their properties, and returns the combined definition.
//
// The function stops at a directory containing root=true or at the filesystem root.
func (r *Resolver) Resolve(path string) (*Definition, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Get definition from the library (handles root=true and filesystem root)
	def, err := editorconfig.GetDefinitionForFilename(absPath)
	if err != nil {
		return nil, err
	}

	// Collect all .editorconfig files in the hierarchy
	files, err := r.collectEditorconfigFiles(absPath)
	if err != nil {
		return nil, err
	}

	return &Definition{
		Raw:   def.Raw,
		Files: files,
	}, nil
}

// collectEditorconfigFiles walks from the file's directory up to the root,
// collecting all .editorconfig files encountered.
func (r *Resolver) collectEditorconfigFiles(filePath string) ([]string, error) {
	var files []string

	dir := filepath.Dir(filePath)
	visited := make(map[string]bool)

	for r.rootDir == "" || strings.HasPrefix(dir, r.rootDir) {

		// Avoid infinite loop
		if visited[dir] {
			break
		}
		visited[dir] = true

		configPath := filepath.Join(dir, ".editorconfig")
		info, err := os.Stat(configPath)
		if err == nil && !info.IsDir() {
			files = append(files, configPath)
		}

		// Check if this directory has root=true
		if r.isRootDir(dir) {
			break
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	// Reverse to get root-to-nearest order
	for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
		files[i], files[j] = files[j], files[i]
	}

	return files, nil
}

// isRootDir checks if the directory contains a .editorconfig with root=true.
func (r *Resolver) isRootDir(dir string) bool {
	configPath := filepath.Join(dir, ".editorconfig")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	// Simple check for root=true in the file
	content := strings.ToLower(string(data))
	return strings.Contains(content, "root=true") || strings.Contains(content, "root = true")
}

// InheritedProperties analyzes properties in the given file and identifies
// which ones are inherited from parent .editorconfig files.
//
// Returns:
// - inherited: map of property keys that exist in both current and parent files
// - inheritedFrom: map of property keys to the parent file path where they're defined
func (r *Resolver) InheritedProperties(path string) (inherited map[string]string, inheritedFrom map[string]string, err error) {
	inherited = make(map[string]string)
	inheritedFrom = make(map[string]string)

	// Get parent directory properties
	parentProps, parentFiles, err := r.getParentProperties(path)
	if err != nil {
		return nil, nil, err
	}

	if len(parentProps) == 0 {
		// No parent files, nothing to inherit
		return inherited, inheritedFrom, nil
	}

	// Parse current file's properties
	currentProps, err := r.getCurrentProperties(path)
	if err != nil {
		return nil, nil, err
	}

	// Find properties that exist in both current and parent
	for key, currentValue := range currentProps {
		if _, exists := parentProps[key]; exists {
			// Property exists in both - it's inherited
			inherited[key] = currentValue
			inheritedFrom[key] = parentFiles[key]
		}
	}

	return inherited, inheritedFrom, nil
}

// getParentProperties returns properties from parent .editorconfig files only
// (not including the file's own directory).
func (r *Resolver) getParentProperties(path string) (map[string]string, map[string]string, error) {
	props := make(map[string]string)
	files := make(map[string]string)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, err
	}

	dir := filepath.Dir(absPath)

	// Walk up from parent directory
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		configPath := filepath.Join(parent, ".editorconfig")
		info, err := os.Stat(configPath)
		if err == nil && !info.IsDir() {
			// Parse this parent .editorconfig
			def, err := editorconfig.GetDefinitionForFilename(configPath)
			if err == nil {
				for k, v := range def.Raw {
					props[k] = v
					files[k] = configPath
				}
			}
		}

		// Check if parent is root
		if r.isRootDir(parent) {
			break
		}

		dir = parent
	}

	return props, files, nil
}

// getCurrentProperties returns properties defined in the .editorconfig
// closest to the given file (in its directory).
func (r *Resolver) getCurrentProperties(path string) (map[string]string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(absPath)
	configPath := filepath.Join(dir, ".editorconfig")

	_, err = os.Stat(configPath)
	if err != nil {
		// No .editorconfig in this directory
		return make(map[string]string), nil
	}

	def, err := editorconfig.GetDefinitionForFilename(configPath)
	if err != nil {
		return nil, err
	}

	return def.Raw, nil
}

// ResolveWithSource resolves EditorConfig for a file and identifies
// which file each property comes from.
func (r *Resolver) ResolveWithSource(path string) (map[string]string, map[string]string, error) {
	def, err := r.Resolve(path)
	if err != nil {
		return nil, nil, err
	}

	// Build a map of property -> source file
	propSource := make(map[string]string)

	// For each file in hierarchy, parse and track properties
	for _, file := range def.Files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Parse the file
		ec, err := editorconfig.Parse(strings.NewReader(string(data)))
		if err != nil {
			continue
		}

		// Add definitions from this file
		for _, d := range ec.Definitions {
			for k := range d.Raw {
				// Later files override earlier ones
				propSource[k] = file
			}
		}
	}

	return def.Raw, propSource, nil
}

// FindRedundantProperties finds properties in the current file that are redundant
// because they're inherited from a parent .editorconfig file with the same value.
func (r *Resolver) FindRedundantProperties(path string) (map[string]string, error) {
	redundant := make(map[string]string)

	// Get parent properties
	parentProps, _, err := r.getParentProperties(path)
	if err != nil {
		return nil, err
	}

	if len(parentProps) == 0 {
		return redundant, nil
	}

	// Get current file properties
	currentProps, err := r.getCurrentProperties(path)
	if err != nil {
		return nil, err
	}

	// Find redundant properties (same value = inherited)
	for key, currentValue := range currentProps {
		if parentValue, exists := parentProps[key]; exists && currentValue == parentValue {
			redundant[key] = parentValue
		}
	}

	return redundant, nil
}

// IsEmpty checks if the error is an empty result (no .editorconfig found).
func IsEmpty(err error) bool {
	return errors.Is(err, errEmpty)
}

var errEmpty = errors.New("no editorconfig found")
