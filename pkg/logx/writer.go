package logx

import "os"

type stdoutWriter struct{}

func (stdoutWriter) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}
