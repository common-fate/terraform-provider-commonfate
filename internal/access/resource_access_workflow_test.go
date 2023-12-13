package access

// func TestAccAccessWorkflow(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			// Create and Read testing
// 			{

// 				Config: providerConfig + `

// 				  resource "commonfate_access_workflow" "test" {
// 					name     = "test"
// 					access_duration_seconds = 2000
// 					priority = 1
// 					try_extend_after_seconds = 1000
// 				  }

// `,
// 				Check: resource.ComposeAggregateTestCheckFunc(

// 					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "name", "test"),
// 					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "access_duration_seconds", "2000"),
// 					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "priority", "1"),
// 					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "try_extend_after_seconds", "1000"),

// 					// Verify dynamic values have any value set in the state.
// 					resource.TestCheckResourceAttrSet("commonfate_access_workflow.test", "id"),
// 				),
// 			},
// 			// ImportState testing
// 			// {
// 			// 	ResourceName:      "commonfate_access_workflow.test",
// 			// 	ImportState:       true,
// 			// 	ImportStateVerify: true,
// 			// },
// 			// // Update and Read testing
// 			// {
// 			// 	Config: providerConfig + `

// 			// 	resource "commonfate_access_workflow" "test" {
// 			// 		name     = "test-updated"
// 			// 		access_duration = "1h"
// 			// 		priority = 2
// 			// 		try_extend_after="10m"
// 			// 	  }

// 			// 			`,
// 			// 	Check: resource.ComposeAggregateTestCheckFunc(
// 			// 		// Verify first order item updated
// 			// 		resource.TestCheckResourceAttr("commonfate_access_workflow.test", "name", "test-updated"),
// 			// 		resource.TestCheckResourceAttr("commonfate_access_workflow.test", "access_duration", "1h"),
// 			// 		resource.TestCheckResourceAttr("commonfate_access_workflow.test", "priority", "2"),
// 			// 		resource.TestCheckResourceAttr("commonfate_access_workflow.test", "try_extend_after", "10m"),

// 			// 		// Verify dynamic values have any value set in the state.
// 			// 		// resource.TestCheckResourceAttrSet("commonfate_access_workflow.test", "id"),
// 			// 	),
// 			// },
// 			// Delete testing automatically occurs in TestCase
// 		},
// 	})
// }
