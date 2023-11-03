package version

import (
	"context"
	"testing"

	"github.com/jannfis/argocd-agent/internal/version"
	"github.com/jannfis/argocd-agent/pkg/api/grpc/versionapi"
	"github.com/stretchr/testify/assert"
)

func Test_Version(t *testing.T) {
	t.Run("Get version identifier", func(t *testing.T) {
		s := NewServer(nil)
		r, err := s.Version(context.Background(), &versionapi.VersionRequest{})
		assert.NoError(t, err)
		assert.Equal(t, version.QualifiedVersion(), r.Version)
	})
}
