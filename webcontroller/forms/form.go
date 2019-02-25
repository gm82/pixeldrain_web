package forms

import (
	"fmt"
	"html/template"
	"net/http"
)

// Form is a form which can be rendered in HTML and submitted
type Form struct {
	// Name of the form. When this form is submitted this name will be in the `form` parameter
	Name string

	Title       string        // Shown in a large font above the form
	PreFormHTML template.HTML // Content to be rendered above the form

	Fields []Field

	BackLink    string // Empty for no back link
	SubmitLabel string // Label for the submit button
	SubmitRed   bool   // If the submit button should be red or green

	PostFormHTML template.HTML // Content to be rendered below the form

	// Fields to render if the form has been submitted once
	Submitted      bool            // If the form has been submitted
	SubmitSuccess  bool            // If the submission was a success
	SubmitMessages []template.HTML // Messages telling the user the results

	// Used for letting the browser know which user is logged in
	Username string
}

// Field is a single input field in a form
type Field struct {
	// Used for reading the data. Entered data is POSTed back to the same URL with this name
	Name string

	// Is entered in the input field by default. If this is empty when running
	// Form.ReadInput() it will be set to the value entered by the user
	DefaultValue string

	// The value entered by the user. Filled in when running Form.ReadInput()
	EnteredValue string

	// Text next to the input field
	Label string

	// Text below the input field
	Description string

	// Separates fields with a horizontal rule
	Separator bool

	Type FieldType

	// Only used when Type = `captcha`. When using reCaptcha the field name has
	// to be `recaptcha_response`
	CaptchaSiteKey string
}

// FieldType defines the type a form field has and how it should be rendered
type FieldType string

// Fields which can be in a form
const (
	FieldTypeText            FieldType = "text"
	FieldTypeUsername        FieldType = "username"
	FieldTypeEmail           FieldType = "email"
	FieldTypeCurrentPassword FieldType = "current-password"
	FieldTypeNewPassword     FieldType = "new-password"
	FieldTypeCaptcha         FieldType = "captcha"
)

// ReadInput reads the form of a request and fills in the values for each field.
// The return value will be true if this form was submitted and false if the
// form was not submitted
func (f *Form) ReadInput(r *http.Request) (success bool) {
	if r.FormValue("form") != f.Name {
		f.Submitted = false
		return false
	}
	f.Submitted = true

	var val string

	for i, field := range f.Fields {
		val = r.FormValue(field.Name)
		field.EnteredValue = val

		if field.DefaultValue == "" {
			field.DefaultValue = val
		}
		f.Fields[i] = field // Update the new values in the array
	}

	return true
}

// FieldVal is a utility function for getting the entered value of a field by
// its name. By using this function you don't have to use nondescriptive array
// indexes to get the values. It panics if the field name is not found in the
// form
func (f *Form) FieldVal(name string) (enteredValue string) {
	for _, field := range f.Fields {
		if field.Name == name {
			return field.EnteredValue
		}
	}
	panic(fmt.Errorf("FieldVal called on unregistered field name '%s'", name))
}
