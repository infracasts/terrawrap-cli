// All modifications to original contents, hereto known as the Derivative Works,
// are copyright 2022 InfraCasts, LLC, and licensed under the Apache License, Version 2.0.
//
// Copyright © 2021 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"fmt"
	"github.com/infracasts/terrawrap-cli/terraform"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

var (
	varPrefix, attrPrefix, outputPath, author, license, resourceName string
	standAlone, disableVarPrefix, disableAttrPrefix                  bool
)

func init() {
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "output/terraform_modules", "Output path")
	generateCmd.Flags().StringVarP(&author, "author", "a", "YOUR NAME <email@example.com>", "author name for copyright attribution")
	generateCmd.Flags().StringVarP(&license, "license", "l", "apache", "license to set for the generated code")
	generateCmd.Flags().StringVarP(&resourceName, "resource-name", "n", "default", `local name of the resource generated (e.g. "default" in `+"`"+`resource "aws_secretsmanager_secret" "default"`+"`"+`)`)

	generateCmd.Flags().BoolVarP(&standAlone, "stand-alone", "s", false, "modules should be created in their own directory named for the resource type (e.g. '$OUTPUT_PATH/aws_secretsmanager_secret/*.tf')")

	// Variable/Output prefix disable
	generateCmd.Flags().BoolVarP(&disableVarPrefix, "no-var-prefix", "", false, "disable naming prefix for variables")
	generateCmd.Flags().BoolVarP(&disableAttrPrefix, "no-out-prefix", "", false, "disable naming prefix for outputs")

	generateCmd.Flags().StringVarP(&varPrefix, "variable-prefix", "v", "", "variable prefix (default: <resource_type>_<resource_name>_<variable>)")
	generateCmd.Flags().StringVarP(&attrPrefix, "output-prefix", "p", "", "output prefix (default: <resource_type>_<resource_name>_<attribute>)")

	cobra.CheckErr(viper.BindPFlag("author", generateCmd.Flags().Lookup("author")))
	cobra.CheckErr(viper.BindPFlag("output", generateCmd.Flags().Lookup("output")))
	cobra.CheckErr(viper.BindPFlag("license", generateCmd.Flags().Lookup("license")))
	cobra.CheckErr(viper.BindPFlag("stand-alone", generateCmd.Flags().Lookup("stand-alone")))

	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate [terraform_resource_type]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Generate a Terraform Module from the target resource",
	Long: `Terrawrap will generate a terraform module given the resource name 
provided (e.g. aws_secretsmanager_secret).`,
	Run: func(cmd *cobra.Command, args []string) {
		resourceType := args[0]
		// Set Variable/Attribute prefixes
		if !disableVarPrefix && len(varPrefix) == 0 {
			varPrefix = setVariablePrefix(resourceType, resourceName)
		}
		if !disableAttrPrefix && len(attrPrefix) == 0 {
			attrPrefix = setAttrPrefix(resourceType, resourceName)
		}

		// initialize/download relevant docs
		tfProvider, err := fetchDocProvider(resourceType)
		cobra.CheckErr(err)

		// Initialize modules directory / output path
		module, err := initializeModulesBase(resourceType, resourceName)
		cobra.CheckErr(err)

		// Parse resource from markdown
		resource, err := parseResource(resourceType, resourceName, varPrefix, attrPrefix, module, tfProvider)
		cobra.CheckErr(err)

		cobra.CheckErr(generateResource(resource))

		fmt.Printf("Your new module is ready at\n%s\n", module.AbsolutePath)
	},
}

func setVariablePrefix(resourceType, resourceName string) string {
	return strings.Join([]string{resourceType, resourceName}, "_")
}

func setAttrPrefix(resourceType, resourceName string) string {
	return strings.Join([]string{resourceType, resourceName}, "_")
}

func fetchDocProvider(resourceType string) (*terraform.Provider, error) {
	var (
		err                           error
		providerName, providerDocPath string
		tfProvider                    *terraform.Provider
	)

	providerTypeFormat := regexp.MustCompile(`(?P<Provider>[A-Za-z0-9]+)_`)
	subMatches := providerTypeFormat.FindStringSubmatch(resourceType)
	if len(subMatches) < 1 {
		return tfProvider, fmt.Errorf("failed to identify provider from resource type. " +
			"ensure the resource type matches the provider format " +
			"(e.g. aws_secretsmanager_secret)")
	}

	providerName = subMatches[providerTypeFormat.SubexpIndex("Provider")]
	if providerName == "" {
		return tfProvider, fmt.Errorf("failed to identify provider from resource type. " +
			"ensure the resource type matches the provider format " +
			"(e.g. aws_secretsmanager_secret)")
	}

	tfProvider, err = terraform.GetProvider(providerName)
	if err != nil {
		return tfProvider, fmt.Errorf("%w", err)
	}

	// via terrawrap root command initialization
	providerDocPath = path.Join(cfgPath, "provider_docs", tfProvider.Name, tfProvider.Version)
	tfProvider.SetRootDocPath(providerDocPath)
	if _, err = os.Stat(providerDocPath); os.IsNotExist(err) {
		// create directory
		if err = os.MkdirAll(providerDocPath, 0754); err != nil {
			return tfProvider, err
		}
	} else {
		var files []os.DirEntry

		files, err = os.ReadDir(providerDocPath)
		if err != nil {
			return tfProvider, fmt.Errorf("couldn't read provider docs directory %s: %w", providerDocPath, err)
		}

		// Directory exist, files exist; return
		if len(files) != 0 {
			log.Printf("found non-empty provider docs directory %s, skipping download...", providerDocPath)
			return tfProvider, nil
		}
	}

	if err = tfProvider.DownloadDocs(); err != nil {
		return tfProvider, fmt.Errorf("unable to download docs: %w", err)
	}

	return tfProvider, err
}

// this could easily be extended to generate multiple
func generateResource(resource terraform.TFResource) error {
	if err := resource.Create(); err != nil {
		return fmt.Errorf("failed to create module: %w", err)
	}

	return nil
}

func initializeModulesBase(resourceType, resourceName string) (*terraform.Module, error) {
	module := &terraform.Module{
		Copyright:     copyrightLine(), // TODO: allow override
		TerrawrapLine: terrawrapLine(resourceType, resourceName),
	}

	wd, err := os.Getwd()
	if err != nil {
		return module, fmt.Errorf("failed to find current directory: %w", err)
	}

	if outputPath != "." {
		wd = fmt.Sprintf("%s/%s", wd, outputPath)
	}

	module.AbsolutePath = wd

	if err := module.Initialize(); err != nil {
		return module, fmt.Errorf("failed to initialize module directory: %w", err)
	}

	return module, nil
}

// docerizeResourceType parses and strips the resource type of any prefix or suffix
// in order to match the filename format used by Hashicorp in their docs.
func docerizeResourceType(resourceType string) string {
	re := regexp.MustCompile(`^aws_`)
	return re.ReplaceAllString(resourceType, "")
}

// TODO: allow generating multiple types of things if needed
func parseResource(resourceType, resourceName, varPrefix, attrPrefix string, module *terraform.Module, tfProvider *terraform.Provider) (terraform.TFResource, error) {
	var (
		err    error
		source []byte
	)

	docerizedResourceType := docerizeResourceType(resourceType)
	resource := terraform.NewTFResource()
	resource.Module = module
	resource.Name = resourceName
	resource.VarPrefix = varPrefix
	resource.AttrPrefix = attrPrefix
	resource.Type = resourceType
	resource.DocerizedType = docerizedResourceType
	resource.DocPath = path.Join(tfProvider.DocPath(), fmt.Sprintf("r/%s.html.markdown", docerizedResourceType))
	resource.AbsolutePath = module.AbsolutePath
	if standAlone {
		resource.AbsolutePath = resource.AbsolutePath + fmt.Sprintf("/%s", resourceType)
	}

	err = resource.SetHashicorpResource()
	if err != nil {
		return resource, fmt.Errorf("failed to set hashicorp resource: %w", err)
	}

	source, err = os.ReadFile(resource.DocPath)
	if err != nil {
		return resource, fmt.Errorf("failed to read doc file %s: %w", resource.DocPath, err)
	}

	md := goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	doc := md.Parser().Parse(text.NewReader(source))

	err = ast.Walk(doc, WalkerFn(source, &resource))
	if err != nil {
		return resource, fmt.Errorf("failed to walk document tree: %w", err)
	}

	return resource, nil
}

func WalkerFn(source []byte, resource *terraform.TFResource) ast.Walker {
	var (
		h1Found                    bool
		currentSection, subSection string
	)

	SectionArgumentReference := regexp.MustCompile("argument[s]?[-]+reference")
	SectionAttributesReference := regexp.MustCompile("attribute[s]?[-]+reference")
	EntryFormat := regexp.MustCompile(`(?P<Name>[-_A-Za-z0-9]+) - \(?(?P<Optional>Optional|Required)?(?:[, ]+)?(?P<Deprecated>DEPRECATED)?\)?(?:[ ]*)(?P<Description>.*$)`)

	return func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if !h1Found && node.Kind() != ast.KindHeading {
			return ast.WalkContinue, nil
		}

		if node.Kind() == ast.KindHeading {
			var sectionId string
			heading := node.(*ast.Heading)

			if heading.Level != 1 && !h1Found {
				return ast.WalkContinue, nil
			} else if heading.Level == 1 && !h1Found {
				h1Found = true
			}

			// Iterate over the attributes of this node, looking for the "id"
			for _, attr := range node.Attributes() {
				if bytes.Equal(attr.Name, []byte("id")) {
					sectionId = string(attr.Value.([]byte))
				}
			}

			// We really only care if we're in the Arguments or Attributes
			// sections
			if SectionArgumentReference.MatchString(sectionId) || SectionAttributesReference.MatchString(sectionId) {
				currentSection = sectionId
				subSection = ""
			} else {
				subSection = sectionId
			}

			return ast.WalkContinue, nil
		}

		// TODO: don't
		// Skip all sub-sections for now
		if len(subSection) != 0 {
			return ast.WalkContinue, nil
		}

		// Skip all nodes that aren't list items
		if node.Parent().Kind() != ast.KindList || node.Kind() != ast.KindListItem {
			return ast.WalkContinue, nil
		}

		nodeText := node.Text(source)
		subMatches := EntryFormat.FindStringSubmatch(string(nodeText))
		entry := terraform.IOEntry{Resource: resource}

		if len(subMatches) > 1 {
			var deprecated, optional, ok bool
			if len(subMatches[EntryFormat.SubexpIndex("Deprecated")]) > 0 {
				deprecated = true
			}

			if len(subMatches[EntryFormat.SubexpIndex("Optional")]) > 0 {
				optional = true
			}

			entry.Name = subMatches[EntryFormat.SubexpIndex("Name")]
			entry.Optional = optional
			entry.Deprecated = deprecated
			entry.Description = subMatches[EntryFormat.SubexpIndex("Description")]
			// Per github.com/hashicorp/terraform-provider-aws/internal/helper/schema/resource.go
			// ~line 1200, ID must always be string and isn't defined in the data sources or resource attributes
			if entry.Schema, ok = resource.Schema[entry.Name]; !ok && entry.Name != "id" {
				log.Printf("[WARN] failed to discover schema for %s.%s, so it was omitted. It may be a sub-resource.", currentSection, entry.Name)
			}
		}

		// if we're currently in the argument section
		if SectionArgumentReference.MatchString(currentSection) {
			resource.AppendArgument(terraform.InputEntry(entry))
		} else {
			resource.AppendAttribute(terraform.OutputEntry(entry))
		}

		return ast.WalkContinue, nil
	}
}

func copyrightLine() string {
	author := viper.GetString("author")

	year := time.Now().Format("2006")

	return "Copyright © " + year + " " + author
}

func terrawrapLine(resourceType, resourceName string) string {
	str := fmt.Sprintf(`
AWS Resource: %s - %s
Generated with love by Terrawrap, an InfraCasts, LLC tool!
https://infracasts.com

The code generated below was generated using MPL v2.0 licensed code 
and documentation, and as such is is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this file, You
can obtain one at https://mozilla.org/MPL/2.0/.
`, resourceType, resourceName)
	str = strings.TrimSpace(str)
	return str
}
