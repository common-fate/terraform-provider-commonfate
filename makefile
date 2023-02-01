provider:
	#rm ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64/terraform-provider-commonfate
	go build -o terraform-provider-commonfate
	cp terraform-provider-commonfate ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64
	chmod +x ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64/terraform-provider-commonfate

clean: 
	# rm -rf examples/verify/.terraform
	# rm examples/verify/.terraform.lock.hcl
	# rm examples/verify/terraform.tfstate
	# rm examples/verify/terraform.tfstate.backup
	rm -rf examples/sample/.terraform
	rm examples/sample/.terraform.lock.hcl
	rm examples/sample/terraform.tfstate
	rm examples/sample/terraform.tfstate.backup

generate:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --rendered-provider-name="Common Fate" --provider-name="commonfate"  --examples-dir="examples" --website-source-dir="templates"

all:
	make clean
	make provider