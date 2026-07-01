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
		Use:	"make <lang>",
		Short: 	"Create a custom template for a language.",
		Long:  	"Create a custom template for a language.\nThis creates a new custom template " +
		 		"for the specified language, based on the language's default template.",
		Args:  	cobra.ExactArgs(1),
		RunE:	func(cmd *cobra.Command, args []string) error {
			// validate flags
			open, err := cmd.Flags().GetBool("open")
			if err != nil {
				return fmt.Errorf("Failed to parse --open flag:\n%w", err)
			}

			// get and validate language
			l, err := parseLanguage(args[1])
			if err != nil {
				return err
			}

			s, err := ctx.Scaffolder()
			if err != nil {
				return err
			}

			// check if a custom template already exists, and if so confirm overwrite
			exists, err := s.TemplateExists(l)
			if err != nil {
				return err
			}
			if exists {
				msg := fmt.Sprintf("A custom template for %s already exists. Overwrite?", l.Name)
				confirm, err := promptYesNo(msg)
				if err != nil {
					return err
				}
				if !confirm {
					fmt.Println("Aborted.")
					return nil
				}
			}

			// create the custom template
			printActionStart("Create custom template")
			if err := s.WriteCustomTemplate(l); err != nil {
				return fmt.Errorf("Failed to create custom template for %s:\n%w", l.Name, err)
			}
			printActionSuccess()


			if open {
				return openInEditor(ctx.Config(), s.GetTemplatePath(l))
			}
			return nil
		},
	}

	templateMakeCmd.Flags().BoolP(
		"open", "o", false,
		"Open the custom template in your editor after creating it.",
	)
	return templateMakeCmd
}

func NewTemplateOpenCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:	"open [lang]",
		Short: 	"Open the template file, or the templates directory, in your editor.",
		Long: 	"Open the template file for a given language in your editor, or the templates " +
				"directory if no language is specified",
		Args:  	cobra.MaximumNArgs(1),
		RunE:	func(cmd *cobra.Command, args []string) error {
			cfg := ctx.Config()

			dirToOpen := cfg.TemplatesDir
			if len(args) > 0 {
				// get and validate language
				l, err := parseLanguage(args[1])
					if err != nil {
						return err
					}

				s, err := ctx.Scaffolder()
				if err != nil {
					return err
				}

				exists, err := s.TemplateExists(l)
				if err != nil {
					return err
				}
				if exists {
					dirToOpen = s.GetTemplatePath(l)
				} else {
					return fmt.Errorf(
						"No custom template found for language %s. " +
						"Use 'leet template make %s' to create one.",
						l.Name, l.Slug)
				}
			}

			if err := openInEditor(ctx.Config(), dirToOpen); err != nil {
				return fmt.Errorf("failed to open directory in editor: %w", err)
			}
			return nil
		},
	}
}

func NewTemplateDeleteCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   	"delete <lang>",
		Short: 	"Delete a custom template for a language.",
		Args:  	cobra.ExactArgs(1),
		RunE:	func(cmd *cobra.Command, args []string) error {
			// get and validate language
			l, err := parseLanguage(args[1])
			if err != nil {
				return err
			}

			s, err := ctx.Scaffolder()
			if err != nil {
				return err
			}

			if exists, err := s.TemplateExists(l); err != nil {
				return err
			} else if !exists {
				fmt.Printf("No custom template found for language %s, aborting.", l.Name)
				return nil
			}
			if err := os.Remove(s.GetTemplatePath(l)); err != nil {
				return fmt.Errorf("Failed to delete template for %s:\n%w", l.Name, err)
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
			s, err := ctx.Scaffolder()
			if err != nil {
				return err
			}

			// gather languages that have custom templates
			var customTemplates []language.Language
			for _, l := range language.All() {
				if exists, err := s.TemplateExists(l); err != nil {
					return err
				} else if exists {
					customTemplates = append(customTemplates, l)
				}
			}

			if len(customTemplates) == 0 {
				fmt.Println("No custom templates found.")
				return nil
			}

			fmt.Println("CUSTOM TEMPLATES")
			// print the list of custom templates
			for _, t := range customTemplates {
				fmt.Printf("- %s (%s)\n", t.Name, t.Slug)
			}
			return nil
		},
	}
}
