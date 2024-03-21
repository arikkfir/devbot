package strings

import (
	"bytes"
	"context"
	"github.com/secureworks/errors"
	"io"
	"time"
)

func AsyncReadLinesFromBuffer(ctx context.Context, i *bytes.Buffer, s chan string) error {
	var partialLine string
	sync := func() error {
		buffer := make([]byte, 1024)
		bytesRead, err := i.Read(buffer)
		if bytesRead > 0 {
			line := partialLine
			for i := 0; i < bytesRead; i++ {
				if buffer[i] == '\n' {
					s <- line
					line = ""
				} else {
					line += string(buffer[i])
				}
			}
			partialLine = line
		}
		return err
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			var errToReturn error
			for {
				if err := sync(); err != nil {
					if !errors.Is(err, io.EOF) {
						errToReturn = err
					}
					break
				}
			}
			if partialLine != "" {
				s <- partialLine
			}
			return errToReturn
		case <-ticker.C:
			for {
				if err := sync(); err != nil {
					if !errors.Is(err, io.EOF) {
						return err
					}
					break
				}
			}
		}
	}
}
