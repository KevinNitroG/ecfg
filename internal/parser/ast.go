// AST node types for EditorConfig parser.
//
// The AST structure mirrors EditorConfig's document model:
//   - Document: Root node containing preamble and sections
//   - Preamble: Optional key-value pairs before the first section (can contain root=true)
//   - Section: Glob pattern header ([*.go]) with associated key-value pairs
//   - KeyValue: Property assignment (indent_style = tab)
//   - Comment: Comment lines (# or ; prefix)
//
// Every node includes a Range field for precise position tracking, enabling
// LSP features like hover tooltips, diagnostics, and completion suggestions.
//
// Error Recovery:
//   - Parser continues after errors, collecting them in Document.Errors
//   - Partial nodes created for incomplete syntax (missing ], missing =, etc.)
//   - Never panics - always returns a valid Document even for malformed input
package parser

// Node is the base interface for all AST nodes.
// Every node provides its position range and type identifier.
type Node interface {
	GetRange() Range
	Type() NodeType
}

// NodeType identifies the type of an AST node.
type NodeType int

const (
	NodeDocument NodeType = iota // Root document node
	NodeComment                  // Comment line
	NodePreamble                 // Preamble section (before first [section])
	NodeSection                  // Section with glob header
	NodeKeyValue                 // Key-value pair (property = value)
)

// String returns the name of the node type.
func (n NodeType) String() string {
	switch n {
	case NodeDocument:
		return "Document"
	case NodeComment:
		return "Comment"
	case NodePreamble:
		return "Preamble"
	case NodeSection:
		return "Section"
	case NodeKeyValue:
		return "KeyValue"
	default:
		return "Unknown"
	}
}

// Document represents the root of an EditorConfig file AST.
// It contains an optional preamble (key-value pairs before the first section),
// zero or more sections, top-level comments, and any parse errors encountered.
type Document struct {
	Range    Range        // Position of entire document
	Preamble *Preamble    // Optional preamble (nil if no key-value pairs before first section)
	Sections []*Section   // Sections in document order
	Comments []*Comment   // Top-level comments (not associated with preamble or sections)
	Errors   []ParseError // Parse errors collected during parsing
}

// GetRange returns the position range of the document.
func (d *Document) GetRange() Range { return d.Range }

// Type returns NodeDocument.
func (d *Document) Type() NodeType { return NodeDocument }

// Preamble represents key-value pairs before the first section.
// The preamble is optional and typically contains root=true or global properties.
type Preamble struct {
	Range    Range       // Position of preamble content
	Pairs    []*KeyValue // Key-value pairs in the preamble
	Comments []*Comment  // Comments within the preamble
}

// GetRange returns the position range of the preamble.
func (p *Preamble) GetRange() Range { return p.Range }

// Type returns NodePreamble.
func (p *Preamble) Type() NodeType { return NodePreamble }

// Section represents a section with a glob pattern header and associated properties.
// Example: [*.go] followed by indent_style = tab
type Section struct {
	Range       Range       // Position of entire section (from [ to last key-value)
	Header      string      // Glob pattern without brackets (e.g., "*.go")
	HeaderRange Range       // Position of [header] for diagnostics
	Pairs       []*KeyValue // Key-value pairs within this section
	Comments    []*Comment  // Comments within this section
}

// GetRange returns the position range of the section.
func (s *Section) GetRange() Range { return s.Range }

// Type returns NodeSection.
func (s *Section) Type() NodeType { return NodeSection }

// KeyValue represents a property assignment (key = value).
// Tracks separate ranges for key and value to support hover and completion.
type KeyValue struct {
	Range      Range  // Position of entire key-value line
	Key        string // Property key (trimmed)
	KeyRange   Range  // Position of key for hover tooltip
	Value      string // Property value (trimmed)
	ValueRange Range  // Position of value for completion after =
}

// GetRange returns the position range of the key-value pair.
func (kv *KeyValue) GetRange() Range { return kv.Range }

// Type returns NodeKeyValue.
func (kv *KeyValue) Type() NodeType { return NodeKeyValue }

// Comment represents a comment line (# or ; prefix).
// Comments are preserved in the AST for potential formatting features.
type Comment struct {
	Range Range  // Position of comment line
	Text  string // Comment text including # or ; prefix
}

// GetRange returns the position range of the comment.
func (c *Comment) GetRange() Range { return c.Range }

// Type returns NodeComment.
func (c *Comment) Type() NodeType { return NodeComment }

// ParseError represents an error encountered during parsing.
// Collected in Document.Errors to allow partial AST construction.
type ParseError struct {
	Range   Range  // Position where error occurred
	Message string // Human-readable error message
	Code    string // Machine-readable error code (e.g., "missing-section-close")
}

// Error returns the error message.
func (e ParseError) Error() string {
	return e.Message
}
