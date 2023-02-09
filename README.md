# Common Fate Terraform Provider

Common Fate Terraform provider using Terraform plugin framework

## Setting up the provider locally

First create a terraformrc file in your home directory by running `touch ~/.terraformrc` and add the following code

- If you already have a terraformrc this step can be skipped

Then open the file just created in your editor of choice and add the following

```
plugin_cache_dir   = "$HOME/.terraform.d/plugin-cache"
disable_checkpoint = true

provider_installation {
  dev_overrides {
                "registry.terraform.io/common-fate/commonfate" = "/Users/PATH_TO_GITHUB_REPO/common-fate-terraform-proto"
 }

  direct {}
}
```
- This tells Terraform to use a local copy of the provider over the deployed instance running at `registry.terraform.io`

save and exit.

Then build the provider by running `make provider`
- This will create the binary of the provider in your local directory.

# Running the example Terraform

`cd` into the `/examples` folder and there are example terraform files that can be used for testing, make sure to `cd` into the example folder.

You should now be able to run `terraform init` and `terraform plan` within the example folders

In the plugins repo.
- Run `make provider` to build a new version of the provider

## Tests

Terraform's official documentation for tests can be found here: https://developer.hashicorp.com/terraform/plugin/sdkv2/testing/acceptance-tests

Run the following to run the acceptance tests:

```
TF_ACC=1 go test -v ./...
```

## Docs

### Updating docs

To update the docs, edit the [template file](./templates/index.md.tmpl). To update and generate the docs, follow the instructions on [generating docs](#generating-docs)

### Generating docs

Run the following command:

```
make generate
```

The updated docs can be found in the [docs file](./docs/index.md)
