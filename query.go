package pacman

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type Repository string

const (
	SyncRoot  = "/var/lib/pacman/sync"
	LocalRoot = "/var/lib/pacman/local"

	RepositoryCore     Repository = "core"
	RepositoryExtra    Repository = "extra"
	RepositoryMultilib Repository = "multilib"
)

var packageCache map[string]*Package = make(map[string]*Package)

// QueryPackageAll queries a specific package database for the package `name`. This may be faster if you know which repository the package exists in.
func QueryPackageDatabase(name string, db Repository) (*Package, error) {
	if pkg, ok := packageCache[name]; ok {
		return pkg, nil
	}
	f, err := os.Open(databasePathFromRepository(db))
	if err != nil {
		return nil, err
	}
	g, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	t := tar.NewReader(g)
	for {
		hdr, err := t.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(hdr.Name, "/desc") {
			continue
		}
		if name != parsePackageName(hdr.Name) {
			continue
		}
		pkg := &Package{}
		b, err := io.ReadAll(t)
		if err != nil {
			return nil, err
		}
		if err := pkg.UnmarshalText(b); err != nil {
			return nil, err
		}
		if pkg.Name != "" {
			packageCache[pkg.Name] = pkg
		}
		if pkg.Name == name {
			return pkg, nil
		}
	}
	return nil, nil
}

// QueryLocal queries a specific package from the set of installed packages. That is, if `name` is not installed, nil is returned even if the package may exist in the local databases.
func QueryLocal(name string) (*Package, error) {
	if pkg, ok := packageCache[name]; ok {
		return pkg, nil
	}
	dirs, err := os.ReadDir(LocalRoot)
	if err != nil {
		return nil, err
	}
	for _, d := range dirs {
		if name == parsePackageName(d.Name()) {
			pkg := &Package{}
			f, err := os.OpenInRoot(path.Join(LocalRoot, d.Name()), "desc")
			if err != nil {
				return nil, err
			}
			b, err := io.ReadAll(f)
			if err != nil {
				return nil, err
			}
			if err := pkg.UnmarshalText(b); err != nil {
				return nil, err
			}
			if pkg.Name != "" {
				packageCache[pkg.Name] = pkg
			}
			if pkg.Name == name {
				return pkg, nil
			}
		}
	}
	return nil, fmt.Errorf("could not find package: %s", name)
}

// QueryPackageAllDatabases queries all package databases for the package `name`.
func QueryPackageAllDatabases(name string) (*Package, error) {
	var err error
	var pkg *Package
	for _, db := range []Repository{RepositoryCore, RepositoryExtra, RepositoryMultilib} {
		pkg, err = QueryPackageDatabase(name, db)
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, err
		}
		// Package was found + no error
		if pkg != nil {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("could not find package: %s", name)
}

func databasePathFromRepository(db Repository) string {
	return path.Join(SyncRoot, string(db)+".db")
}

func parsePackageName(s string) string {
	// Strip trailing version info (everything after second-to-last '-').
	idx := strings.LastIndex(s, "-")
	if idx == -1 {
		return s
	}
	s = s[:idx]
	idx = strings.LastIndex(s, "-")
	if idx == -1 {
		return s
	}
	return s[:idx]
}
