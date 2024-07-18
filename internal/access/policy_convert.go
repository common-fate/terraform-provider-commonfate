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
	Text               types.String                  `tfsdk:"text"`
	EmbeddedExpression *StructuredEmbeddedExpression `tfsdk:"structured_embedded_expression"`
}

type CedarScopeEntity struct {
	Expression string  `tfsdk:"expression"`
	Resource   eid.EID `tfsdk:"resource"`
}

type Policy struct {
	Effect    types.String          `tfsdk:"effect"`
	Principal *CedarScopeEntity     `tfsdk:"principal"`
	Action    *CedarScopeEntity     `tfsdk:"action"`
	Resource  *CedarScopeEntity     `tfsdk:"resource"`
	When      *CedarConditionEntity `tfsdk:"when"`
	Unless    *CedarConditionEntity `tfsdk:"unless"`
}

const cedarPolicyTemplate = `{{.Effect.ValueString}} (
    principal{{if .Principal}} {{.Principal.Expression}} {{.Principal.Resource.Type.ValueString}}::{{.Principal.Resource.ID}}{{end}},
    action{{if .Action}} {{.Action.Expression}} {{.Action.Resource.Type.ValueString}}::{{.Action.Resource.ID}}{{end}},
    resource{{if .Resource}} {{.Resource.Expression}} {{.Resource.Resource.Type.ValueString}}::{{.Resource.Resource.ID}}{{end}}
){{if .When}}
when {
{{if .When.Text}} {{.When.Text.ValueString}} {{else if .When.EmbeddedExpression}} {{.When.EmbeddedExpression.Resource.ValueString}} {{.When.EmbeddedExpression.Expression.ValueString}} {{.When.EmbeddedExpression.Value.ValueString}}{{end}}
}{{end}}{{if .Unless}}
unless {
{{if .Unless.Text}} {{.Unless.Text.ValueString}} {{else if .Unless.EmbeddedExpression}} {{.Unless.EmbeddedExpression.Resource.ValueString}} {{.Unless.EmbeddedExpression.Expression.ValueString}} {{.Unless.EmbeddedExpression.Value.ValueString}}{{end}}
}{{end}};`

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
