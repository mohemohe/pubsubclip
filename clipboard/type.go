package clipboard

import "time"

type (
	Msg struct {
		Format   Format
		Payload  string
		CopiedBy string
		CopiedAt time.Time
	}
	Format int
)

const (
	FormatUnknown Format = iota
	FormatText
	FormatImage
)
