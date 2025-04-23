package cmd

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed templates/usecase/*.tmpl
var usecaseTemplates embed.FS

func loadUsecaseTemplates() (*template.Template, error) {
	sub, err := fs.Sub(usecaseTemplates, "templates/usecase")
	if err != nil {
		return nil, err
	}
	// initializing all templates (*.tmpl)
	return template.ParseFS(sub, "*.tmpl")
}

var usecaseCmd = &cobra.Command{
	Use:   "usecase [module] [name]",
	Short: "Generate a usecase scaffold",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		module, name := args[0], args[1]

		// مسیر مقصد: internal/<module>/usecase/<name>
		destDir := filepath.Join("internal", module, "usecase", name)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}

		tpl, err := loadUsecaseTemplates()
		if err != nil {
			return err
		}

		files := []struct {
			tmplName string
			outName  string
		}{
			{"inport.tmpl", "inport.go"},
			{"outport.tmpl", "outport.go"},
			{"interactor.tmpl", "interactor.go"},
		}

		for _, f := range files {
			outPath := filepath.Join(destDir, f.outName)
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			data := struct {
				Package string
				Module  string
			}{
				Package: name,
				Module:  module,
			}

			if err := tpl.ExecuteTemplate(outFile, f.tmplName, data); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(usecaseCmd)
}
