package gcp

// func TestAccAccessSelector(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			// Create and Read testing
// 			{

// 				Config: providerConfig + `

// 				resource "commonfate_access_selector" "test" {
// 					workflow_id   = "test"
// 					name          = "test"
// 					selector_type = "GCP"
// 					role          = "test"

// 					targets = [
// 					{ type = "test", name = "test" },
// 					{ type = "test", name = "test" },
// 					]

// 				}

// `,
// 				Check: resource.ComposeAggregateTestCheckFunc(

// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "workflow_id", "test"),
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "name", "test"),
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "selector_type", "GCP"),
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "role", "test"),

// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "targets[0].type", "test"),
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "targets[0].name", "test"),

// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "targets[1].type", "test"),
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "targets[1].name", "test"),

// 					// Verify dynamic values have any value set in the state.
// 					resource.TestCheckResourceAttrSet("commonfate_access_Selector.test", "id"),
// 				),
// 			},
// 			// ImportState testing
// 			// {
// 			// 	ResourceName:      "commonfate_access_Selector.test",
// 			// 	ImportState:       true,
// 			// 	ImportStateVerify: true,
// 			// },
// 			// Update and Read testing
// 			{
// 				Config: providerConfig + `

// 				resource "commonfate_access_selector" "test" {
// 					workflow_id   = "test-updated"
// 					name          = "test-updated"
// 					selector_type = "AWS"
// 					role          = "test-updated"

// 					targets = [
// 					{ type = "test-updated", name = "test-updated" },
// 					{ type = "test-updated", name = "test-updated" },
// 					]

// 				}

// 						`,
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					// Verify first order item updated
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "workflow_id", "test-updated"),
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "name", "test-updated"),
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "selector_type", "AWS"),
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "role", "test-updated"),

// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "targets[0].type", "test-updated"),
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "targets[0].name", "test-updated"),

// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "targets[1].type", "test-updated"),
// 					resource.TestCheckResourceAttr("commonfate_access_Selector.test", "targets[1].name", "test-updated"),

// 					// Verify dynamic values have any value set in the state.
// 					// resource.TestCheckResourceAttrSet("commonfate_access_Selector.test", "id"),
// 				),
// 			},
// 			// Delete testing automatically occurs in TestCase
// 		},
// 	})
// }
