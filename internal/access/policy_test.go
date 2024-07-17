package access

import (
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
				Effect: types.StringValue("permit"),
			},
			wantPolicy: `permit (
    principal,
    action,
    resource
);`,
		},
		{
			name: "simple cedar policy converts correctly",
			policy: Policy{
				Effect: types.StringValue("permit"),
				Principal: &CedarEntity{
					Expression: "==",
					Resource: eid.EID{
						Type: types.StringValue("CF::User"),
						ID:   types.StringValue("user1"),
					},
				},
				Action: &CedarEntity{
					Expression: "==",
					Resource: eid.EID{
						Type: types.StringValue("Action::Access"),
						ID:   types.StringValue("Request"),
					},
				},
				Resource: &CedarEntity{
					Expression: "==",
					Resource: eid.EID{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := PolicyToString(tt.policy)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantPolicy, got)
		})
	}
}
