package tests

import "testing"

func TestChannel_Create(t *testing.T) {
	given, when, then := NewChannelStage(t)

	given.
		a_channel_does_not_exist_named("foo")

	when.
		a_channel_is_created_named("foo")

	then.
		state_contains_the_channel()
}

func TestChannel_Delete(t *testing.T) {
	given, when, then := NewChannelStage(t)

	given.
		a_channel_is_created_named("foo")

	when.
		the_channel_is_deleted()

	then.
		state_does_not_contain_the_channel()
}
