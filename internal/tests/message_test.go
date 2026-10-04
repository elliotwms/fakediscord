package tests

import "testing"

func TestMessage_Send(t *testing.T) {
	given, when, then := NewMessageStage(t)

	given.
		a_message()

	when.
		the_message_is_sent()

	then.
		the_message_should_be_received().and().
		the_message_can_be_fetched().and().
		the_message_has_the_author_as_the_session_user()
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
		the_message_should_have_n_reactions_to_emoji(1, "🧀")
}

func TestMessage_ReactionDelete(t *testing.T) {
	given, when, then := NewMessageStage(t)

	given.
		a_message().and().
		the_message_is_sent().and().
		the_message_is_reacted_to_with("🧀").and().
		the_message_should_have_n_reactions_to_emoji(1, "🧀")

	when.
		the_message_reactions_are_removed()

	then.
		the_message_should_have_n_reactions_to_emoji(0, "🧀")
}

func TestMessage_React_CustomEmoji(t *testing.T) {
	given, when, then := NewMessageStage(t)

	given.
		a_message().and().
		the_message_is_sent().and().
		we_listen_for_message_reaction_events().and()

	when.
		the_message_is_reacted_to_with("cheese:custom_id")

	then.
		the_message_should_have_n_reactions_to_emoji(1, "cheese:custom_id").and().
		a_message_reaction_add_event_should_have_been_received_with_id_and_name("custom_id", "cheese")
}

func TestMessage_React_CustomEmojiDelete(t *testing.T) {
	given, when, then := NewMessageStage(t)

	given.
		a_message().and().
		the_message_is_sent().and().
		we_listen_for_message_reaction_events().and().
		the_message_is_reacted_to_with("cheese:custom_id")

	when.
		the_message_reaction_is_removed("cheese:custom_id")

	then.
		the_message_should_have_n_reactions_to_emoji(0, "cheese:custom_id").and().
		a_message_reaction_remove_event_should_have_been_received_with_id_and_name("custom_id", "cheese")
}

func TestMessage_WithAttachment(t *testing.T) {
	for filename, contentType := range map[string]string{
		"cheese.jpg": "image/jpeg",
		"hello.txt":  "text/plain",
	} {
		t.Run(filename, func(t *testing.T) {
			given, when, then := NewMessageStage(t)

			given.
				a_message().and().
				an_attachment(filename, contentType)

			when.
				the_message_is_sent()

			then.
				the_message_should_have_an_attachment()
		})
	}
}

func TestMessage_WithImageAttachment(t *testing.T) {
	given, when, then := NewMessageStage(t)

	given.
		a_message().and().
		an_attachment("cheese.jpg", "image/jpeg")

	when.
		the_message_is_sent()

	then.
		the_message_should_have_an_attachment().and().
		the_first_attachment_should_have_a_resolution_set()
}

func TestMessage_WithLink(t *testing.T) {
	given, when, then := NewMessageStage(t)

	given.
		a_message().and().
		the_message_has_a_link()

	when.
		the_message_is_sent()

	then.
		the_message_should_have_an_embed()
}
