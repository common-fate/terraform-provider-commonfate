package access

import (
	"bytes"
	"fmt"
	"strings"
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
	Effect types.String  `tfsdk:"effect"`
	Advice *types.String `tfsdk:"advice"`

	Principal   *eid.EID   `tfsdk:"principal"`
	PrincipalIn *[]eid.EID `tfsdk:"principal_in"`
	PrincipalIs *eid.EID   `tfsdk:"principal_is"`

	Action   *eid.EID   `tfsdk:"action"`
	ActionIn *[]eid.EID `tfsdk:"action_in"`
	ActionIs *eid.EID   `tfsdk:"action_is"`

	Resource   *eid.EID   `tfsdk:"resource"`
	ResourceIn *[]eid.EID `tfsdk:"resource_in"`
	ResourceIs *eid.EID   `tfsdk:"resource_is"`

	When   *CedarConditionEntity `tfsdk:"when"`
	Unless *CedarConditionEntity `tfsdk:"unless"`
}

// builds the scope fields (principal, action, resource) since they will all follow the same patterns for being built
func buildCedarScopeField(scopeType string, includeTrailingComma bool) string {
	toLowerName := strings.ToLower(scopeType)
	//We make some variables in the template to work out if for a given number of scopeIn fields if we need to add the delimiting comma eg.
	//{{$len := len .%sIn }}{{$actLen := minus $len 1}} this is making a variable $len = length(principalIn) then $actLen = $len - 1 which is the actual length of the list

	out := fmt.Sprintf(`%s{{if .%s}} == {{.%s.Type.ValueString}}::{{.%s.ID}}{{end}}{{if .%sIs}} is {{.%sIs.Type.ValueString}}::{{.%sIs.ID}}{{end}}{{if .%sIn}}{{$len := len .%sIn }}{{$actLen := minus $len 1}} in [{{range $i, $val := .%sIn}}{{$val.Type.ValueString}}::{{$val.ID}}{{if (ne $i $actLen )}}, {{end}}{{end}}]{{end}}`, toLowerName, scopeType, scopeType, scopeType, scopeType, scopeType, scopeType, scopeType, scopeType, scopeType)

	if includeTrailingComma {
		out = out + ", "
	}
	return out
}

const cedarAdviceTemplate = `{{if .Advice}}@advice({{.Advice}}){{end}}`
const cedarEffectTemplate = `{{.Effect.ValueString}}`

var cedarPrincipalTemplate = buildCedarScopeField("Principal", true)
var cedarActionTemplate = buildCedarScopeField("Action", true)
var cedarResourceTemplate = buildCedarScopeField("Resource", false)

const cedarWhenTemplate = `{{if .When}}
when {
{{if .When.Text}} {{.When.Text.ValueString}} {{else if .When.EmbeddedExpression}} {{.When.EmbeddedExpression.Resource.ValueString}} {{.When.EmbeddedExpression.Expression.ValueString}} {{.When.EmbeddedExpression.Value.ValueString}} {{end}}
}{{end}}`

const cedarUnlessTemplate = `{{if .Unless}}
unless {
{{if .Unless.Text}} {{.Unless.Text.ValueString}} {{else if .Unless.EmbeddedExpression}} {{.Unless.EmbeddedExpression.Resource.ValueString}} {{.Unless.EmbeddedExpression.Expression.ValueString}} {{.Unless.EmbeddedExpression.Value.ValueString}} {{end}}
}{{end}}`

var cedarPolicyTemplateTest = cedarAdviceTemplate + cedarEffectTemplate + " ( " + cedarPrincipalTemplate + cedarActionTemplate + cedarResourceTemplate + " )" + cedarWhenTemplate + cedarUnlessTemplate + ";"

func PolicyToString(policy Policy) (string, error) {

	//adds minus function to template to allow checking length of the resources
	funcMap := template.FuncMap{
		"minus": func(i, k int) int {
			return i - k
		},
	}

	tmpl, err := template.New("cedarPolicy").Funcs(funcMap).Parse(cedarPolicyTemplateTest)
	if err != nil {
		if err != nil {
			return "", err
		}
	}

	var result bytes.Buffer
	err = tmpl.Execute(&result, policy)
	if err != nil {
		return "", err
	}

	res := result.String()

	return res, nil
}
