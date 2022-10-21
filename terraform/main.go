package terraform

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/infracasts/terraform-provider-aws-expose-internal/provider"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/infracasts/terrawrap-cli/tpl"
)

type AttributeList struct {
	Attributes []OutputEntry
	keys       map[string]struct{}
}

func NewAttributeList() *AttributeList {
	return &AttributeList{
		Attributes: make([]OutputEntry, 0),
		keys:       make(map[string]struct{}),
	}
}

func (l *AttributeList) Append(item OutputEntry) {
	if _, ok := l.keys[item.Name]; ok {
		return
	}

	l.Attributes = append(l.Attributes, item)
	l.keys[item.Name] = struct{}{}
}

type ArgumentList struct {
	Arguments []InputEntry
	keys      map[string]struct{}
}

func NewArgumentList() *ArgumentList {
	return &ArgumentList{
		Arguments: make([]InputEntry, 0),
		keys:      make(map[string]struct{}),
	}
}

func (l *ArgumentList) Append(item InputEntry) {
	if _, ok := l.keys[item.Name]; ok {
		return
	}

	l.Arguments = append(l.Arguments, item)
	l.keys[item.Name] = struct{}{}
}

type TFResource struct {
	*Module
	Name          string
	VarPrefix     string
	AttrPrefix    string
	Type          string
	DocerizedType string
	DocPath       string
	AbsolutePath  string
	*schema.Resource
	arguments  *ArgumentList
	attributes *AttributeList
}

func NewTFResource() TFResource {
	return TFResource{
		arguments:  NewArgumentList(),
		attributes: NewAttributeList(),
	}
}

func (r *TFResource) SetHashicorpResource() error {
	var (
		hcProvider *schema.Provider
		ok         bool
		err        error
	)

	hcProvider, err = provider.New(context.Background())
	if err != nil {
		return fmt.Errorf("failed to initialize hashicorp terraform provider: %w", err)
	}

	r.Resource, ok = hcProvider.ResourcesMap[r.Type]
	if !ok {
		return fmt.Errorf("failed to discover resource of type %s in resource map", r.Name)
	}

	return err
}

func (r *TFResource) AppendArgument(entry InputEntry) {
	r.arguments.Append(entry)
}

func (r *TFResource) AppendAttribute(entry OutputEntry) {
	r.attributes.Append(entry)
}

func (r TFResource) MaxAttributeLength() int {
	var i, itemLen int
	for _, item := range r.Attributes() {
		itemLen = len(item.Name)
		if itemLen > i {
			i = itemLen
		}
	}

	return itemLen
}

func (r TFResource) MaxArgumentLength() int {
	var i, itemLen int
	for _, item := range r.Arguments() {
		itemLen = len(item.Name)
		if itemLen > i {
			i = itemLen
		}
	}

	return itemLen
}

func (r TFResource) Arguments() []InputEntry {
	return r.arguments.Arguments
}

func (r TFResource) Attributes() []OutputEntry {
	return r.attributes.Attributes
}

func (r *TFResource) Create() error {
	// check if AbsolutePath exists
	if _, err := os.Stat(r.AbsolutePath); os.IsNotExist(err) {
		// create directory
		if err := os.MkdirAll(r.AbsolutePath, 0754); err != nil {
			return err
		}
	}
	//create main.tf
	mainTemplatable := NewTFMain(r)
	if err := createOrAppendToFile(mainTemplatable); err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}

	// create variables.tf
	inputTemplatable := NewTFInput(r)
	if err := createOrAppendToFile(inputTemplatable); err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}

	// create output.tf
	outputTemplatable := NewTFOutput(r)
	if err := createOrAppendToFile(outputTemplatable); err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}

	return nil
}

type IOEntry struct {
	*schema.Schema
	Resource    *TFResource
	Name        string
	TFType      string
	Optional    bool
	Deprecated  bool
	Description string
}

func (e IOEntry) Padding(max int) string {
	return strings.Repeat(" ", max-e.NameLength())
}

func (e IOEntry) NameLength() int {
	return len(e.Name)
}

func (e IOEntry) ValueType() string {
	if e.Schema == nil {
		return schema.TypeString.String()
	} else {
		return formatType(e.Schema)
	}
}

func formatType(s *schema.Schema) string {
	switch s.Type {
	case schema.TypeBool:
		return "bool"
	case schema.TypeInt, schema.TypeFloat:
		return "number"
	case schema.TypeString:
		return "string"
	case schema.TypeList, schema.TypeSet:
		// TODO
		// note: not supporting the recursive functionality right now to
		// get the parsing needed of lists of resources etc.
		if reflect.TypeOf(s.Elem).String() == "*schema.Schema" {
			return "list(" + formatType(s.Elem.(*schema.Schema)) + ")"
		}
		return "list(any)"
	case schema.TypeMap:
		if reflect.TypeOf(s.Elem).String() == "schema.Schema" {
			return "map(" + formatType(s.Elem.(*schema.Schema)) + ")"
		}
		return "map(any)"
	default:
		return ""
	}
}

func (e IOEntry) DefaultValue() string {
	if e.Schema == nil {
		return ""
	} else {
		val, _ := e.Schema.DefaultValue()
		if val == nil {
			return ""
		} else if e.Schema.Type == schema.TypeString {
			return fmt.Sprintf(`"%v"`, val)
		}
		return fmt.Sprintf("%v", val)
	}
}

// Arguments
type InputEntry IOEntry

func (e InputEntry) Padding(max int) string {
	return IOEntry(e).Padding(max)
}

func (e InputEntry) ValueType() string {
	return IOEntry(e).ValueType()
}

func (e InputEntry) DefaultValue() string {
	return IOEntry(e).DefaultValue()
}

func (e InputEntry) PrefixedName() string {
	return strings.Join([]string{e.Resource.VarPrefix, e.Name}, "_")
}

// Attributes
type OutputEntry IOEntry

func (e OutputEntry) Padding(max int) string {
	return IOEntry(e).Padding(max)
}

func (e OutputEntry) ValueType() string {
	return IOEntry(e).ValueType()
}

func (e OutputEntry) DefaultValue() string {
	return IOEntry(e).DefaultValue()
}

func (e OutputEntry) PrefixedName() string {
	return strings.Join([]string{e.Resource.AttrPrefix, e.Name}, "_")
}

// TFMain represents the main instantiation of a terraform resource
type TFMain struct {
	*TFResource
	templateName string
	filename     string
}

func NewTFMain(resource *TFResource) TFMain {
	return TFMain{
		templateName: "main",
		TFResource:   resource,
		filename:     "main.tf",
	}
}

func (in TFMain) TemplateName() string {
	return in.templateName
}

func (in TFMain) FilePath() string {
	return fmt.Sprintf("%s/%s", in.AbsolutePath, in.filename)
}

func (in TFMain) Template() *template.Template {
	t := template.New(in.TemplateName())
	t = t.Funcs(template.FuncMap{"tfStringFormat": tpl.TFStringFormatter})
	return template.Must(t.Parse(string(tpl.ResourceTemplate())))
}

func (in TFMain) Create(file *os.File) error {
	err := in.Template().Execute(file, in)
	if err != nil {
		return fmt.Errorf("failed to execute %s template: %w", in.TemplateName(), err)
	}
	return nil
}

// TFOutput represents the outputs of a terraform resource
type TFOutput struct {
	*TFResource
	templateName string
	filename     string
}

func NewTFOutput(resource *TFResource) TFOutput {
	return TFOutput{
		templateName: "output",
		TFResource:   resource,
		filename:     "outputs.tf",
	}
}

func (o TFOutput) TemplateName() string {
	return o.templateName
}

func (o TFOutput) FilePath() string {
	return fmt.Sprintf("%s/%s", o.AbsolutePath, o.filename)
}

func (o TFOutput) Template() *template.Template {
	t := template.New(o.TemplateName())
	t = t.Funcs(template.FuncMap{"tfStringFormat": tpl.TFStringFormatter})
	return template.Must(t.Parse(string(tpl.OutputTemplate())))

}

func (o TFOutput) Create(file *os.File) error {
	err := o.Template().Execute(file, o)
	if err != nil {
		return fmt.Errorf("failed to execute %s template: %w", o.TemplateName(), err)
	}
	return nil
}

// TFInput represents the inputs of a terraform resource
type TFInput struct {
	*TFResource
	templateName string
	filename     string
}

func NewTFInput(resource *TFResource) TFInput {
	return TFInput{
		templateName: "input",
		TFResource:   resource,
		filename:     "variables.tf",
	}
}

func (in TFInput) TemplateName() string {
	return in.templateName
}

func (in TFInput) FilePath() string {
	return fmt.Sprintf("%s/%s", in.AbsolutePath, in.filename)
}

func (in TFInput) Template() *template.Template {
	t := template.New(in.TemplateName())
	t = t.Funcs(template.FuncMap{"tfStringFormat": tpl.TFStringFormatter})
	return template.Must(t.Parse(string(tpl.VariablesTemplate())))
}

func (in TFInput) Create(file *os.File) error {
	err := in.Template().Execute(file, in)
	if err != nil {
		return fmt.Errorf("failed to execute %s template: %w", in.TemplateName(), err)
	}
	return nil
}

type Templatable interface {
	TemplateName() string
	FilePath() string
	Template() *template.Template
	Create(file *os.File) error
}

func createOrAppendToFile(templatable Templatable) error {
	var (
		outputFile *os.File
		err        error
	)

	if _, err := os.Stat(templatable.FilePath()); err == nil {
		log.Println("appending to ", templatable.FilePath())
		outputFile, err = os.OpenFile(templatable.FilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to open existing %s file: %w", templatable.TemplateName(), err)
		}
		// Add a couple of new lines to the file
		// TODO: this could be made smart
		_, err = outputFile.WriteString("\n\n")
		if err != nil {
			return fmt.Errorf("failed to append new lines to %s file: %w", templatable.TemplateName(), err)
		}
	} else {
		log.Println("creating ", templatable.FilePath())
		outputFile, err = os.Create(templatable.FilePath())
		if err != nil {
			return fmt.Errorf("failed to create %s file: %w", templatable.TemplateName(), err)
		}
	}

	err = templatable.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to generate file: %w", err)
	}

	defer outputFile.Close()

	return nil
}
