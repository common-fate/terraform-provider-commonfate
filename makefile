provider:
	go build -o terraform-provider-commonfate


generate:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --rendered-provider-name="Common Fate" --provider-name="commonfate"  --examples-dir="examples" --website-source-dir="templates"

all:
	make clean
	make provider