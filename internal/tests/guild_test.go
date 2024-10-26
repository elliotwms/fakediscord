package tests

import "testing"

func TestGuild_Create(t *testing.T) {
	given, when, then := NewGuildStage(t)

	given.
		a_guild_named("create_test").and().
		the_session_expects_a_guild_create_event_for_the_guild()

	when.
		the_guild_is_created()

	then.
		no_error_should_be_returned().and().
		the_session_should_have_received_the_guild_create_event()
}

func TestGuild_Get(t *testing.T) {
	given, when, then := NewGuildStage(t)

	given.
		a_guild_named("create_test").and().
		the_guild_is_created()

	when.
		the_guild_is_fetched()

	then.
		no_error_should_be_returned()
}

func TestGuild_Delete(t *testing.T) {
	given, when, then := NewGuildStage(t)

	t.Run("delete", func(t *testing.T) {
		given.
			a_guild_named("delete_test").and().
			the_session_expects_a_guild_delete_event_for_the_guild().and().
			the_guild_is_created().and()

		when.
			the_guild_is_deleted()

		then.
			no_error_should_be_returned().and().
			the_session_should_have_received_the_guild_deleted_event()
	})

	t.Run("get_deleted", func(t *testing.T) {
		// test guild not found
		when.
			the_guild_is_fetched()

		then.
			an_error_should_be_returned()
	})
}
