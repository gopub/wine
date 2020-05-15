package template

import (
	"html/template"
	"io"
)

type Manager struct {
	templates []*template.Template
	funcMap   template.FuncMap
}

func NewManager() *Manager {
	return &Manager{
		templates: make([]*template.Template, 0),
		funcMap:   make(template.FuncMap),
	}
}

// AddGlobTemplate adds a template by parsing template files with pattern
func (m *Manager) AddGlobTemplate(pattern string) {
	tmpl := template.Must(template.ParseGlob(pattern))
	m.AddTemplate(tmpl)
}

// AddFilesTemplate adds a template by parsing template files
func (m *Manager) AddFilesTemplate(files ...string) {
	tmpl := template.Must(template.ParseFiles(files...))
	m.AddTemplate(tmpl)
}

// AddTextTemplate adds a template by parsing texts
func (m *Manager) AddTextTemplate(name string, texts ...string) {
	tmpl := template.New(name)
	for _, txt := range texts {
		tmpl = template.Must(tmpl.Parse(txt))
	}
	m.AddTemplate(tmpl)
}

// AddTemplate adds a template
func (m *Manager) AddTemplate(tmpl *template.Template) {
	if m.funcMap != nil {
		tmpl.Funcs(m.funcMap)
	}
	m.templates = append(m.templates, tmpl)
}

// AddTemplateFuncMap adds template functions
func (m *Manager) AddTemplateFuncMap(funcMap template.FuncMap) {
	if len(funcMap) == 0 {
		return
	}

	if m.funcMap == nil {
		m.funcMap = funcMap
	} else {
		for name, f := range funcMap {
			m.funcMap[name] = f
		}
	}

	for _, tmpl := range m.templates {
		tmpl.Funcs(funcMap)
	}
}

func (m *Manager) Execute(w io.Writer, name string, params interface{}) {
	for _, tmpl := range m.templates {
		var err error
		if name == "" {
			err = tmpl.Execute(w, params)
		} else {
			err = tmpl.ExecuteTemplate(w, name, params)
		}

		if err == nil {
			break
		}
	}
}

func (m *Manager) Templates() []*template.Template {
	return m.templates
}
