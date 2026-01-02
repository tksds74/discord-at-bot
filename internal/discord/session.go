package discord

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type SessionConfig interface {
	Token() string
	Intent() discordgo.Intent
	Handlers() []any
}

type sessionConfig struct {
	token    string
	intent   discordgo.Intent
	handlers []any
}

func (config *sessionConfig) Token() string {
	return config.token
}

func (config *sessionConfig) Intent() discordgo.Intent {
	return config.intent
}

func (config *sessionConfig) Handlers() []any {
	return append([]any(nil), config.handlers...)
}

func (config *sessionConfig) validate() error {
	if config.token == "" {
		return errors.New("token is required")
	}
	if len(config.handlers) == 0 {
		return errors.New("no handlers registered")
	}
	return nil
}

type sessionConfigOption func(*sessionConfig) error

func NewSessionConfig(opts ...sessionConfigOption) (*sessionConfig, error) {
	config := &sessionConfig{}
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	if err := config.validate(); err != nil {
		return nil, err
	}
	return config, nil
}

func WithToken(token string) sessionConfigOption {
	return func(config *sessionConfig) error {
		if (token) == "" {
			return fmt.Errorf("discord token is required")
		}
		config.token = token
		return nil
	}
}

func WithIntent(intent discordgo.Intent) sessionConfigOption {
	return func(config *sessionConfig) error {
		config.intent |= intent
		return nil
	}
}

func WithMessageCreateHandler(
	handler func(*discordgo.Session, *discordgo.MessageCreate),
) sessionConfigOption {
	return withHandler(handler)
}

func WithMessageReactionAddHandler(
	handler func(*discordgo.Session, *discordgo.MessageReactionAdd),
) sessionConfigOption {
	return withHandler(handler)
}

func WithMessageReactionRemoveHandler(
	handler func(*discordgo.Session, *discordgo.MessageReactionRemove),
) sessionConfigOption {
	return withHandler(handler)
}

func WithInteractionCreateHandler(
	handler func(*discordgo.Session, *discordgo.InteractionCreate),
) sessionConfigOption {
	return withHandler(handler)
}

func withHandler(handler any) sessionConfigOption {
	return func(config *sessionConfig) error {
		config.handlers = append(config.handlers, handler)
		return nil
	}
}

type SessionManager struct {
	session *discordgo.Session
}

func (manager *SessionManager) Open(config SessionConfig) error {
	if manager.session != nil {
		_ = manager.session.Close()
		manager.session = nil
	}

	session, err := discordgo.New("Bot " + config.Token())
	if err != nil {
		return err
	}

	if config.Intent() != 0 {
		session.Identify.Intents = config.Intent()
	}

	for _, handler := range config.Handlers() {
		session.AddHandler(handler)
	}

	if err := session.Open(); err != nil {
		return err
	}

	manager.session = session
	return nil
}

func (manager *SessionManager) Close() error {
	if manager.session == nil {
		return nil
	}

	err := manager.session.Close()
	manager.session = nil
	return err
}
