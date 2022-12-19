package predictor

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/posener/complete"
	"github.com/sirkon/jsonexec"
	"github.com/sirkon/ttgenlib/internal/ordmap"
)

type PackagePath struct{}

// Predict to satisfy complete.Predictor.
func (p PackagePath) Predict(args complete.Args) []string {
	var dst struct {
		Path string
		Dir  string
	}
	if err := jsonexec.Run(&dst, "go", "list", "-m", "--json"); err != nil {
		return nil
	}

	dir, basePrefix := filepath.Split(args.Last)
	if dir == "" {
		dir = "./"
	}

	reldirs := ordmap.New[string, bool]()
	err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if info == nil {
			return nil
		}

		if info.IsDir() {
			switch path {
			case ".git", ".idea":
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		d, _ := filepath.Split(path)
		d = strings.TrimRight(d, string(os.PathSeparator))
		if reldirs.Has(d) {
			return nil
		}

		_, db := filepath.Split(d)
		if !strings.HasPrefix(db, basePrefix) {
			return nil
		}

		reldirs.Set(d, true)
		return nil
	})
	if err != nil {
		return nil
	}

	return reldirs.Keys()
}

var _ complete.Predictor = PackagePath{}
