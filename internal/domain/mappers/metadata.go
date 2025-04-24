package mappers

import (
	"gitlab.com/orkys/backend/gateway/internal/domain/valueobject"
	"gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/commons"
)

func MetadataToProto(metadata *valueobject.Metadata) *commons.Metadata {
	return &commons.Metadata{
		CreatedAt: metadata.CreatedAt,
		UpdatedAt: metadata.UpdatedAt,
		DeletedAt: metadata.DeletedAt,
		CreatedBy: metadata.CreatedBy,
		UpdatedBy: metadata.UpdatedBy,
		DeletedBy: metadata.DeletedBy,
	}
}

func MetadataFromProto(metadata *commons.Metadata) *valueobject.Metadata {
	return &valueobject.Metadata{
		CreatedAt: metadata.CreatedAt,
		UpdatedAt: metadata.UpdatedAt,
		DeletedAt: metadata.DeletedAt,
		CreatedBy: metadata.CreatedBy,
		UpdatedBy: metadata.UpdatedBy,
		DeletedBy: metadata.DeletedBy,
	}
}
