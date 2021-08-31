package resolvers

import (
	"context"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
)

type SiteInfoResolver struct {
	*Resolver
}

func (r *Resolver) SiteInfo() api.SiteInfoResolver {
	return &SiteInfoResolver{r}
}

func (SiteInfoResolver) FaceDetectionEnabled(ctx context.Context, obj *models.SiteInfo) (bool, error) {
	return face_detection.GlobalFaceDetector != nil, nil
}
