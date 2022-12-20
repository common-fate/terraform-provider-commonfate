provider:
	# rm ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64/terraform-provider-commonfate
	go build -o terraform-provider-commonfate
	cp terraform-provider-commonfate ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64
	chmod +x ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64/terraform-provider-commonfate

clean: 
	rm -rf examples/verify/.terraform
	rm examples/verify/.terraform.lock.hcl
	rm examples/verify/terraform.tfstate
	rm examples/verify/terraform.tfstate.backup


all:
	make clean
	make provider