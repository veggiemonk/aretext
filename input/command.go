package input

import (
	"log"
	"os"

	"github.com/aretext/aretext/exec"
	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/syntax"
	"github.com/pkg/errors"
)

func commandMenuItems(config Config) func() []menu.Item {
	return func() []menu.Item {
		return []menu.Item{
			{
				Name:   "quit",
				Action: exec.NewAbortIfUnsavedChangesMutator(exec.NewQuitMutator(), true),
			},
			{
				Name:   "force quit",
				Action: exec.NewQuitMutator(),
			},
			{
				Name:   "save",
				Action: exec.NewSaveDocumentMutator(false),
			},
			{
				Name:   "force save",
				Action: exec.NewSaveDocumentMutator(true),
			},
			{
				Name:   "force reload",
				Action: exec.NewReloadDocumentMutator(true),
			},
			{
				Name:   "find and open",
				Action: exec.NewAbortIfUnsavedChangesMutator(ShowFileMenuMutator(config), true),
			},
			{
				Name:   "set syntax json",
				Action: exec.NewSetSyntaxMutator(syntax.LanguageJson),
			},
			{
				Name:   "set syntax go",
				Action: exec.NewSetSyntaxMutator(syntax.LanguageGo),
			},
			{
				Name:   "set syntax none",
				Action: exec.NewSetSyntaxMutator(syntax.LanguageUndefined),
			},
		}
	}
}

func ShowFileMenuMutator(config Config) exec.Mutator {
	findFileMenuItems := func() []menu.Item {
		dir, err := os.Getwd()
		if err != nil {
			log.Printf("Error loading menu items: %v\n", errors.Wrapf(err, "os.GetCwd"))
			return nil
		}

		items := make([]menu.Item, 0, 0)
		file.Walk(dir, config.DirNamesToHide, func(path string) {
			items = append(items, menu.Item{
				Name:   file.RelativePathCwd(path),
				Action: exec.NewLoadDocumentMutator(path, true, true),
			})
		})

		return items
	}

	return exec.NewShowMenuMutator("file path", findFileMenuItems, true, false)
}
