package wine

import (
	"html/template"
)

type templateManager struct {
	templates     []*template.Template
	templateFuncs template.FuncMap
}

func newTemplateManager() *templateManager {
	return &templateManager{
		templates:     make([]*template.Template, 0),
		templateFuncs: make(template.FuncMap),
	}
}

// AddGlobTemplate adds a template by parsing template files with pattern
func (m *templateManager) AddGlobTemplate(pattern string) {
	tmpl := template.Must(template.ParseGlob(pattern))
	m.AddTemplate(tmpl)
}

// AddFilesTemplate adds a template by parsing template files
func (m *templateManager) AddFilesTemplate(files ...string) {
	tmpl := template.Must(template.ParseFiles(files...))
	m.AddTemplate(tmpl)
}

// AddTextTemplate adds a template by parsing texts
func (m *templateManager) AddTextTemplate(name string, texts ...string) {
	tmpl := template.New(name)
	for _, txt := range texts {
		tmpl = template.Must(tmpl.Parse(txt))
	}
	m.AddTemplate(tmpl)
}

// AddTemplate adds a template
func (m *templateManager) AddTemplate(tmpl *template.Template) {
	if m.templateFuncs != nil {
		tmpl.Funcs(m.templateFuncs)
	}
	m.templates = append(m.templates, tmpl)
}

// AddTemplateFuncMap adds template functions
func (m *templateManager) AddTemplateFuncMap(funcMap template.FuncMap) {
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
