---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commonfate_aws_eks_access_entry_template Resource - commonfate"
subcategory: ""
description: |-
  Registers an AWS EKS Access Entry Template with Common Fate
---

# commonfate_aws_eks_access_entry_template (Resource)

Registers an AWS EKS Access Entry Template with Common Fate



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the EKS Access Entry Template

### Optional

- `cluster_access_policies` (Attributes List) The cluster access policies associated with the template (see [below for nested schema](#nestedatt--cluster_access_policies))
- `kubernetes_groups` (List of String) The Kubernetes groups associated with the template
- `namespace_access_policies` (Attributes List) The namespace access policies associated with the template (see [below for nested schema](#nestedatt--namespace_access_policies))
- `tags` (Map of String) The tags associated with the template

### Read-Only

- `id` (String) The EKS Access Entry Template ID

<a id="nestedatt--cluster_access_policies"></a>
### Nested Schema for `cluster_access_policies`

Required:

- `policy_arn` (String) The ARN of the cluster access policy


<a id="nestedatt--namespace_access_policies"></a>
### Nested Schema for `namespace_access_policies`

Required:

- `namespaces` (List of String) The namespaces associated with the policy
- `policy_arn` (String) The ARN of the namespace access policy

