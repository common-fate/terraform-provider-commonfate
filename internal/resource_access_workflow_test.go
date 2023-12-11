package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAccessWorkflow(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{

				Config: providerConfig + `

				
				  
				  resource "commonfate_access_workflow" "test" {
					name     = "daily" //this is optional
					duration = "2h"
					priority = 1
				  }
				  

`,
				Check: resource.ComposeAggregateTestCheckFunc(

					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "name", "test"),
					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "duration", "test"),
					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "priority", "1"),

					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("commonfate_access_workflow.test", "id"),
				),
			},
			// ImportState testing
			// {
			// 	ResourceName:      "commonfate_access_workflow.test",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
			// Update and Read testing
			{
				Config: providerConfig + `

				
				  resource "commonfate_access_workflow" "test" {
					name   = "daily-updated"
					duration = "1h"
					priority = 2
				  }
				  
						`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify first order item updated
					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "name", "daily-updated"),
					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "duration", "1h"),
					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "priority", "2"),

					// Verify dynamic values have any value set in the state.
					// resource.TestCheckResourceAttrSet("commonfate_access_workflow.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
