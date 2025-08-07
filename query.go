package pacman

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path"
	"strings"
)

const SyncRoot = "/var/lib/pacman/sync"

var dataBases []string = []string{
	path.Join(SyncRoot, "core.db"),
	path.Join(SyncRoot, "extra.db"),
	path.Join(SyncRoot, "multilib.db"),
}

var packageCache map[string]*Package = make(map[string]*Package)

func QueryPackageDatabase(packageName, db string) (*Package, error) {
	if pkg, ok := packageCache[packageName]; ok {
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
		if strings.HasPrefix(hdr.Name, packageName) {
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
			if pkg.Name == packageName {
				return pkg, nil
			}
		}
	}
	return nil, nil
}

func QueryPackage(packageName string) (*Package, error) {
	var err error
	var pkg *Package
	for _, db := range dataBases {
		pkg, err = QueryPackageDatabase(packageName, db)
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
	return nil, err
}
