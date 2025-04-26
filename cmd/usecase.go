package cmd

import (
	"embed"
	"fmt"
	"github.com/a-aslani/wotop/util"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

// usecaseTemplates embeds all template files located in the "templates/usecase" directory.
//
//go:embed templates/usecase/*.t
//go:embed templates/usecase/*.tmpl
var usecaseTemplates embed.FS

// loadUsecaseTemplates loads and parses all template files (*.tmpl) from the "templates/usecase" directory.
// Returns a compiled *template.Template or an error if the templates cannot be loaded.
func loadUsecaseTemplates() (*template.Template, error) {
	sub, err := fs.Sub(usecaseTemplates, "templates/usecase")
	if err != nil {
		return nil, err
	}
	// Parse all template files in the subdirectory.
	return template.ParseFS(sub, "*.tmpl")
}

// usecaseCmd defines a Cobra command for generating a usecase scaffold.
// It takes two arguments: [domain] and [name], and generates a set of files
// in the "internal/<domain>/usecase/<name>" directory.
var usecaseCmd = &cobra.Command{
	Use:   "usecase [domain] [name]",     // Command usage format
	Short: "Generate a usecase scaffold", // Short description of the command
	Args:  cobra.ExactArgs(2),            // Requires exactly two arguments
	RunE: func(cmd *cobra.Command, args []string) error {
		// Extract arguments: domain and raw name
		domain, rawName := args[0], args[1]
		// Convert the raw name to snake_case
		snakeName := util.SnakeCase(rawName)

		// Destination directory: internal/<domain>/usecase/<name>
		destDir := filepath.Join("internal", domain, "usecase", snakeName)
		// Create the destination directory and any necessary parent directories
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}

		// Load the usecase templates
		tpl, err := loadUsecaseTemplates()
		if err != nil {
			return err
		}

		// Define the list of template files and their corresponding output file names
		files := []struct {
			tmplName string // Template file name
			outName  string // Output file name
		}{
			{"inport.tmpl", "inport.go"},
			{"outport.tmpl", "outport.go"},
			{"interactor.tmpl", "interactor.go"},
		}

		// Iterate over the files, generate content from templates, and write to output files
		for _, f := range files {
			// Full path to the output file
			outPath := filepath.Join(destDir, f.outName)

			// Create the output file
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			// Data to be passed to the template
			data := struct {
				Package string // Package name (snake_case)
				Domain  string // Domain name
			}{
				Package: snakeName,
				Domain:  domain,
			}

			// Execute the template and write the output to the file
			if err := tpl.ExecuteTemplate(outFile, f.tmplName, data); err != nil {
				return err
			}

			// Print a success message for the generated file
			fmt.Printf("✅ Generated usecase at %s\n", outPath)
		}
		return nil
	},
}

// init adds the usecaseCmd to the root command.
func init() {
	rootCmd.AddCommand(usecaseCmd)
}
