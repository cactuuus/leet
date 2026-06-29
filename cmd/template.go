package cmd

import (
	"fmt"
	"os"

	"github.com/cactuuus/leet/internal/language"
	"github.com/spf13/cobra"
)

func NewTemplateCmd(ctx AppContext) *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   			"template",
		Short: 			"Manage language templates.",
		SilenceUsage: 	true,
	}

	templateCmd.AddCommand(NewTemplateMakeCmd(ctx))
	templateCmd.AddCommand(NewTemplateOpenCmd(ctx))
	templateCmd.AddCommand(NewTemplateDeleteCmd(ctx))
	templateCmd.AddCommand(NewTemplateListCmd(ctx))

	return templateCmd
}

func NewTemplateMakeCmd(ctx AppContext) *cobra.Command {
	templateMakeCmd := &cobra.Command{
		Use:   			"make <lang>",
		Short: 			"Create a custom template for a language.",
		Long:  			"Create a custom template for a language.\nThis creates a new custom template for the specified language, based on the language's default template.",
		Args:  			cobra.ExactArgs(1),
		SilenceUsage: 	true,
		RunE: 			func(cmd *cobra.Command, args []string) error {
			// validate flags
			open, err := cmd.Flags().GetBool("open")
			if err != nil {
				return fmt.Errorf("failed to parse flags: %w", err)
			}

			// get and validate language
			l, ok := language.Get(args[0])
			if !ok {
				return fmt.Errorf("unknown language: %q", args[0])
			}

			s := ctx.Scaffolder()

			// check if a custom template already exists, and if so confirm overwrite
			exists, err := s.TemplateExists(l)
			if err != nil {
				return err
			}
			if exists {
				confirm, err := promptYesNo(fmt.Sprintf("A custom template for %s already exists. Overwrite?", l.Name))
				if err != nil {
					return err
				}
				if !confirm {
					fmt.Println("Aborted.")
					return nil
				}
			}

			// create the custom template
			fmt.Print("Create custom template... ")
			if err := s.WriteCustomTemplate(l); err != nil {
				return fmt.Errorf("failed to create custom template for %s: %w", l.Name, err)
			}
			fmt.Print("✓\n")

			if open {
				fmt.Print("Opening in editor... ")
				if err := openInEditor(ctx.Config().Editor, s.GetTemplatePath(l)); err != nil {
					return fmt.Errorf("failed to open directory in editor: %w", err)
				}
				fmt.Print("✓\n")
			}

			return nil
		},
	}

	templateMakeCmd.Flags().BoolP("open", "o", false, "Open the custom template in your editor after creating it.")
	return templateMakeCmd
}

func NewTemplateOpenCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   			"open [lang]",
		Short: 			"Open the template file, or the templates directory, in your editor.",
		Long: 			"Open the template file for a given language, or the templates directory, in your editor.\nIf no language is specified, the templates directory is opened.",
		Args:  			cobra.MaximumNArgs(1),
		SilenceUsage: 	true,
		RunE: 			func(cmd *cobra.Command, args []string) error {
			cfg := ctx.Config()

			var dirToOpen string
			if len(args) == 0 {
				dirToOpen = cfg.TemplatesDir
			} else {
				// get and validate language
				l, ok := language.Get(args[0])
				if !ok {
					return fmt.Errorf("unknown language: %q", args[0])
				}
				s := ctx.Scaffolder()
				exists, err := s.TemplateExists(l)
				if err != nil {
					return err
				}
				if exists {
					dirToOpen = s.GetTemplatePath(l)
				} else {
					return fmt.Errorf("no custom template found for language %s. Use 'leet template make %s' to create one.", l.Name, l.Slug)
				}
			}

			fmt.Print("Opening in editor... ")
			if err := openInEditor(ctx.Config().Editor, dirToOpen); err != nil {
				return fmt.Errorf("failed to open directory in editor: %w", err)
			}
			fmt.Print("✓\n")
			return nil
		},
	}
}

func NewTemplateDeleteCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   			"delete <lang>",
		Short: 			"Delete a custom template for a language.",
		Args:  			cobra.ExactArgs(1),
		SilenceUsage: 	true,
		RunE: 			func(cmd *cobra.Command, args []string) error {
			// get and validate language
			l, ok := language.Get(args[0])
			if !ok {
				return fmt.Errorf("unknown language: %q", args[0])
			}

			s := ctx.Scaffolder()

			exists, err := s.TemplateExists(l)
			if err != nil {
				return err
			}
			if !exists {
				fmt.Printf("No custom template found for language %s, aborting.", l.Name)
				return nil
			}
			err = os.Remove(s.GetTemplatePath(l))
			if err != nil {
				return fmt.Errorf("failed to delete template for %s: %w", l.Name, err)
			}
			return nil
		},
	}
}

func NewTemplateListCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   			"list",
		Short: 			"List all custom templates.",
		SilenceUsage: 	true,
		RunE: 			func(cmd *cobra.Command, args []string) error {
			s := ctx.Scaffolder()
			fmt.Println("Custom templates:")
			for _, l := range language.All() {
				exists, err := s.TemplateExists(l)
				if err != nil {
					return err
				}
				if exists {
					fmt.Printf("- %s (%s)\n", l.Name, l.Slug)
				}
			}
			return nil
		},
	}
}
