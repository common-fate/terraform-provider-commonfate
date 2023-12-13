resource "commonfate-policyset" "policy-1" {
  id   = "demo"
  text = <<EOH
    permit(
      principal,
      action == Access::Action::"Request",
      resource
    );
EOH
}
