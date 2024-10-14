package pp

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"go.llib.dev/testcase/internal/env"
)

func Test_envPP(t *testing.T) {
	defer initDefaultWriter()

	t.Run("when not supplied", func(t *testing.T) {
		env.UnsetEnv(t, "PP")
		buf := stubDefaultWriter(t)
		initDefaultWriter()
		PP("OK")
		assertNotEmpty(t, buf.Bytes())
	})

	t.Run("when provided but empty", func(t *testing.T) {
		env.SetEnv(t, "PP", "")
		buf := stubDefaultWriter(t)
		initDefaultWriter()
		PP("OK")
		assertNotEmpty(t, buf.Bytes())
	})

	t.Run("when provided but not a valid file path", func(t *testing.T) {
		env.SetEnv(t, "PP", ".")
		buf := stubDefaultWriter(t)
		initDefaultWriter()
		PP("OK")
		assertNotEmpty(t, buf.Bytes())
	})

	t.Run("when existing file provided", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "")
		assertNoError(t, err)
		env.SetEnv(t, "PP", f.Name())
		buf := stubDefaultWriter(t)
		initDefaultWriter()
		PP("OK")
		assertEmpty(t, buf.Bytes())
		bs, err := io.ReadAll(f)
		assertNoError(t, err)
		assertNotEmpty(t, bs)
	})

	t.Run("when non existing file provided", func(t *testing.T) {
		fpath := filepath.Join(t.TempDir(), "test.txt")
		env.SetEnv(t, "PP", fpath)
		buf := stubDefaultWriter(t)
		initDefaultWriter()
		PP("OK")
		assertEmpty(t, buf.Bytes())
		bs, err := os.ReadFile(fpath)
		assertNoError(t, err)
		assertNotEmpty(t, bs)
	})

}
