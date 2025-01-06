package tests

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestInteraction_Create(t *testing.T) {
	given, when, then := NewInteractionStage(t)

	given.
		a_registered_message_command_handler().and().
		a_valid_interaction()

	when.
		the_interaction_is_triggered()

	then.
		no_error_should_be_returned().and().
		the_interaction_should_be_valid().and().
		the_command_handler_should_have_been_triggered()
}

func TestInteraction_Create_Invalid_Empty(t *testing.T) {
	given, when, then := NewInteractionStage(t)

	given.
		an_interaction()

	when.
		the_interaction_is_triggered()

	then.
		an_error_should_be_returned()

}
func TestInteraction_Create_Invalid(t *testing.T) {
	tests := map[string]func(i *discordgo.InteractionCreate){
		"missing type": func(i *discordgo.InteractionCreate) {
			i.Type = 0
		},
		"missing guild_id": func(i *discordgo.InteractionCreate) {
			i.GuildID = ""
		},
		"missing channel_id": func(i *discordgo.InteractionCreate) {
			i.ChannelID = ""
		},
		"missing application_id": func(i *discordgo.InteractionCreate) {
			i.AppID = ""
		},
		"connection for application_id not found": func(i *discordgo.InteractionCreate) {
			i.AppID = "0290742494824366183" // a valid but non-existent ID
		},
	}

	for err, modifier := range tests {
		t.Run(err, func(t *testing.T) {
			given, when, then := NewInteractionStage(t)

			given.
				a_valid_interaction().and().
				the_interaction_has(modifier)

			when.
				the_interaction_is_triggered()

			then.
				an_error_should_be_returned().and().
				the_error_should_contain(err)
		})
	}
}

func TestInteraction_Callback(t *testing.T) {
	given, when, then := NewInteractionStage(t)

	given.
		the_interaction_is_triggered()

	when.
		the_interaction_callback_is_triggered().and().
		the_interaction_message_is_updated()

	then.
		a_message_should_have_been_posted_in_the_channel()
}

func TestInteraction_Callback_WithMessage(t *testing.T) {
	given, when, then := NewInteractionStage(t)

	given.
		the_interaction_is_triggered()

	when.
		the_interaction_callback_is_triggered_with_a_message()

	then.
		a_message_should_have_been_posted_in_the_channel()
}
