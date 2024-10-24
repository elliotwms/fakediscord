package tests

import "testing"

func TestGuildCommand_Create(t *testing.T) {
	given, when, then := NewGuildCommandStage(t)

	given.
		a_guild_command()

	when.
		the_command_is_created()

	then.
		no_error_should_be_returned()
}

func TestGuildCommand_Create_Duplicate(t *testing.T) {
	given, when, then := NewGuildCommandStage(t)

	given.
		a_guild_command().and().
		the_command_is_created()

	when.
		the_command_is_recreated()

	then.
		no_error_should_be_returned().and().
		the_id_should_not_have_changed()
}

func TestGuildCommand_Get(t *testing.T) {
	given, when, then := NewGuildCommandStage(t)

	given.
		a_guild_command().and().
		the_command_is_created()

	when.
		the_command_is_fetched()

	then.
		no_error_should_be_returned()
}

func TestGuildCommand_Delete(t *testing.T) {
	given, when, then := NewGuildCommandStage(t)

	given.
		a_guild_command().and().
		the_command_is_created()

	when.
		the_command_is_deleted()

	then.
		no_error_should_be_returned()

	when.
		the_command_is_fetched()

	then.
		an_error_should_be_returned()
}
