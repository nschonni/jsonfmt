package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alecthomas/kingpin"
	zglob "github.com/mattn/go-zglob"
	"github.com/pmezard/go-difflib/difflib"
)

var (
	version = "master"
	commit  = "none"
	date    = "unknown"

	app   = kingpin.New("jsonfmt", "Like gofmt, but for JSON")
	write = app.Flag("write", "write changes to the files").Short('w').Bool()
	globs = app.Arg("files", "glob of the files you want to check").Default("**/*.json").Strings()
)

func main() {
	app.Version(fmt.Sprintf("%v, commit %v, built at %v", version, commit, date))
	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')

	kingpin.MustParse(app.Parse(os.Args[1:]))

	var failed bool
	for _, glob := range *globs {
		matches, err := zglob.Glob(glob)
		app.FatalIfError(err, "failed to parse glob: %s", glob)

		for _, match := range matches {
			bts, err := ioutil.ReadFile(match)
			app.FatalIfError(err, "failed to read file: %s", match)
			var out bytes.Buffer
			err = json.Indent(&out, bytes.TrimSpace(bts), "", "  ") // TODO: support to customize indent
			app.FatalIfError(err, "failed to format json file: %s", match)
			out.Write([]byte{'\n'})
			if bytes.Equal(bts, out.Bytes()) {
				continue
			}
			if *write {
				err := ioutil.WriteFile(match, out.Bytes(), 0)
				app.FatalIfError(err, "failed to write json file: %s", match)
				continue
			}

			diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(bts)),
				B:        difflib.SplitLines(string(out.Bytes())),
				FromFile: "Original",
				ToFile:   "Formatted",
				Context:  3,
			})
			app.FatalIfError(err, "failed to diff file: %s", match)
			app.Errorf("file %s differs:\n%s\n", match, diff)
			failed = true
		}
	}
	if failed {
		app.Fatalf("some files are not properly formated, check above")
	}
}
