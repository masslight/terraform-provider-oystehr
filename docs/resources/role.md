---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "oystehr_role Resource - Oystehr"
subcategory: ""
description: |-
  
---

# oystehr_role (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `access_policy` (Attributes) The access policy associated with the role. (see [below for nested schema](#nestedatt--access_policy))
- `name` (String) The name of the role.

### Optional

- `description` (String) A description of the role.

### Read-Only

- `id` (String) The ID of the role.

<a id="nestedatt--access_policy"></a>
### Nested Schema for `access_policy`

Optional:

- `rule` (Attributes List) A list of rules in the access policy. (see [below for nested schema](#nestedatt--access_policy--rule))

<a id="nestedatt--access_policy--rule"></a>
### Nested Schema for `access_policy.rule`

Required:

- `action` (List of String) The actions the rule allows or denies.
- `effect` (String) The effect of the rule (Allow or Deny).
- `resource` (List of String) The resources the rule applies to.

Optional:

- `condition` (Map of String) Conditions for the rule.
