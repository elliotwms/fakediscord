package tests

import "testing"

func TestMessage_Send(t *testing.T) {
	given, when, then := NewMessageStage(t)

	given.
		a_message()

	when.
		the_message_is_sent()

	then.
		the_message_is_received().and().
		the_message_can_be_fetched()
}

func TestMessage_Pin(t *testing.T) {
	given, when, then := NewMessageStage(t)

	given.
		a_message().and().
		the_message_is_sent()

	when.
		the_message_is_pinned()

	then.
		the_message_has_been_pinned()
}

func TestMessage_React(t *testing.T) {
	given, when, then := NewMessageStage(t)

	given.
		a_message().and().
		the_message_is_sent()

	when.
		the_message_is_reacted_to_with("🧀")

	then.
		the_message_has_n_reactions_to_emoji(1, "🧀")
}

func TestMessage_ReactionDelete(t *testing.T) {
	given, when, then := NewMessageStage(t)

	given.
		a_message().and().
		the_message_is_sent().and().
		the_message_is_reacted_to_with("🧀").and().
		the_message_has_n_reactions_to_emoji(1, "🧀")

	when.
		the_message_reactions_are_removed()

	then.
		the_message_has_n_reactions_to_emoji(0, "🧀")
}
