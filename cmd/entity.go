package cmd

import (
	"embed"
	"fmt"
	"github.com/a-aslani/wotop/util"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// Embed the templates for entity generation.
//
//go:embed templates/entity/*.tmpl
var entityFS embed.FS

// loadEntityTemplates loads and parses the embedded entity templates.
// It uses the `fs.Sub` function to access the "templates/entity" subdirectory
// and parses all files with the `.tmpl` extension.
func loadEntityTemplates() (*template.Template, error) {
	sub, err := fs.Sub(entityFS, "templates/entity")
	if err != nil {
		return nil, err
	}
	return template.ParseFS(sub, "*.tmpl")
}

// toCamelCase converts a string to CamelCase format.
// For example, "user_state" or "userState" will be converted to "UserState".
// It splits the input string by underscores (`_`) or hyphens (`-`) and capitalizes each part.
func toCamelCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})
	for i, p := range parts {
		parts[i] = strings.Title(p)
	}
	return strings.Join(parts, "")
}

// entityCmd defines a Cobra command for generating an entity scaffold.
// Usage: `entity [domain] [name]`
// - `domain`: The domain name where the entity will be created.
// - `name`: The name of the entity to generate.
var entityCmd = &cobra.Command{
	Use:   "entity [domain] [name]",
	Short: "Generate an entity scaffold",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Extract domain and raw entity name from arguments.
		domain, rawName := args[0], args[1]
		entityName := toCamelCase(rawName)

		// Define the destination directory: internal/<domain>/model/entity
		destDir := filepath.Join("internal", domain, "model", "entity")
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}

		// Load the entity templates.
		tpl, err := loadEntityTemplates()
		if err != nil {
			return err
		}

		// Generate the file name in snake_case format.
		// Example: "userState" → "user_state.go"
		fileName := util.SnakeCase(rawName) + ".go"
		filePath := filepath.Join(destDir, fileName)

		// Create the output file.
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		// Define the data to pass to the template.
		data := struct {
			Entity string
		}{Entity: entityName}

		// Execute the template and write the output to the file.
		if err := tpl.ExecuteTemplate(f, "entity.tmpl", data); err != nil {
			return err
		}

		// Print a success message with the generated file path.
		fmt.Printf("✅ Generated entity at %s\n", filePath)
		return nil
	},
}

// init adds the `entityCmd` to the root command.
func init() {
	rootCmd.AddCommand(entityCmd)
}
