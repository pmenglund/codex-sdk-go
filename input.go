package codex

import (
	"errors"
	"fmt"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

const (
	// InputTypeText represents a plain text input.
	InputTypeText = "text"
	// InputTypeImage represents a remote image input.
	InputTypeImage = "image"
	// InputTypeLocalImage represents a local image input.
	InputTypeLocalImage = "localImage"
	// InputTypeSkill represents a skill invocation input.
	InputTypeSkill = "skill"
)

// Input represents a structured user input message.
type Input struct {
	// Type must be one of the InputType* constants.
	Type         string                 `json:"type"`
	Text         string                 `json:"text,omitempty"`
	TextElements []protocol.TextElement `json:"textElements,omitempty"`
	URL          string                 `json:"url,omitempty"`
	Path         string                 `json:"path,omitempty"`
	Name         string                 `json:"name,omitempty"`
}

// TextInput creates a text input entry.
func TextInput(text string) Input {
	return Input{Type: InputTypeText, Text: text}
}

// ImageInput creates a remote image input entry.
func ImageInput(url string) Input {
	return Input{Type: InputTypeImage, URL: url}
}

// LocalImageInput creates a local image input entry.
func LocalImageInput(path string) Input {
	return Input{Type: InputTypeLocalImage, Path: path}
}

// SkillInput creates a skill input entry.
func SkillInput(name, path string) Input {
	return Input{Type: InputTypeSkill, Name: name, Path: path}
}

func (i Input) validate() error {
	switch i.Type {
	case InputTypeText:
		if i.Text == "" && len(i.TextElements) == 0 {
			return errors.New("text input is empty")
		}
	case InputTypeImage:
		if i.URL == "" {
			return errors.New("image input URL is empty")
		}
	case InputTypeLocalImage:
		if i.Path == "" {
			return errors.New("local image input path is empty")
		}
	case InputTypeSkill:
		if i.Name == "" {
			return errors.New("skill input name is empty")
		}
		if i.Path == "" {
			return errors.New("skill input path is empty")
		}
	default:
		return fmt.Errorf("unknown input type %q", i.Type)
	}
	return nil
}
