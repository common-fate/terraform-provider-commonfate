# Common Fate Terraform Provider
Common Fate Terraform provider using Terraform plugin framework


## Setting up the provider locally

First create a terraformrc file in your home directory by running `touch ~/.terraformrc` and add the following code
- If you already have a terraformrc this step can be skipped

Then open the file just created in your editor of choice and add the following

```
plugin_cache_dir   = "$HOME/.terraform.d/plugin-cache"
disable_checkpoint = true
```
save and exit.

Next we will make a local directory for the provider to live. To do this run: 
```
mkdir -p ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64
```

In the root of the repo;

Then build the provider by running `make provider` 
- This will copy the terraform binary into the local directory we made above.

# Running the example Terraform
`cd` into the `/examples` folder and there are example terraform files that can be used for testing, make sure to `cd` into the example folder.

You should now be able to run `terraform init` and `terraform plan` within the example folders

- If you get a permission denied error when `terraform plan` you will need to make sure the binary is executable with:

```
chmod +x ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64/terraform-provider-commonfate
```
In the plugins repo.


- There is a `make clean` command that will reset the terraform state if it ever gets in a broken state
- Run `make provider` to build a new version of the provider
- Running `make all` will run both a `make clean` and a `make provider`

## Tests
Terraform's official documentation for tests can be found here: https://developer.hashicorp.com/terraform/plugin/sdkv2/testing/acceptance-tests

Run the following to run the acceptance tests:
```
TF_ACC=1 go test -v ./... 
```


## Generating Docs
Run the following command: 
```
tfplugindocs generate --rendered-provider-name="CommonFate" --provider-name="commonfate"  --examples-dir="examples" --website-source-dir="templates"
```