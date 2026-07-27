package logs

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func LogWriter(pathLog *os.File, errCh <-chan error) {
	writer := bufio.NewWriter(pathLog)
	defer writer.Flush()

	for err := range errCh {
		fmt.Fprintf(writer, "[%s] %v\n", time.Now().Format(time.RFC3339), err)
		writer.Flush()
	}
}
