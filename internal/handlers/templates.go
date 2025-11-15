package handlers

import (
    "html/template"
    "log"
    "net/http"
    "path/filepath"
)

func RenderTemplate(w http.ResponseWriter, files ...string) error {
    log.Printf("[TEMPLATE] Rendering %v.", files)

    tmpl, err := template.ParseFiles(files...)
    if err != nil {
        log.Printf("[TEMPLATE] Error parsing templates %v: %v", files, err)
        return err
    }

    return tmpl.ExecuteTemplate(w, filepath.Base(files[len(files)-1]), nil)
}

func RenderTemplateWithData(w http.ResponseWriter, data any, files ...string) error {
    log.Printf("[TEMPLATE] Rendering %v with data type %T.", files, data)

    tmpl, err := template.ParseFiles(files...)
    if err != nil {
        log.Printf("[TEMPLATE] Error parsing templates %v: %v", files, err)
        return err
    }

    return tmpl.ExecuteTemplate(w, filepath.Base(files[len(files)-1]), data)
}
