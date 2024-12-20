---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commonfate_proxy_rds_database Resource - commonfate"
subcategory: ""
description: |-
  Registers a RDS database with a Common Fate Proxy.
---

# commonfate_proxy_rds_database (Resource)

Registers a RDS database with a Common Fate Proxy.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `aws_account_id` (String) The AWS account id the database is in
- `database` (String) The name of the database
- `endpoint` (String) The endpoint of the database
- `engine` (String) A SQL engine of the RDS database
- `instance_id` (String) The name of the parent instance id that the database will be connected to
- `name` (String) A unique name for the resource
- `proxy_id` (String) The ID of the proxy in the same account as the database.
- `region` (String) The region the database is in
- `users` (Attributes List) A list of users that exist in the database (see [below for nested schema](#nestedatt--users))

### Read-Only

- `id` (String) The internal resource

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Required:

- `name` (String) The name for the user
- `password_secrets_manager_arn` (String) The secrets manager arn for the RDS database password
- `username` (String) The user name for the user

Optional:

- `default_local_port` (Number) The default local port to use for the user when running `granted rds proxy`
- `endpoint` (String) Override default endpoint behaviour by specifying a endpoint on a per user basis.


