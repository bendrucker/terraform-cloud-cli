package configwrite

import "github.com/hashicorp/hcl/v2"

func errorDiags(diags hcl.Diagnostics) hcl.Diagnostics {
	result := make(hcl.Diagnostics, 0)
	for _, diag := range diags {
		if diag.Severity != hcl.DiagError {
			continue
		}
		result = append(result, diags...)
	}
	return result
}
