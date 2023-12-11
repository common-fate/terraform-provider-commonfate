package internal

// func TestAccPolicySet(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			// Create and Read testing
// 			{

// 				Config: providerConfig + `

// 				  resource "commonfate_policyset" "test" {
// 					id     = "test"
// 					text = "test"

// 				  }

// `,
// 				Check: resource.ComposeAggregateTestCheckFunc(

// 					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "id", "test"),
// 					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "text", "test"),
// 				),
// 			},
// 			// ImportState testing
// 			{
// 				ResourceName:      "commonfate_policyset.test",
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 			// Update and Read testing
// 			{
// 				Config: providerConfig + `

// 				resource "commonfate_policyset" "test" {
// 					id     = "test-updated"
// 					text = "test-updated"
// 				  }

// 						`,
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					// Verify first order item updated
// 					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "id", "test-updated"),
// 					resource.TestCheckResourceAttr("commonfate_access_workflow.test", "text", "test-updated"),

// 					// Verify dynamic values have any value set in the state.
// 					// resource.TestCheckResourceAttrSet("commonfate_access_workflow.test", "id"),
// 				),
// 			},
// 			// Delete testing automatically occurs in TestCase
// 		},
// 	})
// }
