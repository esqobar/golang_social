package mailer

import (
	"bytes"
	"html/template"
)

func parseTemplate(templateFile string, data any) (subject string, body string, err error) {
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return "", "", err
	}

	subjectBuf := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(subjectBuf, "subject", data); err != nil {
		return "", "", err
	}

	bodyBuf := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(bodyBuf, "body", data); err != nil {
		return "", "", err
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}
