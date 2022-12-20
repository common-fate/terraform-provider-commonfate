package commonfate

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRuleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "commonfate_access_rule" "test" {
   name ="test"
  description="test"
  groups=["common_fate_administrators"]
  approval= {
      users=["jack@commonfate.io"]
  }
  
  target=[
    {
      field="accountId"
      value=["123456789123"]
    },
    {
      field="permissionSetArn"
      value=["arn:aws:sso:::permissionSet/ssoins-825968feece9a0b6/ps-dda57372ebbfeb95"]
    }
  ]
  target_provider_id="aws-sso-v2"
  duration="3600"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(

					resource.TestCheckResourceAttr("commonfate_access_rule.test", "name", "test"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "description", "test"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "groups.0", "common_fate_administrators"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "approval.users.0", "jack@commonfate.io"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "target.0.field", "accountId"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "target.0.value.0", "123456789123"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "target.1.field", "permissionSetArn"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "target.1.value.0", "arn:aws:sso:::permissionSet/ssoins-825968feece9a0b6/ps-dda57372ebbfeb95"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "target_provider_id", "aws-sso-v2"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "duration", "3600"),

					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("commonfate_access_rule.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "commonfate_access_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
						resource "commonfate_access_rule" "test" {
			   name ="test-updated"
			  description="test-updated"
			  groups=["common_fate_administrators"]
			  approval= {
			      users=["jack@commonfate.io"]
			  }

			  target=[
			    {
			      field="accountId"
			      value=["123456789123"]
			    },
			    {
			      field="permissionSetArn"
			      value=["arn:aws:sso:::permissionSet/ssoins-825968feece9a0b6/ps-dda57372ebbfeb95"]
			    }
			  ]
			  target_provider_id="aws-sso-v2"
			  duration="3600"
			}
						`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify first order item updated
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "name", "test-updated"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "description", "test-updated"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "groups.0", "common_fate_administrators"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "approval.users.0", "jack@commonfate.io"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "approval.users.1", "jordi@commonfate.io"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "target.0.field", "accountId"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "target.0.value.0", "123456789123"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "target.1.field", "permissionSetArn"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "target.1.value.0", "arn:aws:sso:::permissionSet/ssoins-825968feece9a0b6/ps-dda57372ebbfeb95"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "target_provider_id", "aws-sso-v2"),
					resource.TestCheckResourceAttr("commonfate_access_rule.test", "duration", "3600"),

					// Verify dynamic values have any value set in the state.
					// resource.TestCheckResourceAttrSet("commonfate_access_rule.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
