package tests

import "testing"

func TestChannel_Create(t *testing.T) {
	given, when, then := NewChannelStage(t)

	given.
		a_channel_does_not_exist_named("foo")

	when.
		a_channel_is_created_named("foo")

	then.
		state_contains_the_channel().and().
		get_guild_channels_contains_channel()
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

func TestChannel_CreateThread(t *testing.T) {
	given, when, then := NewChannelStage(t)

	given.
		a_channel_does_not_exist_named("foo")

	when.
		a_channel_is_created_named("foo").and().
		a_thread_is_created_named("bar")

	then.
		state_contains_the_channel().and().
		state_contains_the_thread().and().
		get_guild_channels_contains_channel().and().
		get_guild_channels_does_not_contain_thread().and().
		the_thread_should_have_default_auto_archive_time()
}
