package codex

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

const (
	// InputTypeText represents a plain text input.
	InputTypeText = "text"
	// InputTypeImage represents a remote image input.
	InputTypeImage = "image"
	// InputTypeLocalImage represents a local image input.
	InputTypeLocalImage = "localImage"
	// InputTypeAudio represents a remote audio input.
	InputTypeAudio = "audio"
	// InputTypeLocalAudio represents a local audio input.
	InputTypeLocalAudio = "localAudio"
	// InputTypeSkill represents a skill invocation input.
	InputTypeSkill = "skill"
)

// Input represents a structured user input message.
type Input struct {
	// Type must be one of the InputType* constants.
	Type         string                 `json:"type"`
	Text         string                 `json:"text,omitempty"`
	TextElements []protocol.TextElement `json:"text_elements,omitempty"`
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

// AudioInput creates a remote audio input entry.
func AudioInput(url string) Input {
	return Input{Type: InputTypeAudio, URL: url}
}

// LocalAudioInput creates a local audio input entry.
func LocalAudioInput(path string) Input {
	return Input{Type: InputTypeLocalAudio, Path: path}
}

// SkillInput creates a skill input entry.
func SkillInput(name, path string) Input {
	return Input{Type: InputTypeSkill, Name: name, Path: path}
}

// MentionInput creates a text input containing a single mention placeholder.
func MentionInput(name string) Input {
	text := "@" + name
	return Input{
		Type: InputTypeText,
		Text: text,
		TextElements: []protocol.TextElement{{
			ByteRange:   protocol.TextElementByteRange{Start: 0, End: len(text)},
			Placeholder: stringPtr(name),
		}},
	}
}

func (i Input) validate() error {
	switch i.Type {
	case InputTypeText:
		if i.Text == "" && len(i.TextElements) == 0 {
			return errors.New("text input is empty")
		}
		for _, elem := range i.TextElements {
			if elem.ByteRange.Start < 0 || elem.ByteRange.End < elem.ByteRange.Start || elem.ByteRange.End > len(i.Text) {
				return errors.New("text element byte range is invalid")
			}
			if elem.Placeholder != nil && *elem.Placeholder == "" {
				return errors.New("text element placeholder is empty")
			}
		}
		if !utf8.ValidString(i.Text) {
			return errors.New("text input is not valid utf-8")
		}
	case InputTypeImage:
		if i.URL == "" {
			return errors.New("image input URL is empty")
		}
	case InputTypeLocalImage:
		if i.Path == "" {
			return errors.New("local image input path is empty")
		}
	case InputTypeAudio:
		if i.URL == "" {
			return errors.New("audio input URL is empty")
		}
	case InputTypeLocalAudio:
		if i.Path == "" {
			return errors.New("local audio input path is empty")
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
