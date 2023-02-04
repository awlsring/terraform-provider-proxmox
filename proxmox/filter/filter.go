package filter

import (
	"encoding/base64"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/crypto/sha3"
)

type FilterConfig []string

func (f *FilterConfig) Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:             schema.TypeString,
					Description:      "The name of the attribute to filter on.",
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(*f, false)),
					Required:         true,
				},
				"values": {
					Type:        schema.TypeList,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Description: "The value(s) to be used in the filter.",
					Required:    true,
				},
			},
		},
	}
}

func MakeListId(d *schema.ResourceData) (string, error) {
	idMap := map[string]interface{}{
		"filter":   d.Get("filter"),
	}

	result, err := json.Marshal(idMap)
	if err != nil {
		return "", err
	}

	hash := sha3.Sum512(result)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}