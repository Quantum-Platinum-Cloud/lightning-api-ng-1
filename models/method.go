package models

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/lightninglabs/lightning-api-ng/defs"
	"github.com/lightninglabs/lightning-api-ng/markdown"
	"golang.org/x/exp/slices"
)

type Method struct {
	Service *Service

	Name               string
	Description        string
	Source             string
	CommandLine        string
	CommandLineHelp    string
	RequestType        string
	RequestFullType    string
	RequestTypeSource  string
	RequestStreaming   bool
	ResponseType       string
	ResponseFullType   string
	ResponseTypeSource string
	ResponseStreaming  bool
	RestMapping        *RestMapping
	CodeSamples        *CodeSamples

	// These values are set after the method is created. They are simply
	// used to avoid having to look up the messages for each method call.
	request        *Message
	response       *Message
	nestedMessages []*Message
	nestedEnums    []*Enum
}

// NewMethod creates a new method from a method definition.
func NewMethod(methodDef *defs.ServiceMethod, service *Service) *Method {
	m := &Method{
		Service:     service,
		Name:        methodDef.Name,
		Description: parseDescription(methodDef.Description),
		Source:      methodDef.Source,
		CommandLine: methodDef.CommandLine,
		CommandLineHelp: markdown.CleanDescription(
			methodDef.CommandLineHelp, false,
		),
		RequestType:        methodDef.RequestType,
		RequestFullType:    methodDef.RequestFullType,
		RequestTypeSource:  methodDef.RequestTypeSource,
		RequestStreaming:   methodDef.RequestStreaming,
		ResponseType:       methodDef.ResponseType,
		ResponseFullType:   methodDef.ResponseFullType,
		ResponseTypeSource: methodDef.ResponseTypeSource,
		ResponseStreaming:  methodDef.ResponseStreaming,
	}

	if (methodDef.RESTMappings != nil) &&
		(len(methodDef.RESTMappings) > 0) {

		m.RestMapping = NewRestMapping(*methodDef.RESTMappings[0])
	}

	m.CodeSamples = NewCodeSamples(m)

	return m
}

// Request returns the request message instance using the request full type.
func (m *Method) Request() *Message {
	if m.request == nil {
		msg, err := m.Service.Pkg.App.GetMessage(m.RequestFullType)
		if err != nil {
			panic(fmt.Sprintf("Error getting message %s for %s: %s",
				m.RequestFullType, m.Name, err))
		}
		msg.Source = m.RequestTypeSource
		m.request = msg
	}
	return m.request
}

// Response returns the request message instance using the request full type.
func (m *Method) Response() *Message {
	if m.response == nil {
		msg, err := m.Service.Pkg.App.GetMessage(m.ResponseFullType)
		if err != nil {
			panic(fmt.Sprintf("Error getting message %s for %s: %s",
				m.ResponseFullType, m.Name, err))
		}
		msg.Source = m.ResponseTypeSource
		m.response = msg
	}
	return m.response
}

// NestedMessages returns a slice of all the nested messages for this method.
func (m *Method) NestedMessages() []*Message {
	if m.nestedMessages == nil {
		// Create a new map and populated it with the nested messages.
		messages := make(map[string]*Message)
		m.Service.Pkg.App.GetNestedMessages(m.Request(), messages, 10)
		m.Service.Pkg.App.GetNestedMessages(m.Response(), messages, 10)

		// Convert the map to a slice.
		m.nestedMessages = make([]*Message, 0, len(messages))
		for _, msg := range messages {
			m.nestedMessages = append(m.nestedMessages, msg)
		}
		slices.SortFunc(m.nestedMessages, func(i, j *Message) bool {
			return strings.ToLower(i.FullName) <
				strings.ToLower(j.FullName)
		})
	}
	return m.nestedMessages
}

// NestedEnums returns a slice of all the nested enums for this method.
func (m *Method) NestedEnums() []*Enum {
	if m.nestedEnums == nil {
		// Create a new map and populated it with the nested enums.
		enums := make(map[string]*Enum)
		m.Service.Pkg.App.GetNestedEnums(m.Request(), enums, 10)
		m.Service.Pkg.App.GetNestedEnums(m.Response(), enums, 10)

		// Convert the map to a slice.
		m.nestedEnums = make([]*Enum, 0, len(enums))
		for _, enum := range enums {
			m.nestedEnums = append(m.nestedEnums, enum)
		}
		slices.SortFunc(m.nestedEnums, func(i, j *Enum) bool {
			return strings.ToLower(i.FullName) <
				strings.ToLower(j.FullName)
		})
	}
	return m.nestedEnums
}

// IsDeprecated returns true if the method contains the word "deprecated" in
// the description.
func (m *Method) IsDeprecated() bool {
	return strings.Contains(strings.ToLower((m.Description)), "deprecated")
}

// HasNestedMessages returns true if the method has nested messages.
func (m *Method) HasNestedMessages() bool {
	return len(m.NestedMessages()) > 0
}

// HasNestedEnums returns true if the method has nested enums.
func (m *Method) HasNestedEnums() bool {
	return len(m.NestedEnums()) > 0
}

// HasRestMapping returns true if the method has a REST mapping.
func (m *Method) HasRestMapping() bool {
	return m.RestMapping != nil && m.RestMapping.Path != ""
}

// RestMethod returns the REST method of the method.
func (m *Method) RestMethod() string {
	if m.HasRestMapping() {
		return m.RestMapping.Method
	}
	return ""
}

// RestPath returns the REST path of the method.
func (m *Method) RestPath() string {
	if m.HasRestMapping() {
		return m.RestMapping.Path
	}
	return ""
}

// StreamingDirection returns the streaming direction of the method.
func (m *Method) StreamingDirection() string {
	switch {
	case m.RequestStreaming && m.ResponseStreaming:
		return "bidirectional"
	case m.ResponseStreaming:
		return "server"
	case m.RequestStreaming:
		return "client"
	default:
		return ""
	}
}

// ExportMarkdown exports the method to a markdown file.
func (m *Method) ExportMarkdown(servicePath string) error {
	// update the rest types & placement before rendering to markdown
	// because there may be multiple methods with the same request type
	if m.RestMapping != nil {
		m.RestMapping.UpdateMessage(m.Request())
	}
	fileName := strcase.ToKebab(m.Name)
	filePath := fmt.Sprintf("%s/%s.mdx", servicePath, fileName)
	fmt.Printf("Exporting method %s to %s\n", m.Name, filePath)

	// execute the template for the method
	filePath = fmt.Sprintf("%s/%s.mdx", servicePath,
		markdown.ToKebabCase(m.Name))

	err := markdown.ExecuteMethodTemplate(m.Service.Pkg.App.Templates, m,
		filePath)

	if err != nil {
		return err
	}

	return nil
}

// parseDescription removes the first line from the description if it contains
// the CLI command.
func parseDescription(description string) string {
	if description == "" {
		return ""
	}

	lines := strings.Split(description, "\n")
	if strings.Contains(lines[0], ": `") {
		// If the first line looks like "lncli: `closechannel`", it is
		// a command, so skip it.
		return strings.Join(lines[1:], "\n")
	}

	return description
}
