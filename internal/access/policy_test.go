package access

import (
	"strings"
	"testing"

	"github.com/common-fate/terraform-provider-commonfate/pkg/eid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestReviewer_Approve(t *testing.T) {

	tests := []struct {
		name       string
		policy     Policy
		wantPolicy string
	}{

		{
			name: "simple allow all cedar policy converts correctly",
			policy: Policy{
				Effect:       types.StringValue("permit"),
				AnyPrincipal: types.BoolValue(true),
				AnyAction:    types.BoolValue(true),
				AnyResource:  types.BoolValue(true),
			},
			wantPolicy: `permit ( principal, action, resource );`,
		},
		{
			name: "simple allow all with advice correctly converts",
			policy: Policy{
				Effect: types.StringValue("permit"),
				Annotation: &CedarAnnotation{
					Name:  types.StringValue("advice"),
					Value: types.StringValue("test"),
				},
				AnyPrincipal: types.BoolValue(true),
				AnyAction:    types.BoolValue(true),
				AnyResource:  types.BoolValue(true),
			},
			wantPolicy: `@advice("test")
permit (
 principal,
 action,
 resource );`,
		},
		{
			name: "simple cedar policy converts correctly",
			policy: Policy{
				Effect: types.StringValue("permit"),
				Principal: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("CF::User"),
						ID:   types.StringValue("user1"),
					},
				},
				Action: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("Action::Access"),
						ID:   types.StringValue("Request"),
					},
				},
				Resource: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("Test::Vault"),
						ID:   types.StringValue("test1"),
					},
				},
			},
			wantPolicy: `permit (
 principal == CF::User::"user1",
 action == Action::Access::"Request",
 resource == Test::Vault::"test1" 
);`,
		},
		{
			name: "simple cedar policy with text when converts correctly",
			policy: Policy{
				Effect: types.StringValue("permit"),
				Principal: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("CF::User"),
						ID:   types.StringValue("user1"),
					},
				},
				Action: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("Action::Access"),
						ID:   types.StringValue("Request"),
					},
				},
				Resource: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("Test::Vault"),
						ID:   types.StringValue("test1"),
					},
				},
				When: &[]CedarConditionEntity{
					{Text: types.StringValue("true")},
				},
			},
			wantPolicy: `permit (
 principal == CF::User::"user1",
 action == Action::Access::"Request",
 resource == Test::Vault::"test1" 
)
when { true };`,
		},

		{
			name: "simple cedar policy with text unless converts correctly",
			policy: Policy{
				Effect: types.StringValue("permit"),
				Principal: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("CF::User"),
						ID:   types.StringValue("user1"),
					},
				},
				Action: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("Action::Access"),
						ID:   types.StringValue("Request"),
					},
				},
				Resource: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("Test::Vault"),
						ID:   types.StringValue("test1"),
					},
				},

				Unless: &[]CedarConditionEntity{
					{Text: types.StringValue("true")},
				},
			},
			wantPolicy: `permit (
 principal == CF::User::"user1",
 action == Action::Access::"Request",
 resource == Test::Vault::"test1" 
)
unless { true };`,
		},

		{
			name: "test in condition",
			policy: Policy{
				Effect: types.StringValue("permit"),
				PrincipalIn: &[]eid.EID{
					eid.EID{
						Type: types.StringValue("CF::User"),
						ID:   types.StringValue("user1"),
					},
				},

				ActionIn: &[]eid.EID{
					eid.EID{
						Type: types.StringValue("Action::Access"),
						ID:   types.StringValue("Request"),
					},
				},

				ResourceIn: &[]eid.EID{
					eid.EID{
						Type: types.StringValue("Test::Vault"),
						ID:   types.StringValue("test1"),
					},
				},
			},
			wantPolicy: `permit (
 principal in [CF::User::"user1"],
 action in [Action::Access::"Request"],
 resource in [Test::Vault::"test1"] 
);`,
		},
		{
			name: "test in condition with multiple values",
			policy: Policy{
				Effect: types.StringValue("permit"),
				PrincipalIn: &[]eid.EID{
					eid.EID{
						Type: types.StringValue("CF::User"),
						ID:   types.StringValue("user1"),
					},
					eid.EID{
						Type: types.StringValue("CF::User"),
						ID:   types.StringValue("user2"),
					},
				},

				ActionIn: &[]eid.EID{
					eid.EID{
						Type: types.StringValue("Action::Access"),
						ID:   types.StringValue("Request"),
					},
					eid.EID{
						Type: types.StringValue("Action::Access"),
						ID:   types.StringValue("Close"),
					},
				},

				ResourceIn: &[]eid.EID{
					eid.EID{
						Type: types.StringValue("Test::Vault"),
						ID:   types.StringValue("test1"),
					},
					eid.EID{
						Type: types.StringValue("Test::Vault"),
						ID:   types.StringValue("test2"),
					},
				},
			},
			wantPolicy: `permit (
 principal in [CF::User::"user1", CF::User::"user2"],
 action in [Action::Access::"Request", Action::Access::"Close"],
 resource in [Test::Vault::"test1", Test::Vault::"test2"] 
);`,
		},
		{
			name: "test is condition",
			policy: Policy{
				Effect: types.StringValue("permit"),
				PrincipalIs: &eid.EID{
					Type: types.StringValue("CF::User"),
					ID:   types.StringValue("user1"),
				},

				ActionIs: &eid.EID{
					Type: types.StringValue("Action::Access"),
					ID:   types.StringValue("Request"),
				},

				ResourceIs: &eid.EID{
					Type: types.StringValue("Test::Vault"),
					ID:   types.StringValue("test1"),
				},
			},
			wantPolicy: `permit (
 principal is CF::User::"user1",
 action is Action::Access::"Request",
 resource is Test::Vault::"test1" 
);`,
		},
		{
			name: "test having multiple when conditions",
			policy: Policy{
				Effect: types.StringValue("permit"),
				Principal: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("CF::User"),
						ID:   types.StringValue("user1"),
					},
				},
				Action: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("Action::Access"),
						ID:   types.StringValue("Request"),
					},
				},
				Resource: &ScopeConditionType{
					EID: &eid.EID{
						Type: types.StringValue("Test::Vault"),
						ID:   types.StringValue("test1"),
					},
				},
				When: &[]CedarConditionEntity{
					{
						Text: types.StringValue("true"),
					},
				},
			},
			wantPolicy: `permit (
 principal == CF::User::"user1",
 action == Action::Access::"Request",
 resource == Test::Vault::"test1" 
)when { true };`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := PolicyToString(tt.policy)
			if err != nil {
				t.Fatal(err)
			}

			//remove newlines in testing to get more consistent outcomes
			gotMinusNewlines := strings.ReplaceAll(got, "\n", "")
			expectedMinusNewlines := strings.ReplaceAll(tt.wantPolicy, "\n", "")

			assert.Equal(t, expectedMinusNewlines, gotMinusNewlines)
		})
	}
}
