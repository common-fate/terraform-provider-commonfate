---
page_title: "Provider: Common Fate"
description: |-
  The Common Fate provider is used to manage access to resources.
---

# Common Fate Provider

Use the Common Fate provider to interact with your access rules managed under Common Fate. 

For example:
```terraform
terraform {
  required_providers {
    commonfate = {
      source = "commonfate.com/commonfate/commonfate"
      version = "1.0.0"

    }
  }
}

provider "commonfate" {
  host = "https://example-commonfate.com"
  
}
```

You must have exported valid AWS credentials before you will be able to create access rules using the Terraform provider.

# Authorization and Configuration
To allow Common Fate to make access rules it will need to communicate with the deployed developer API for your Common Fate instance. 

This API is an AWS API Gateway with IAM Authorisation. Thus to communicate with it you will need to have active credentials when interacting with Common Fate via Terraform.

### Creating a Invoke API Policy

To complete the step below you will need to create a policy that allows `execute-api:Invoke` then you will need to assume that role and export the credentials to your local environment.

1. Go to the AWS console in the account Common Fate is deployed to and go to Identity and Access Management (IAM)
2. Go to Policies and click “Create Policy” 
3. Then click on the JSON tab
4. Copy in the following Policy
```
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
    - It should look like this: [https://dfksjgvbee.execute-api.ap-southeast-2.amazonaws.com/pro](https://f6tj3cg8rf.execute-api.ap-southeast-2.amazonaws.com/prod/gov/v1)d
    - Extract the “dfksjgvbee” from the URL, this will be the API Gateway ID
- Click next, and then next again
- Lastly give the policy a name and click “Create policy”

Add the policy to a permission set or role and get credentials for the policy then you will be able to create access rules with Terraform. See below on how to authenticate with AWS and use the Common Fate Terraform provider

### Exporting Credentials

Right now the only way to pass credentials to terraform is through environment variables. More methods of auth will come at a later date.

Credentials must be provided by using the AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and optionally AWS_SESSION_TOKEN environment variables. The region can be set using the AWS_REGION or AWS_DEFAULT_REGION environment variables.

For example:

```json

provider "commonfate" {
	host = "https://flkdj3s9fs.execute-api.ap-southeast-2.amazonaws.com/prod"
}
$ export AWS_ACCESS_KEY_ID="anaccesskey"
$ export AWS_SECRET_ACCESS_KEY="asecretkey"
$ export AWS_REGION="us-west-2"
$ terraform plan

```

Alternatively and a more preferred method of exporting credentials is to use [Granted](https://granted.dev/). Which will automatically create credentials for a given role and export them to your environment.

```json
provider "commonfate" {
	host = "https://flkdj3s9fs.execute-api.ap-southeast-2.amazonaws.com/prod"
}
$ assume cf-deployment-terraform
$ terraform plan
```

Now that you have set up your auth with the governance API, we can run through the demo Terraform provider.

## Resource "Access Rule"
With the Common Fate Terraform module you will be able to create and manage access rules, a commonly used resource in Common Fate. 
Access rules control who can request access to what, and the requirements surrounding their requests.

For example:
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