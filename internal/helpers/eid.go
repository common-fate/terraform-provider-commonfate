package helpers

import (
	entityv1alpha1 "github.com/common-fate/sdk/gen/commonfate/entity/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var EIDAttrs = map[string]schema.Attribute{
	"type": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The entity type",
	},
	"id": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The entity ID",
	},
}

type EID struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

func (u EID) ToAPI() *entityv1alpha1.EID {
	return &entityv1alpha1.EID{
		Type: u.Type.ValueString(),
		Id:   u.ID.ValueString(),
	}
}

func UidFromAPI(input *entityv1alpha1.EID) EID {
	if input == nil {
		return EID{}
	}

	return EID{
		Type: types.StringValue(input.Type),
		ID:   types.StringValue(input.Id),
	}
}

func UidPtrFromAPI(input *entityv1alpha1.EID) *EID {
	if input == nil {
		return nil
	}

	return &EID{
		Type: types.StringValue(input.Type),
		ID:   types.StringValue(input.Id),
	}
}
