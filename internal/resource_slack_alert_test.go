package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSlackAlert(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{

				Config: providerConfig + `

				
				  
				  resource "commonfate_slack_alert" "test" {
					workflow_id   = "test"
					slack_channel_id = "test"
					slack_workspace_id = "test"
				  }
				  

`,
				Check: resource.ComposeAggregateTestCheckFunc(

					resource.TestCheckResourceAttr("commonfate_slack_alert.test", "workflow_id", "test"),
					resource.TestCheckResourceAttr("commonfate_slack_alert.test", "slack_channel_id", "test"),
					resource.TestCheckResourceAttr("commonfate_slack_alert.test", "slack_workspace_id", "test"),

					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("commonfate_slack_alert.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "commonfate_slack_alert.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `

				
				  
				  resource "commonfate_slack_alert" "test" {
					workflow_id   = "test-updated"
					slack_workspace_id = "test-updated"
					slack_channel_id = "test-updated"
				  }
				  
						`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify first order item updated
					resource.TestCheckResourceAttr("commonfate_slack_alert.test", "workflow_id", "test-updated"),
					resource.TestCheckResourceAttr("commonfate_slack_alert.test", "slack_channel_id", "test-updated"),
					resource.TestCheckResourceAttr("commonfate_slack_alert.test", "slack_workspace_id", "test-updated"),

					// Verify dynamic values have any value set in the state.
					// resource.TestCheckResourceAttrSet("commonfate_slack_alert.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
