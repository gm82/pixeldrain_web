package webcontroller

import (
	"fmt"
	"html/template"
	"net/http"

	"fornaxian.tech/pixeldrain_server/api/restapi/apiclient"
	"github.com/Fornaxian/log"
)

func (wc *WebController) adminGlobalsForm(td *TemplateData, r *http.Request) (f Form) {
	if !td.Authenticated || !td.User.IsAdmin {
		return Form{Title: ";-)"}
	}

	f = Form{
		Name:        "admin_globals",
		Title:       "Pixeldrain global configuration",
		PreFormHTML: template.HTML("<p>Careful! The slightest typing error could bring the whole website down</p>"),
		BackLink:    "/admin",
		SubmitLabel: "Submit",
	}

	globals, err := td.PixelAPI.AdminGetGlobals()
	if err != nil {
		f.SubmitMessages = []template.HTML{template.HTML(err.Error())}
		return f
	}
	var globalsMap = make(map[string]string)
	for _, v := range globals.Globals {
		f.Fields = append(f.Fields, Field{
			Name:         v.Key,
			DefaultValue: v.Value,
			Label:        v.Key,
			Type: func() FieldType {
				switch v.Key {
				case
					"email_address_change_body",
					"email_password_reset_body",
					"email_register_user_body":
					return FieldTypeTextarea
				case
					"api_ratelimit_limit",
					"api_ratelimit_rate",
					"cron_interval_seconds",
					"file_inactive_expiry_days",
					"max_file_size",
					"pixelstore_min_redundancy":
					return FieldTypeNumber
				default:
					return FieldTypeText
				}
			}(),
		})
		globalsMap[v.Key] = v.Value
	}

	if f.ReadInput(r) {
		var successfulUpdates = 0
		for k, v := range f.Fields {
			if v.EnteredValue == globalsMap[v.Name] {
				continue // Change changes, no need to update
			}

			// Value changed, try to update global setting
			if err = td.PixelAPI.AdminSetGlobals(v.Name, v.EnteredValue); err != nil {
				if apiErr, ok := err.(apiclient.Error); ok {
					f.SubmitMessages = append(f.SubmitMessages, template.HTML(apiErr.Message))
				} else {
					log.Error("%s", err)
					f.SubmitMessages = append(f.SubmitMessages, template.HTML(
						fmt.Sprintf("Failed to set '%s': %s", v.Name, err),
					))
					return f
				}
			} else {
				f.Fields[k].DefaultValue = v.EnteredValue
				successfulUpdates++
			}

		}
		if len(f.SubmitMessages) == 0 {
			// Request was a success
			f.SubmitSuccess = true
			f.SubmitMessages = []template.HTML{template.HTML(
				fmt.Sprintf("Success! %d values updated", successfulUpdates),
			)}
		}
	}
	return f
}

func (wc *WebController) adminAbuseForm(td *TemplateData, r *http.Request) (f Form) {
	if !td.Authenticated || !td.User.IsAdmin {
		return Form{Title: ";-)"}
	}

	f = Form{
		Name:        "admin_file_removal",
		Title:       "Admin file removal",
		PreFormHTML: template.HTML("<p>Paste any pixeldrain file links in here to remove them</p>"),
		Fields: []Field{
			{
				Name:  "text",
				Label: "Files to delete",
				Type:  FieldTypeTextarea,
			}, {
				Name:         "type",
				Label:        "Type",
				DefaultValue: "unknown",
				Description:  "Can be 'unknown', 'copyright', 'terrorism' or 'child_abuse'",
				Type:         FieldTypeText,
			}, {
				Name:         "reporter",
				Label:        "Reporter",
				DefaultValue: "pixeldrain",
				Type:         FieldTypeText,
			},
		},
		BackLink:    "/admin",
		SubmitLabel: "Submit",
	}

	if f.ReadInput(r) {
		resp, err := td.PixelAPI.AdminBlockFiles(
			f.FieldVal("text"),
			f.FieldVal("type"),
			f.FieldVal("reporter"),
		)
		if err != nil {
			formAPIError(err, &f)
			return
		}

		successMsg := template.HTML("The following files were blocked:<br/><ul>")
		for _, v := range resp.FilesBlocked {
			successMsg += template.HTML("<li>pixeldrain.com/u/" + v + "</li>")
		}
		successMsg += "<ul>"

		// Request was a success
		f.SubmitSuccess = true
		f.SubmitMessages = []template.HTML{successMsg}
	}
	return f
}
