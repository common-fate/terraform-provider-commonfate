# common-fate-terraform-proto
Common Fate Terraform provider using Terraform plugin framework


## Setting up the provider locally
First head into `~/.terraformrc` and add the following code

```
plugin_cache_dir   = "$HOME/.terraform.d/plugin-cache"
disable_checkpoint = true
```
save and exit 

In the root of the repo;

Then build the provider by running `go build -o terraform-provider-example` 
- This creates the binary of the provider and saves it to the current directory 

Then we want to copy that binary into a special plugins folder so that terriform knows where to find it when initialising a terraform tile

From your home directory create the following path using this command:
```
mkdir -p .terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64
```

Once the folder structure has been created (note this is for osx exclusive, change the trailing os type for windows/linux)

We want to copy the binary into the folder we created. With this command: 

```
cp terraform-provider-commonfate ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64
```

You should now be able to run `terraform init` and `terraform plan` on the examples

- If you get a permission denied error when `terraform plan` you will need to make sure the binary is executable with:
```
chmod +x ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64/terraform-provider-commonfate
```
In the plugins repo.