package access

import (
	"bytes"
	"text/template"

	"github.com/common-fate/terraform-provider-commonfate/pkg/eid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StructuredEmbeddedExpression struct {
	Resource   types.String
	Expression types.String
	Value      types.String
}

type EmbeddedExpression struct {
	Text               *types.String
	EmbeddedExpression *StructuredEmbeddedExpression
}

type CedarEntity struct {
	Expression string
	Resource   eid.EID
}

type Policy struct {
	Effect    types.String
	Principal *CedarEntity
	Action    *CedarEntity
	Resource  *CedarEntity
	When      *EmbeddedExpression
	Unless    *EmbeddedExpression
}

type PolicyModel struct {
	ID       types.String  `tfsdk:"id"`
	Text     *types.String `tfsdk:"text"`
	Policies *[]Policy     `tfsdk:"Policies"`
}

const cedarPolicyTemplate = `{{.Effect.ValueString}} (
    principal{{if .Principal}} {{.Principal.Expression}} {{.Principal.Resource.Type.ValueString}}::{{.Principal.Resource.ID}}{{else}}{{end}},
    action{{if .Action}} {{.Action.Expression}} {{.Action.Resource.Type.ValueString}}::{{.Action.Resource.ID}}{{else}}{{end}},
    resource{{if .Resource}} {{.Resource.Expression}} {{.Resource.Resource.Type.ValueString}}::{{.Resource.Resource.ID}}{{else}}{{end}}
);{{if .When}} when {
    {{if .When}}
    {{if .When.Text}}
    {{.When.Text}}
    {{else if .When.EmbeddedExpression}}
    {{.When.EmbeddedExpression.Resource}}{{.When.EmbeddedExpression.Expression}}{{.When.EmbeddedExpression.Value}}
    {{end}}
    {{else}}
	{{end}}
}{{end}}`

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

	return result.String(), nil
}
