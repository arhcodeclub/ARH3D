package handlers

import (
//    "fmt"
    "html/template"
    "net/http"
    "os"
    "path/filepath"

    "github.com/arhcodeclub/arh3d/internal/db"
    "github.com/arhcodeclub/arh3d/internal/models"
)

// Serve templates
func RenderTemplate(w http.ResponseWriter, files ...string) error {
    tmpl, err := template.ParseFiles(files...)
    if err != nil {
        return err
    }
    return tmpl.ExecuteTemplate(w, filepath.Base(files[len(files)-1]), nil)
}

// NewRequestHandler handles form submission and template rendering
func NewRequestHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        // Serve form template
        err := RenderTemplate(w,
            "internal/http/templates/layout.html",
            "internal/http/templates/new.html",
        )
        if err != nil {
            http.Error(w, "Template error", 500)
            return
        }

    case http.MethodPost:
        // Parse form (max 50MB)
        err := r.ParseMultipartForm(50 << 20)
        if err != nil {
            http.Error(w, "Failed to parse form", 400)
            return
        }

        name := r.FormValue("name")
        inputType := r.FormValue("input_type")
        link := r.FormValue("link")
        description := r.FormValue("description")
        colour := r.FormValue("colour")
        comments := r.FormValue("comments")
        filePath := ""

        // Handle file upload
        if inputType == "file" {
            file, header, err := r.FormFile("file")
            if err == nil {
                defer file.Close()
                os.MkdirAll("uploads", os.ModePerm)
                filePath = filepath.Join("uploads", header.Filename)
                dst, err := os.Create(filePath)
                if err != nil {
                    http.Error(w, "Failed to save file", 500)
                    return
                }
                defer dst.Close()
                _, err = dst.ReadFrom(file)
                if err != nil {
                    http.Error(w, "Failed to save file", 500)
                    return
                }
            }
        }

        request := models.PrintRequest{
            Name:        name,
            InputType:   inputType,
            FilePath:    filePath,
            Link:        link,
            Description: description,
            Colour:      colour,
            Comments:    comments,
        }

        if err := db.DB.Create(&request).Error; err != nil {
            http.Error(w, "Failed to save request", 500)
            return
        }

        http.Redirect(w, r, "/", http.StatusSeeOther)

    default:
        http.Error(w, "Method not allowed", 405)
    }
}

