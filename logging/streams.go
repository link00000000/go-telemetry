package logging

import (
	"bufio"
	"fmt"
	"io"
)

func (logger *Logger) LogReader(reader io.Reader, level Level, format string, args ...any) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}

		logger.Log(level, fmt.Sprintf(format, scanner.Text()), args...)
	}

	return nil
}
