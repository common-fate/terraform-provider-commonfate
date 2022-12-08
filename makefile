provider:
	rm ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64/terraform-provider-commonfate
	go build -o terraform-provider-example
	cp terraform-provider-commonfate ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64
	chmod +x ~/.terraform.d/plugins/commonfate.com/commonfate/commonfate/1.0.0/darwin_amd64/terraform-provider-commonfate