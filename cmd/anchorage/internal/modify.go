package internal

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
)

func FindEditor() string {
	return findEditor([]string{"vim", "vi", "nano", "emacs"})
}

func findEditor(editors []string) string {
	for _, editor := range editors {
		_, err := exec.LookPath(editor)
		if err == nil {
			return editor
		}
	}
	return ""
}

var ErrNotFindEditor = errors.New("not found enable editor")
var ErrNoChanges = errors.New("no changes")

func EditTempFile(ctx context.Context, b []byte) ([]byte, error) {
	editor := FindEditor()
	if editor == "" {
		return nil, ErrNotFindEditor
	}
	temp, err := os.CreateTemp("", "*.yaml")
	if err != nil {
		return nil, err
	}
	defer os.Remove(temp.Name())
	_, err = temp.Write(b)
	if err != nil {
		return nil, err
	}
	_ = temp.Close()
	cmdCtx := exec.CommandContext(ctx, editor, temp.Name())
	cmdCtx.Stdout = os.Stdout
	cmdCtx.Stderr = os.Stderr
	cmdCtx.Stdin = os.Stdin
	err = cmdCtx.Run()
	if err != nil {
		return nil, err
	}
	file, err := os.ReadFile(temp.Name())
	if err != nil {
		return nil, err
	}
	if bytes.Equal(b, file) {
		return nil, ErrNoChanges
	}
	return file, nil
}
