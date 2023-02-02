package proxmox

import (
	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ValidateDiskType() schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		disk, ok := i.(string)
		if !ok {
			return diag.Errorf("expected type of %q to be string", k)
		}
		_, err := proxmox.NewDiskTypeFromValue(disk)
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
}
