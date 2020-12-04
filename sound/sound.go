package sound

import (
	"context"
)

type Output interface {
	Play(ctx context.Context, filePath string) error
}
