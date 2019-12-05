package wine

import (
	"context"
	"html/template"
)

type TemplateManager struct {
	templates     []*template.Template
	templateFuncs template.FuncMap
}

func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates:     make([]*template.Template, 0),
		templateFuncs: make(template.FuncMap),
	}
}

// AddGlobTemplate adds a template by parsing template files with pattern
func (m *TemplateManager) AddGlobTemplate(pattern string) {
	tmpl := template.Must(template.ParseGlob(pattern))
	m.AddTemplate(tmpl)
}

// AddFilesTemplate adds a template by parsing template files
func (m *TemplateManager) AddFilesTemplate(files ...string) {
	tmpl := template.Must(template.ParseFiles(files...))
	m.AddTemplate(tmpl)
}

// AddTextTemplate adds a template by parsing texts
func (m *TemplateManager) AddTextTemplate(name string, texts ...string) {
	tmpl := template.New(name)
	for _, txt := range texts {
		tmpl = template.Must(tmpl.Parse(txt))
	}
	m.AddTemplate(tmpl)
}

// AddTemplate adds a template
func (m *TemplateManager) AddTemplate(tmpl *template.Template) {
	if m.templateFuncs != nil {
		tmpl.Funcs(m.templateFuncs)
	}
	m.templates = append(m.templates, tmpl)
}

// AddTemplateFuncMap adds template functions
func (m *TemplateManager) AddTemplateFuncMap(funcMap template.FuncMap) {
	if funcMap == nil {
		logger.Panic("funcMap is nil")
	}

	if m.templateFuncs == nil {
		m.templateFuncs = funcMap
	} else {
		for name, f := range funcMap {
			m.templateFuncs[name] = f
		}
	}

	for _, tmpl := range m.templates {
		tmpl.Funcs(funcMap)
	}
}

func GetTemplates(ctx context.Context) []*template.Template {
	v, _ := ctx.Value(CKTemplates).([]*template.Template)
	return v
}

func withTemplate(ctx context.Context, templates []*template.Template) context.Context {
	return context.WithValue(ctx, CKTemplates, templates)
}
