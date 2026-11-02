package content

import (
	"fmt"
	"os"
	"path/filepath"
)

// WorkspaceReader reads the owner's current Markdown, including private drafts
// and documents hidden from the graph. Filesystem access is the authorization
// boundary. It never opens SQLite, acquires the writer lock or creates files.
type WorkspaceReader struct {
	root string
}

func OpenWorkspaceReader(root string) (*WorkspaceReader, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("document root must be a directory")
	}
	return &WorkspaceReader{root: resolved}, nil
}

// List returns a fresh snapshot with the same locale fallback and relation
// resolution as the workspace. It does not return publication or session data.
func (r *WorkspaceReader) List(locale string) ([]Document, error) {
	if !ValidLocale(locale) {
		return nil, Fail("UNSUPPORTED_LOCALE", 400)
	}
	// readAll and its helpers only access Markdown and document.json. No method
	// requiring a database or a writer is called on this filesystem view.
	view := Store{root: r.root}
	all, err := view.readAll()
	if err != nil {
		return nil, err
	}
	return selectDocuments(all, locale, false), nil
}
