package tests

import "testing"

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

func TestInteraction_Create_Invalid(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		given, when, then := NewInteractionStage(t)

		given.
			an_interaction()

		when.
			the_interaction_is_triggered()

		then.
			an_error_should_be_returned()

	})

	t.Run("missing type", func(t *testing.T) {
		given, when, then := NewInteractionStage(t)

		given.
			an_interaction().and().
			the_interaction_has_guild_id().and().
			the_interaction_has_channel_id()

		when.
			the_interaction_is_triggered()

		then.
			an_error_should_be_returned().and().
			the_error_should_contain("missing type")
	})

	t.Run("missing guild id", func(t *testing.T) {
		given, when, then := NewInteractionStage(t)

		given.
			an_interaction().and().
			the_interaction_has_type().and().
			the_interaction_has_data().and().
			the_interaction_has_channel_id()

		when.
			the_interaction_is_triggered()

		then.
			an_error_should_be_returned().and().
			the_error_should_contain("missing guild_id")

	})

	t.Run("missing channel id", func(t *testing.T) {
		given, when, then := NewInteractionStage(t)

		given.
			an_interaction().and().
			the_interaction_has_type().and().
			the_interaction_has_data().and().
			the_interaction_has_guild_id()

		when.
			the_interaction_is_triggered()

		then.
			an_error_should_be_returned().and().
			the_error_should_contain("missing channel_id")
	})
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
