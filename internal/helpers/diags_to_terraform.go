package helpers

import (
	accessv1alpha1 "github.com/common-fate/sdk/gen/commonfate/access/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// helpers.DiagsToTerraform maps Common Fate API diagnostics to Terraform diagnostics
func DiagsToTerraform(apiDiags []*accessv1alpha1.Diagnostic, tfDiags *diag.Diagnostics) {
	for _, d := range apiDiags {
		if d.Level == accessv1alpha1.DiagnosticLevel_DIAGNOSTIC_LEVEL_WARNING {
			tfDiags.AddWarning(d.Message, "")
		}
		if d.Level == accessv1alpha1.DiagnosticLevel_DIAGNOSTIC_LEVEL_ERROR {
			tfDiags.AddError(d.Message, "")
		}
	}
}
