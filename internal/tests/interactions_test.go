package tests

import "testing"

func TestInteraction(t *testing.T) {
	given, when, then := NewInteractionStage(t)

	given.
		a_registered_message_command_handler()

	when.
		an_interaction_is_triggered()

	then.
		the_interaction_should_be_valid().and().
		the_command_handler_should_have_been_triggered()
}

func TestInteraction_Callback(t *testing.T) {
	given, when, then := NewInteractionStage(t)

	given.
		an_interaction_is_triggered()

	when.
		the_interaction_callback_is_triggered().and().
		the_interaction_message_is_updated()

	then.
		a_message_should_have_been_posted_in_the_channel()
}

func TestInteraction_Callback_WithMessage(t *testing.T) {
	given, when, then := NewInteractionStage(t)

	given.
		an_interaction_is_triggered()

	when.
		the_interaction_callback_is_triggered_with_a_message()

	then.
		a_message_should_have_been_posted_in_the_channel()
}
