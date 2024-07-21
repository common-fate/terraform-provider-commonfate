package access

import (
	"bytes"
	"text/template"

	"github.com/common-fate/terraform-provider-commonfate/pkg/eid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StructuredEmbeddedExpression struct {
	Resource   types.String `tfsdk:"resource"`
	Expression types.String `tfsdk:"expression"`
	Value      types.String `tfsdk:"value"`
}

type CedarConditionEntity struct {
	Text               *types.String                 `tfsdk:"text"`
	EmbeddedExpression *StructuredEmbeddedExpression `tfsdk:"structured_embedded_expression"`
}

type Policy struct {
	Effect types.String `tfsdk:"effect"`

	Principal   *eid.EID `tfsdk:"principal"`
	PrincipalIn *eid.EID `tfsdk:"principal_in"`
	PrincipalIs *eid.EID `tfsdk:"principal_is"`

	Action   *eid.EID `tfsdk:"action"`
	ActionIn *eid.EID `tfsdk:"action_in"`
	ActionIs *eid.EID `tfsdk:"action_is"`

	Resource   *eid.EID `tfsdk:"resource"`
	ResourceIn *eid.EID `tfsdk:"resource_in"`
	ResourceIs *eid.EID `tfsdk:"resource_is"`

	When   *CedarConditionEntity `tfsdk:"when"`
	Unless *CedarConditionEntity `tfsdk:"unless"`
}

const cedarPolicyTemplate = `{{.Effect.ValueString}} (
    principal{{if .Principal}} == {{.Principal.Type.ValueString}}::{{.Principal.ID}}{{end}}{{if .PrincipalIs}} is {{.PrincipalIs.Type.ValueString}}::{{.PrincipalIs.ID}}{{end}}{{if .PrincipalIn}} in {{.PrincipalIn.Type.ValueString}}::{{.PrincipalIn.ID}}{{end}},
    action{{if .Action}} == {{.Action.Type.ValueString}}::{{.Action.ID}}{{end}}{{if .ActionIs}} is {{.ActionIs.Type.ValueString}}::{{.ActionIs.ID}}{{end}}{{if .ActionIn}} in {{.ActionIn.Type.ValueString}}::{{.ActionIn.ID}}{{end}},
    resource{{if .Resource}} == {{.Resource.Type.ValueString}}::{{.Resource.ID}}{{end}}{{if .ResourceIs}} is {{.ResourceIs.Type.ValueString}}::{{.ResourceIs.ID}}{{end}}{{if .ResourceIn}} in {{.ResourceIn.Type.ValueString}}::{{.ResourceIn.ID}}{{end}}
){{if .When}}
when {
{{if .When.Text}} {{.When.Text.ValueString}} {{else if .When.EmbeddedExpression}} {{.When.EmbeddedExpression.Resource.ValueString}} {{.When.EmbeddedExpression.Expression.ValueString}} {{.When.EmbeddedExpression.Value.ValueString}}{{end}}
}{{end}}{{if .Unless}}
unless {
{{if .Unless.Text}} {{.Unless.Text.ValueString}} {{else if .Unless.EmbeddedExpression}} {{.Unless.EmbeddedExpression.Resource.ValueString}} {{.Unless.EmbeddedExpression.Expression.ValueString}} {{.Unless.EmbeddedExpression.Value.ValueString}}{{end}}}{{end}};`

func PolicyToString(policy Policy) (string, error) {
	tmpl, err := template.New("cedarPolicy").Parse(cedarPolicyTemplate)
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	err = tmpl.Execute(&result, policy)
	if err != nil {
		return "", err
	}

	res := result.String()

	return res, nil
}
