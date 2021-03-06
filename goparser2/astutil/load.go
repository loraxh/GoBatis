package astutil

import (
	"os"
	"path/filepath"
	"strings"
)

func Load(ctx *Context, currentDir, pkgName string) (*File, error) {
	dir, err := searchDir(ctx, currentDir, pkgName)
	if err != nil {
		return nil, err
	}
}

func dirExists(name string) bool {
	st, err := os.Stat(name)
	if err != nil {
		return false
	}
	return st.IsDir()
}

// hasSubdir reports if dir is within root by performing lexical analysis only.
func hasSubdir(root, dir string) (rel string, ok bool) {
	const sep = string(filepath.Separator)
	root = filepath.Clean(root)
	if !strings.HasSuffix(root, sep) {
		root += sep
	}
	dir = filepath.Clean(dir)
	if !strings.HasPrefix(dir, root) {
		return "", false
	}
	return filepath.ToSlash(dir[len(root):]), true
}

func searchDir(ctx *Context, currentDir, pkgName string) (string, error) {
	searchVendor := func(root string, isGoroot bool) bool {
		sub, ok := hasSubdir(root, currentDir)
		if !ok || !strings.HasPrefix(sub, "src/") {
			return false, ""
		}

		for {
			vendor := filepath.Join(root, sub, "vendor")
			if dirExists(vendor) {
				dir := filepath.Join(vendor, pkgName)
				if dirExists(dir) {
					return true, dir
				}
			}
			i := strings.LastIndex(sub, "/")
			if i < 0 {
				break
			}
			sub = sub[:i]
		}
		return false, ""
	}

	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		for _, root := range filepath.SplitPathList(gopath) {
			if root == "" {
				continue
			}
			if strings.HasPrefix(root, "~") {
				continue
			}

			pkgDir := filepath.Join(root, "src", pkgName)
			if dirExists(pkgDir) {
				return pkgDir, nil
			}

			if ok, pkgDir := searchVendor(root, false); ok {
				return pkgDir, nil
			}
		}
	}

	// If modules are not enabled, then the in-process code works fine and we should keep using it.
	switch os.Getenv("GO111MODULE") {
	case "off":
		return "", nil
	default: // "", "on", "auto", anything else
		// Maybe use modules.
	}

	// Look to see if there is a go.mod.
	// Since go1.13, it doesn't matter if we're inside GOPATH.
	parent := currentDir
	for {
		info, err := os.Stat(filepath.Join(parent, "go.mod"))
		if err == nil && !info.IsDir() {
			break
		}
		d := filepath.Dir(parent)
		if len(d) >= len(parent) {
			return ""
		}
		parent = d
	}

	filepath.Join(parent, "go.mod")

	return "", nil
}
