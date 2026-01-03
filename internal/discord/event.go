package discord

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type PrefixCommandListener interface {
	Prefix() string
	Handle(session *discordgo.Session, message *discordgo.MessageCreate) error
}

type PrefixCommandDispatcher struct {
	Listeners []PrefixCommandListener
}

func (dispatcher *PrefixCommandDispatcher) OnMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	for _, listener := range dispatcher.Listeners {
		if !strings.HasPrefix(message.Content, listener.Prefix()) {
			continue
		}

		if err := listener.Handle(session, message); err != nil {
			log.Printf("[DISCORD] failed to handle prefix command: %v", err)
		}
	}
}

type ReactionListener interface {
	EmojiName() string
	Handle(session *discordgo.Session, reaction *discordgo.MessageReaction) error
}

type ReactionDispatcher struct {
	AddListeners    []ReactionListener
	RemoveListeners []ReactionListener
}

func dispatchReaction(listeners []ReactionListener, session *discordgo.Session, reaction *discordgo.MessageReaction) {
	for _, listener := range listeners {
		if listener.EmojiName() != reaction.Emoji.Name {
			continue
		}

		if err := listener.Handle(session, reaction); err != nil {
			log.Printf("[DISCORD] failed to handle reaction: %v", err)
		}
	}
}

func (dispatcher *ReactionDispatcher) OnReactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	dispatchReaction(dispatcher.AddListeners, session, reaction.MessageReaction)
}

func (dispatcher *ReactionDispatcher) OnReactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	dispatchReaction(dispatcher.RemoveListeners, session, reaction.MessageReaction)
}

type SlashCommand interface {
	CreateCommand() *discordgo.ApplicationCommand
}

type InteractionApplicationListener interface {
	SlashCommand
	InteractionListener
}

type InteractionListener interface {
	InteractionType() discordgo.InteractionType
	InteractionID() string
	MatchInteractionID(InteractionID string) bool
	Handle(session *discordgo.Session, interaction *discordgo.Interaction) error
}

type InteractionDispatcher struct {
	Listeners []InteractionListener
}

func (dispatcher *InteractionDispatcher) OnInteractionCreate(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	for _, listener := range dispatcher.Listeners {
		if listener.InteractionType() != interaction.Type {
			continue
		}

		if want := listener.InteractionID(); want != "" {
			got := ""
			switch interaction.Type {
			case discordgo.InteractionMessageComponent:
				got = interaction.MessageComponentData().CustomID
			case discordgo.InteractionModalSubmit:
				got = interaction.ModalSubmitData().CustomID
			case discordgo.InteractionApplicationCommand:
				got = interaction.ApplicationCommandData().Name
			case discordgo.InteractionApplicationCommandAutocomplete:
				// Note: 未対応
			}

			if !listener.MatchInteractionID(got) {
				continue
			}
		}

		if err := listener.Handle(session, interaction.Interaction); err != nil {
			log.Printf("[DISCORD] failed to handle interaction: %v", err)
		}
	}
}
