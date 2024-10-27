package clipboard

import "time"

type (
	Msg struct {
		Format   Format
		Payload  string
		CopiedBy string
		CopiedAt time.Time
	}
	Format string
)

const (
	FormatUnknown Format = "unknown"
	FormatText           = "text"
	FormatImage          = "image"
)

func (msg1 Msg) ContentEqual(msg2 Msg) bool {
	return msg1.Format == msg2.Format && msg1.Payload == msg2.Payload
}

var LastClipboardContent = Msg{
	Format:   FormatUnknown,
	Payload:  "",
	CopiedBy: "",
	CopiedAt: time.Now(),
}
