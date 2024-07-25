package access

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/common-fate/terraform-provider-commonfate/pkg/eid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//unused in the conditions at the moment. Opting for just a text field to specify conditions
//In the future we may want to expand this out to allow for a more robust system for building cedar conditions
// type StructuredEmbeddedExpression struct {
// 	Resource   types.String `tfsdk:"resource"`
// 	Expression types.String `tfsdk:"expression"`
// 	Value      types.String `tfsdk:"value"`
// }

type CedarConditionEntity struct {
	Text types.String `tfsdk:"text"`
	// EmbeddedExpression *StructuredEmbeddedExpression `tfsdk:"structured_embedded_expression"`
}

type Conditions struct {
	Delimiter         types.String `tfsdk:"delimiter"`
	ConditionEntities []CedarConditionEntity
}

// same as the eid type but allows to specify an allow all flag
type ScopeConditionType struct {
	*eid.EID `tfsdk:"entity"`
}

type CedarAnnotation struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type Policy struct {
	Effect     types.String     `tfsdk:"effect"`
	Annotation *CedarAnnotation `tfsdk:"annotation"`

	AnyPrincipal types.Bool          `tfsdk:"any_principal"`
	Principal    *ScopeConditionType `tfsdk:"principal"`
	PrincipalIn  *[]eid.EID          `tfsdk:"principal_in"`
	PrincipalIs  *eid.EID            `tfsdk:"principal_is"`

	AnyAction types.Bool          `tfsdk:"any_action"`
	Action    *ScopeConditionType `tfsdk:"action"`
	ActionIn  *[]eid.EID          `tfsdk:"action_in"`
	ActionIs  *eid.EID            `tfsdk:"action_is"`

	AnyResource types.Bool          `tfsdk:"any_resource"`
	Resource    *ScopeConditionType `tfsdk:"resource"`
	ResourceIn  *[]eid.EID          `tfsdk:"resource_in"`
	ResourceIs  *eid.EID            `tfsdk:"resource_is"`

	When   *[]CedarConditionEntity `tfsdk:"when"`
	Unless *[]CedarConditionEntity `tfsdk:"unless"`
}

// builds the scope fields (principal, action, resource) since they will all follow the same patterns for being built
func buildCedarScopeField(scopeType string, includeTrailingComma bool) string {
	toLowerName := strings.ToLower(scopeType)

	anyScope := fmt.Sprintf(`Any%s`, scopeType)
	allowAll := fmt.Sprintf(`{{if .%s.ValueBool}}{{end}}`, anyScope)

	//We make some variables in the template to work out if for a given number of scopeIn fields if we need to add the delimiting comma eg.
	//{{$len := len .%sIn }}{{$actLen := minus $len 1}} this is making a variable $len = length(principalIn) then $actLen = $len - 1 which is the actual length of the list
	basicScope := fmt.Sprintf(`{{if  and .%s .%s.EID}} == {{.%s.EID.Type.ValueString}}::{{.%s.EID.ID}}{{end}}`, scopeType, scopeType, scopeType, scopeType)

	isScope := fmt.Sprintf(`{{if .%sIs}} is {{.%sIs.Type.ValueString}}::{{.%sIs.ID}}{{end}}`, scopeType, scopeType, scopeType)

	inScope := fmt.Sprintf(`{{if .%sIn}}{{$len := len .%sIn }}{{$actLen := minus $len 1}} in [{{range $i, $val := .%sIn}}{{$val.Type.ValueString}}::{{$val.ID}}{{if (ne $i $actLen )}}, {{end}}{{end}}]{{end}}`, scopeType, scopeType, scopeType)

	out := fmt.Sprintf(`%s%s%s%s%s`, toLowerName, allowAll, basicScope, isScope, inScope)

	if includeTrailingComma {
		out = out + ", "
	}
	return out
}

func buildCedarConditionFields(conditionType string) string {
	toLowerName := strings.ToLower(conditionType)

	cedarTemplate := fmt.Sprintf(`{{if .%s}}{{range $i, $val := .%s}}%s {
{{if not $val.Text.IsNull}} {{$val.Text.ValueString}} {{end}}
}{{end}}{{end}}`, conditionType, conditionType, toLowerName)

	return cedarTemplate
}

const cedarAdviceTemplate = `{{if .Annotation }}@{{.Annotation.Name.ValueString}}({{.Annotation.Value}}){{end}}`
const cedarEffectTemplate = `{{.Effect.ValueString}}`

var cedarPrincipalTemplate = buildCedarScopeField("Principal", true)
var cedarActionTemplate = buildCedarScopeField("Action", true)
var cedarResourceTemplate = buildCedarScopeField("Resource", false)

var cedarWhenTemplate = buildCedarConditionFields("When")
var cedarUnlessTemplate = buildCedarConditionFields("Unless")

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
