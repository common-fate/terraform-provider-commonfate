resource "commonfate-policyset" "policy-1" {
  id="demo"
  text = <<EOH
    permit(
      principal == User::"jane@acme.com",
      action == Action::"GCP::AutoApproval::roles/accessapproval.approver",
      resource == GCP::Project::"dev"
    )
EOH
}
