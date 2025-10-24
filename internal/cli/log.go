package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Tail log file",
	Run: func(cmd *cobra.Command, args []string) {
		logFile := os.Getenv("HOME") + "/.local/share/wgo/log/wgo.log"
		if logFile == "" {
			fmt.Println("Log file not configured")
			return
		}

		file, err := os.Open(logFile)
		if err != nil {
			fmt.Printf("Failed to read log file: %s\n", err)
			return
		}

		const linesToTail = 10
		info, err := file.Stat()
		if err != nil {
			fmt.Printf("Failed to read log file: %s\n", err)
			_ = file.Close()
			return
		}

		size := info.Size()
		seekOffset := size
		var initialOutput []byte
		if size > 0 {
			const chunkSize int64 = 4096
			var buffer []byte
			linesFound := 0
			pos := size
			for pos > 0 && linesFound <= linesToTail {
				readSize := chunkSize
				if pos < chunkSize {
					readSize = pos
				}
				pos -= readSize
				chunk := make([]byte, readSize)
				if _, err := file.ReadAt(chunk, pos); err != nil && err != io.EOF {
					fmt.Printf("Failed to read log file: %s\n", err)
					_ = file.Close()
					return
				}
				buffer = append(chunk, buffer...)
				linesFound += bytes.Count(chunk, []byte{'\n'})
			}

			if len(buffer) == 0 {
				initialOutput = nil
			} else {
				startIdx := 0
				if linesFound > linesToTail {
					skip := linesFound - linesToTail
					for skip > 0 {
						next := bytes.IndexByte(buffer[startIdx:], '\n')
						if next == -1 {
							startIdx = len(buffer)
							break
						}
						startIdx += next + 1
						skip--
					}
				}
				seekOffset = size
				initialOutput = buffer[startIdx:]
			}
		}

		if err := file.Close(); err != nil {
			fmt.Printf("Failed to close log file: %s\n", err)
			return
		}

		if len(initialOutput) > 0 {
			fmt.Print(string(initialOutput))
		}

		t, err := tail.TailFile(logFile, tail.Config{
			Follow:   true,
			ReOpen:   true,
			Location: &tail.SeekInfo{Offset: seekOffset, Whence: io.SeekStart},
		})
		if err != nil {
			fmt.Printf("Failed to tail log file: %s\n", err)
			return
		}

		for line := range t.Lines {
			fmt.Println(line.Text)
		}
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
