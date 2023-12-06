resource "common-fate_access_policy" "policy-1" {
  cedar = <<EOH
    permit(
      principal == User::"jane@acme.com",
      action == Action::"GCP::AutoApproval::roles/accessapproval.approver",
      resource == GCP::Project::"dev"
    )
EOH
}
