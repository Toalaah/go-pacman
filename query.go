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

const SyncRoot = "/var/lib/pacman/sync"

var databases []string = []string{
	path.Join(SyncRoot, "core.db"),
	path.Join(SyncRoot, "extra.db"),
	path.Join(SyncRoot, "multilib.db"),
}

var packageCache map[string]*Package = make(map[string]*Package)

func QueryPackageDatabase(name, db string) (*Package, error) {
	if pkg, ok := packageCache[name]; ok {
		return pkg, nil
	}
	f, err := os.Open(db)
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
		// We *may* have found a matching package, parse + cache it.
		if strings.HasPrefix(hdr.Name, name) {
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
	}
	return nil, nil
}

func QueryPackage(name string) (*Package, error) {
	var err error
	var pkg *Package
	for _, db := range databases {
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
