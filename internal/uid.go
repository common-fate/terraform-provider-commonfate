package internal

import (
	entityv1alpha1 "github.com/common-fate/sdk/gen/commonfate/entity/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var UIDAttrs = map[string]schema.Attribute{
	"type": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The entity type",
	},
	"id": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The entity ID",
	},
}

type UID struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

func (u UID) ToAPI() *entityv1alpha1.UID {
	return &entityv1alpha1.UID{
		Type: u.Type.ValueString(),
		Id:   u.ID.ValueString(),
	}
}

func uidFromAPI(input *entityv1alpha1.UID) UID {
	if input == nil {
		return UID{}
	}

	return UID{
		Type: types.StringValue(input.Type),
		ID:   types.StringValue(input.Id),
	}
}

func uidPtrFromAPI(input *entityv1alpha1.UID) *UID {
	if input == nil {
		return nil
	}

	return &UID{
		Type: types.StringValue(input.Type),
		ID:   types.StringValue(input.Id),
	}
}
