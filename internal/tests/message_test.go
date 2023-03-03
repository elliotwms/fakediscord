package tests

import "testing"

func TestMessage_Send(t *testing.T) {
	given, then, when := NewMessageStage(t)

	given.
		a_message()

	when.
		the_message_is_sent()

	then.
		the_message_is_received().and().
		the_message_can_be_fetched()
}
