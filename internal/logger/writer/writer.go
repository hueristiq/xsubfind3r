package writer

import "github.com/hueristiq/xsubfind3r/internal/logger/levels"

type Writer interface {
	Write(data []byte, level levels.Level)
}
