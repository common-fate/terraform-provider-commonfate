---
page_title: "Provider: Common Fate"
description: |-
  The Common Fate provider is used to configure and manage access to your cloud and critical applications.
---

# Common Fate Provider

Use the Common Fate provider to interact with your Access Rules managed under [Common Fate](https://docs.commonfate.io/common-fate/introduction).

For example:

```terraform
terraform {
  required_providers {
    commonfate = {
      source = "common-fate/commonfate"
      version = "~> 1.0"
    }
  }
}

provider "commonfate" {
  governance_api_url = "https://commonfate-api.example.com"
}
```

## Prerequisites

To utilise the Common Fate Provider, you must have the following:

- A valid Common Fate deployment
- Exported valid AWS credentials to enable you to create Access Rules using the Terraform provider. To create a permission set or role with the appropriate permissions, see the section on [Authorization and Configurtion](#authorization-and-configuration). To understand how to appropriately export credentials, see [Exporting Credentials](#exporting-credentials)

## Configuration

To enable the connection of your Terraform and your Common Fate deployment, you must identify the host within your provider. To achieve this, ensure you have exported valid AWS credentials to your terminal window. You are then required to run the following command in the root of your Common Fate deployment:

```bash
gdeploy status
```

Within the returned table will be a `GovernanceURL` credential, it will look something like this: [https://dfksjgvbee.execute-api.ap-southeast-2.amazonaws.com/prod/](https://dfksjgvbee.execute-api.ap-southeast-2.amazonaws.com/prod/). Copy this value and assign it to the host within your Provider. Below is an example:

```terraform
provider "commonfate" {
  governance_api_url = "https://yfbttt8s59.execute-api.ap-southeast-2.amazonaws.com/prod/"
}
```

Once you have completed authentication with the governance API, you can run through the demo Terraform provider.

## The "Access Rule" Resource

With the Common Fate Terraform module you will be able to create and manage Access Rules, a commonly used resource in Common Fate.
Access Rules control who can request access to what, and the requirements surrounding their requests.

Below is a sample Terraform file. This code snippet demonstrates creating an access rule called "Common Fate Access Rule", allowing anyone in the "common_fate_administrators" group to assume it, without requiring approval.

```terraform
resource "commonfate_access_rule" "aws-admin" {
  name ="This was made in terraform demo"
  description="Access rule made in terraform"
  groups=["common_fate_administrators"]
  approval= {
      users=[""]
  }

  target=[
    {
      field="accountId"
      value=["123456789012"]
    },
    {
      field="permissionSetArn"
      value=["arn:aws:sso:::permissionSet/ssoins-825968feece9a0b6/3hjdfkj3r28ef"]
    }
  ]
  target_provider_id="aws-sso-v2"
  duration="3600"
}
```

To apply your changes, ensure you have exported the appropriate credentials to your terminal window, then run the standard Terraform command:

```bash
terraform apply
```

## Authorization

To allow Common Fate to make Access Rules it will need to communicate with the deployed developer API for your Common Fate instance.

This API is an AWS API Gateway with IAM Authorisation. Thus to communicate with it you will need to have active credentials when interacting with Common Fate via Terraform.

### Creating a Invoke API Policy

To complete the step below you will need to create a policy that allows `execute-api:Invoke` then you will need to assume that role and export the credentials to your local environment.

1. Go to the AWS console in the account Common Fate is deployed to and go to Identity and Access Management (IAM)
2. Go to Policies and click “Create Policy”
3. Then click on the JSON tab
4. Copy in the following Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "VisualEditor0",
      "Effect": "Allow",
      "Action": "execute-api:Invoke",
      "Resource": "arn:aws:execute-api:{REGION}:{AWS_ACCOUNT}:{API_GATEWAYY_ID}/*/*/*"
    }
  ]
}
```

- API_GATEWAY_ID can be found by running `gdeploy status` and finding the `GovernanceURL` from the table
  - It should look like this: [https://dfksjgvbee.execute-api.ap-southeast-2.amazonaws.com/prod/](https://dfksjgvbee.execute-api.ap-southeast-2.amazonaws.com/prod/)
  - Extract the “dfksjgvbee” from the URL, this will be the API Gateway ID
- Click next, and then next again
- Lastly give the policy a name and click “Create policy”

Add the policy to a permission set or role and get credentials for the policy then you will be able to create Access Rules with Terraform. See below on how to authenticate with AWS and use the Common Fate Terraform provider

### Exporting Credentials

Right now the only way to pass credentials to terraform is through environment variables. More methods of auth will come at a later date.

Credentials must be provided by using the AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and optionally AWS_SESSION_TOKEN environment variables. The region can be set using the AWS_REGION or AWS_DEFAULT_REGION environment variables.

For example:

```
provider "commonfate" {
	governance_api_url = "https://flkdj3s9fs.execute-api.ap-southeast-2.amazonaws.com/prod"
}
$ export AWS_ACCESS_KEY_ID="anaccesskey"
$ export AWS_SECRET_ACCESS_KEY="asecretkey"
$ export AWS_REGION="us-west-2"
$ terraform plan
```

Alternatively and a more preferred method of exporting credentials is to use [Granted](https://granted.dev/). Granted will automatically create credentials for a given role and export them to your environment.

```
provider "commonfate" {
	governance_api_url = "https://flkdj3s9fs.execute-api.ap-southeast-2.amazonaws.com/prod"
}
$ assume cf-deployment-terraform
$ terraform plan
```
