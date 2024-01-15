package whatsapp

import (
	"encoding/json"
	"strings"

	"github.com/nyaruka/courier"
	"github.com/nyaruka/courier/utils"
	"github.com/pkg/errors"
)

type MsgTemplating struct {
	Template struct {
		Name string `json:"name" validate:"required"`
		UUID string `json:"uuid" validate:"required"`
	} `json:"template" validate:"required,dive"`
	Namespace string   `json:"namespace"`
	Variables []string `json:"variables"`
	Params    map[string][]struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"params"`
	Language string `json:"language"`
}

func GetTemplating(msg courier.MsgOut) (*MsgTemplating, error) {
	if len(msg.Metadata()) == 0 {
		return nil, nil
	}

	metadata := &struct {
		Templating *MsgTemplating `json:"templating"`
	}{}
	if err := json.Unmarshal(msg.Metadata(), metadata); err != nil {
		return nil, err
	}

	if metadata.Templating == nil {
		return nil, nil
	}

	if err := utils.Validate(metadata.Templating); err != nil {
		return nil, errors.Wrapf(err, "invalid templating definition")
	}

	return metadata.Templating, nil
}

func GetTemplatePayload(templating MsgTemplating, lang string) *Template {
	template := Template{Name: templating.Template.Name, Language: &Language{Policy: "deterministic", Code: lang}}

	for k, v := range templating.Params {
		if strings.HasPrefix(k, "button.") {

			for _, p := range v {
				if strings.HasPrefix(p.Value, "http") {
					component := &Component{Type: "button", Index: strings.TrimPrefix(k, "button."), SubType: "quick_reply"}
					component.Params = append(component.Params, &Param{Type: "url", Text: p.Value})
					template.Components = append(template.Components, component)
				} else {
					component := &Component{Type: "button", Index: strings.TrimPrefix(k, "button."), SubType: "quick_reply"}
					component.Params = append(component.Params, &Param{Type: "payload", Payload: p.Value})
					template.Components = append(template.Components, component)
				}
			}

		} else if k == "header" {
			component := &Component{Type: "header"}
			for _, p := range v {
				if p.Type == "image" {
					component.Params = append(component.Params, &Param{Type: p.Type, Image: &struct {
						Link string "json:\"link,omitempty\""
					}{Link: p.Value}})
				} else if p.Type == "video" {
					component.Params = append(component.Params, &Param{Type: p.Type, Video: &struct {
						Link string "json:\"link,omitempty\""
					}{Link: p.Value}})
				} else if p.Type == "document" {
					component.Params = append(component.Params, &Param{Type: p.Type, Document: &struct {
						Link string "json:\"link,omitempty\""
					}{Link: p.Value}})
				} else {
					component.Params = append(component.Params, &Param{Type: p.Type, Text: p.Value})
				}
			}
			template.Components = append(template.Components, component)

		} else {
			component := &Component{Type: "body"}
			for _, p := range v {
				component.Params = append(component.Params, &Param{Type: p.Type, Text: p.Value})
			}
			template.Components = append(template.Components, component)

		}

	}

	if len(templating.Params) == 0 {
		component := &Component{Type: "body"}

		for _, v := range templating.Variables {
			component.Params = append(component.Params, &Param{Type: "text", Text: v})
		}
		template.Components = append(template.Components, component)

	}

	return &template

}
