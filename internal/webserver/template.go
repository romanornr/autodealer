package webserver

import (
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type pageTemplate struct {
	preloads []string
	template *template.Template
}

type templates struct {
	templates map[string]pageTemplate
	folder    string
	exec      func(string, interface{}) (string, error)
	addErr    error
}

func newTemplates(folder string, reload bool) *templates {
	t := &templates{
		templates: make(map[string]pageTemplate),
		folder:    folder,
	}
	t.exec = t.execTemplateToString
	if reload {
		t.exec = t.exectWithReload
	}
	return t
}

// filepath constructs the templates path from the templates ID
func (t *templates) filepath(name string) string {
	return filepath.Join(t.folder, name+".tmpl")
}

func (t *templates) addTemplate(name string, preloads ...string) *templates {
	if t.addErr != nil {
		return t
	}
	files := make([]string, 0, len(preloads)+1)
	for i := range preloads {
		files = append(files, preloads[i])
	}
	files = append(files, t.filepath(name))
	temp, err := template.New(name).Funcs(templateFuncs).ParseFiles(files...)
	if err != nil {
		t.addErr = fmt.Errorf("error adding templates %s: %w", name, err)
		return t
	}
	t.templates[name] = pageTemplate{
		preloads: preloads,
		template: temp,
	}
	return t
}

// buildErr returns any error encountered during addTemplate the error is cleared
func (t *templates) buildErr() error {
	err := t.addErr
	t.addErr = nil
	return err
}

func (t *templates) reloadTemplates() error {
	var errorStrings []string
	for name, tmpl := range t.templates {
		t.addTemplate(name, tmpl.preloads...)
		if t.buildErr() != nil {
			logrus.Errorf(t.buildErr().Error())
		}
	}
	if errorStrings == nil {
		return nil
	}
	return fmt.Errorf(strings.Join(errorStrings, " | "))
}

func (t *templates) execTemplateToString(name string, data interface{}) (string, error) {
	temp, ok := t.templates[name]
	if !ok {
		return "", fmt.Errorf("templates %s unknown", name)
	}
	var page strings.Builder
	err := temp.template.ExecuteTemplate(&page, name, data)
	return page.String(), err
}

// exectWithReload is the same as execTemplateToString but will reload the templates first
func (t *templates) exectWithReload(name string, data interface{}) (string, error) {
	tmpl, found := t.templates[name]
	if !found {
		return "", fmt.Errorf("templates %s nout found", name)
	}
	t.addTemplate(name, tmpl.preloads...)
	logrus.Debugf("reloaded HTML templates %q\n", name)
	return t.execTemplateToString(name, data)
}

var templateFuncs = template.FuncMap{
	"toUpper": strings.ToUpper,
	// urlBase attempts to get the domain name without TLD
	"urlBase": func(uri string) string {
		u, err := url.Parse(uri)
		if err != nil {
			logrus.Errorf("failed to parse URL: %s", uri)
		}
		return u.Host
	},
}
