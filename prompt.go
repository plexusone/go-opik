package opik

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/grokify/go-comet-ml-opik/internal/api"
)

// Prompt represents a prompt template in Opik.
type Prompt struct {
	client      *Client
	id          string
	name        string
	description string
	template    string
	tags        []string
}

// PromptVersion represents a specific version of a prompt.
type PromptVersion struct {
	client            *Client
	id                string
	promptID          string
	commit            string
	template          string
	changeDescription string
	tags              []string
}

// PromptTemplateStructure represents the type of prompt template.
type PromptTemplateStructure string

const (
	PromptTemplateStructureText PromptTemplateStructure = "text"
	PromptTemplateStructureChat PromptTemplateStructure = "chat"
)

// PromptType represents the type of prompt.
type PromptType string

const (
	PromptTypeMustache PromptType = "mustache"
	PromptTypeFString  PromptType = "fstring"
	PromptTypeJinja2   PromptType = "jinja2"
)

// ID returns the prompt ID.
func (p *Prompt) ID() string {
	return p.id
}

// Name returns the prompt name.
func (p *Prompt) Name() string {
	return p.name
}

// Description returns the prompt description.
func (p *Prompt) Description() string {
	return p.description
}

// Template returns the prompt template.
func (p *Prompt) Template() string {
	return p.template
}

// Tags returns the prompt tags.
func (p *Prompt) Tags() []string {
	return p.tags
}

// ID returns the prompt version ID.
func (v *PromptVersion) ID() string {
	return v.id
}

// PromptID returns the parent prompt ID.
func (v *PromptVersion) PromptID() string {
	return v.promptID
}

// Commit returns the version commit hash.
func (v *PromptVersion) Commit() string {
	return v.commit
}

// Template returns the version template.
func (v *PromptVersion) Template() string {
	return v.template
}

// ChangeDescription returns the version change description.
func (v *PromptVersion) ChangeDescription() string {
	return v.changeDescription
}

// Tags returns the version tags.
func (v *PromptVersion) Tags() []string {
	return v.tags
}

// PromptOption is a functional option for configuring a Prompt.
type PromptOption func(*promptOptions)

type promptOptions struct {
	description       string
	template          string
	changeDescription string
	promptType        PromptType
	templateStructure PromptTemplateStructure
	tags              []string
}

// WithPromptDescription sets the description for the prompt.
func WithPromptDescription(description string) PromptOption {
	return func(o *promptOptions) {
		o.description = description
	}
}

// WithPromptTemplate sets the template for the prompt.
func WithPromptTemplate(template string) PromptOption {
	return func(o *promptOptions) {
		o.template = template
	}
}

// WithPromptChangeDescription sets the change description for the prompt version.
func WithPromptChangeDescription(changeDescription string) PromptOption {
	return func(o *promptOptions) {
		o.changeDescription = changeDescription
	}
}

// WithPromptType sets the type for the prompt (mustache, fstring, jinja2).
func WithPromptType(promptType PromptType) PromptOption {
	return func(o *promptOptions) {
		o.promptType = promptType
	}
}

// WithPromptTemplateStructure sets the template structure (text or chat).
func WithPromptTemplateStructure(structure PromptTemplateStructure) PromptOption {
	return func(o *promptOptions) {
		o.templateStructure = structure
	}
}

// WithPromptTags sets the tags for the prompt.
func WithPromptTags(tags ...string) PromptOption {
	return func(o *promptOptions) {
		o.tags = tags
	}
}

// CreatePrompt creates a new prompt.
func (c *Client) CreatePrompt(ctx context.Context, name string, opts ...PromptOption) (*Prompt, error) {
	options := &promptOptions{
		tags:              []string{},
		templateStructure: PromptTemplateStructureText,
		promptType:        PromptTypeMustache,
	}
	for _, opt := range opts {
		opt(options)
	}

	promptUUID := uuid.New()

	req := api.PromptWrite{
		ID:                api.NewOptUUID(promptUUID),
		Name:              name,
		Description:       api.NewOptString(options.description),
		Template:          api.NewOptString(options.template),
		ChangeDescription: api.NewOptString(options.changeDescription),
		Type:              api.NewOptPromptWriteType(api.PromptWriteType(options.promptType)),
		TemplateStructure: api.NewOptPromptWriteTemplateStructure(api.PromptWriteTemplateStructure(options.templateStructure)),
		Tags:              options.tags,
	}

	resp, err := c.apiClient.CreatePrompt(ctx, api.NewOptPromptWrite(req))
	if err != nil {
		return nil, err
	}

	_ = resp // Response contains location header

	return &Prompt{
		client:      c,
		id:          promptUUID.String(),
		name:        name,
		description: options.description,
		template:    options.template,
		tags:        options.tags,
	}, nil
}

// GetPrompt retrieves a prompt by ID.
func (c *Client) GetPrompt(ctx context.Context, promptID string) (*Prompt, error) {
	promptUUID, err := uuid.Parse(promptID)
	if err != nil {
		return nil, err
	}

	resp, err := c.apiClient.GetPromptById(ctx, api.GetPromptByIdParams{ID: promptUUID})
	if err != nil {
		return nil, err
	}

	switch v := resp.(type) {
	case *api.PromptDetail:
		var id, description string
		if v.ID.Set {
			id = v.ID.Value.String()
		}
		if v.Description.Set {
			description = v.Description.Value
		}

		return &Prompt{
			client:      c,
			id:          id,
			name:        v.Name,
			description: description,
			tags:        v.Tags,
		}, nil
	default:
		return nil, ErrPromptNotFound
	}
}

// GetPromptByName retrieves a prompt version by name and optional commit.
func (c *Client) GetPromptByName(ctx context.Context, name string, commit string) (*PromptVersion, error) {
	req := api.PromptVersionRetrieveDetail{
		Name: name,
	}
	if commit != "" {
		req.Commit = api.NewOptString(commit)
	}

	resp, err := c.apiClient.RetrievePromptVersion(ctx, api.NewOptPromptVersionRetrieveDetail(req))
	if err != nil {
		return nil, err
	}

	switch v := resp.(type) {
	case *api.PromptVersionDetail:
		var id, promptID, commitVal, changeDescription string
		if v.ID.Set {
			id = v.ID.Value.String()
		}
		if v.PromptID.Set {
			promptID = v.PromptID.Value.String()
		}
		if v.Commit.Set {
			commitVal = v.Commit.Value
		}
		if v.ChangeDescription.Set {
			changeDescription = v.ChangeDescription.Value
		}

		return &PromptVersion{
			client:            c,
			id:                id,
			promptID:          promptID,
			commit:            commitVal,
			template:          v.Template,
			changeDescription: changeDescription,
			tags:              v.Tags,
		}, nil
	default:
		return nil, ErrPromptNotFound
	}
}

// ListPrompts lists all prompts.
//
//nolint:dupl // Similar structure to ListDatasets is intentional for consistency
func (c *Client) ListPrompts(ctx context.Context, page, size int) ([]*Prompt, error) {
	params := api.GetPromptsParams{
		Page: api.NewOptInt32(int32(page)), //nolint:gosec // G115: page values are bounded by API limits
		Size: api.NewOptInt32(int32(size)), //nolint:gosec // G115: size values are bounded by API limits
	}

	resp, err := c.apiClient.GetPrompts(ctx, params)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return []*Prompt{}, nil
	}

	prompts := make([]*Prompt, 0, len(resp.Content))
	for _, p := range resp.Content {
		var id, description string
		if p.ID.Set {
			id = p.ID.Value.String()
		}
		if p.Description.Set {
			description = p.Description.Value
		}

		prompts = append(prompts, &Prompt{
			client:      c,
			id:          id,
			name:        p.Name,
			description: description,
			tags:        p.Tags,
		})
	}

	return prompts, nil
}

// DeletePrompt deletes a prompt by ID.
func (c *Client) DeletePrompt(ctx context.Context, promptID string) error {
	promptUUID, err := uuid.Parse(promptID)
	if err != nil {
		return err
	}

	return c.apiClient.DeletePrompt(ctx, api.DeletePromptParams{ID: promptUUID})
}

// Delete deletes this prompt.
func (p *Prompt) Delete(ctx context.Context) error {
	return p.client.DeletePrompt(ctx, p.id)
}

// GetVersions retrieves all versions of this prompt.
func (p *Prompt) GetVersions(ctx context.Context, page, size int) ([]*PromptVersion, error) {
	promptUUID, err := uuid.Parse(p.id)
	if err != nil {
		return nil, err
	}

	params := api.GetPromptVersionsParams{
		ID:   promptUUID,
		Page: api.NewOptInt32(int32(page)), //nolint:gosec // G115: page values are bounded by API limits
		Size: api.NewOptInt32(int32(size)), //nolint:gosec // G115: size values are bounded by API limits
	}

	resp, err := p.client.apiClient.GetPromptVersions(ctx, params)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return []*PromptVersion{}, nil
	}

	versions := make([]*PromptVersion, 0, len(resp.Content))
	for _, v := range resp.Content {
		var id, promptID, commit, changeDescription string
		if v.ID.Set {
			id = v.ID.Value.String()
		}
		if v.PromptID.Set {
			promptID = v.PromptID.Value.String()
		}
		if v.Commit.Set {
			commit = v.Commit.Value
		}
		if v.ChangeDescription.Set {
			changeDescription = v.ChangeDescription.Value
		}

		versions = append(versions, &PromptVersion{
			client:            p.client,
			id:                id,
			promptID:          promptID,
			commit:            commit,
			template:          v.Template,
			changeDescription: changeDescription,
			tags:              v.Tags,
		})
	}

	return versions, nil
}

// PromptVersionOption is a functional option for configuring a PromptVersion.
type PromptVersionOption func(*promptVersionOptions)

type promptVersionOptions struct {
	changeDescription string
	promptType        PromptType
	tags              []string
}

// WithVersionChangeDescription sets the change description for the version.
func WithVersionChangeDescription(changeDescription string) PromptVersionOption {
	return func(o *promptVersionOptions) {
		o.changeDescription = changeDescription
	}
}

// WithVersionType sets the type for the version.
func WithVersionType(promptType PromptType) PromptVersionOption {
	return func(o *promptVersionOptions) {
		o.promptType = promptType
	}
}

// WithVersionTags sets the tags for the version.
func WithVersionTags(tags ...string) PromptVersionOption {
	return func(o *promptVersionOptions) {
		o.tags = tags
	}
}

// CreateVersion creates a new version of this prompt.
func (p *Prompt) CreateVersion(ctx context.Context, template string, opts ...PromptVersionOption) (*PromptVersion, error) {
	options := &promptVersionOptions{
		tags:       []string{},
		promptType: PromptTypeMustache,
	}
	for _, opt := range opts {
		opt(options)
	}

	promptUUID, err := uuid.Parse(p.id)
	if err != nil {
		return nil, err
	}

	versionUUID := uuid.New()

	req := api.CreatePromptVersionDetail{
		Name: p.name,
		Version: api.PromptVersionDetail{
			ID:                api.NewOptUUID(versionUUID),
			PromptID:          api.NewOptUUID(promptUUID),
			Template:          template,
			ChangeDescription: api.NewOptString(options.changeDescription),
			Type:              api.NewOptPromptVersionDetailType(api.PromptVersionDetailType(options.promptType)),
			Tags:              options.tags,
		},
	}

	resp, err := p.client.apiClient.CreatePromptVersion(ctx, api.NewOptCreatePromptVersionDetail(req))
	if err != nil {
		return nil, err
	}

	switch v := resp.(type) {
	case *api.PromptVersionDetail:
		var id, promptID, commit, changeDescription string
		if v.ID.Set {
			id = v.ID.Value.String()
		}
		if v.PromptID.Set {
			promptID = v.PromptID.Value.String()
		}
		if v.Commit.Set {
			commit = v.Commit.Value
		}
		if v.ChangeDescription.Set {
			changeDescription = v.ChangeDescription.Value
		}

		return &PromptVersion{
			client:            p.client,
			id:                id,
			promptID:          promptID,
			commit:            commit,
			template:          v.Template,
			changeDescription: changeDescription,
			tags:              v.Tags,
		}, nil
	default:
		return &PromptVersion{
			client:            p.client,
			id:                versionUUID.String(),
			promptID:          p.id,
			template:          template,
			changeDescription: options.changeDescription,
			tags:              options.tags,
		}, nil
	}
}

// Render renders the template with the given variables.
// Supports mustache-style {{variable}} placeholders.
func (v *PromptVersion) Render(variables map[string]string) string {
	result := v.template
	for key, value := range variables {
		// Replace mustache-style placeholders {{key}}
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// RenderWithDefault renders the template with the given variables,
// using default values for missing variables.
func (v *PromptVersion) RenderWithDefault(variables map[string]string, defaultValue string) string {
	result := v.template

	// First replace known variables
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Then replace any remaining placeholders with the default value
	re := regexp.MustCompile(`\{\{[^}]+\}\}`)
	result = re.ReplaceAllString(result, defaultValue)

	return result
}

// ExtractVariables returns a list of variable names in the template.
func (v *PromptVersion) ExtractVariables() []string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(v.template, -1)

	seen := make(map[string]bool)
	var variables []string
	for _, match := range matches {
		if len(match) > 1 {
			varName := strings.TrimSpace(match[1])
			if !seen[varName] {
				seen[varName] = true
				variables = append(variables, varName)
			}
		}
	}
	return variables
}
