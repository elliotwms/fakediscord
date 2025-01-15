package tests

import "testing"

func TestApplicationCommand_Create(t *testing.T) {
	given, when, then := NewApplicationCommandStage(t)

	given.
		an_application_command()

	when.
		the_command_is_created_globally()

	then.
		no_error_should_be_returned()
}

func TestApplicationCommand_Create_Duplicate(t *testing.T) {
	given, when, then := NewApplicationCommandStage(t)

	given.
		an_application_command().and().
		the_command_is_created_globally()

	when.
		the_command_is_recreated()

	then.
		no_error_should_be_returned().and().
		the_id_should_not_have_changed()
}

func TestApplicationCommand_Get(t *testing.T) {
	given, when, then := NewApplicationCommandStage(t)

	given.
		an_application_command().and().
		the_command_is_created_globally()

	when.
		the_command_is_fetched()

	then.
		no_error_should_be_returned()
}

func TestApplicationCommand_Delete(t *testing.T) {
	given, when, then := NewApplicationCommandStage(t)

	given.
		an_application_command().and().
		the_command_is_created_globally()

	when.
		the_command_is_deleted()

	then.
		no_error_should_be_returned()

	when.
		the_command_is_fetched()

	then.
		an_error_should_be_returned()
}

func TestApplicationCommand_Guild_Create(t *testing.T) {
	given, when, then := NewApplicationCommandStage(t)

	given.
		a_guild().and().
		an_application_command()

	when.
		the_command_is_created_in_guild()

	then.
		no_error_should_be_returned()
}

func TestApplicationCommand_Guild_Create_Duplicate(t *testing.T) {
	given, when, then := NewApplicationCommandStage(t)

	given.
		a_guild().and().
		an_application_command().and().
		the_command_is_created_in_guild()

	when.
		the_command_is_recreated()

	then.
		no_error_should_be_returned().and().
		the_id_should_not_have_changed()
}

func TestApplicationCommand_Guild_Get(t *testing.T) {
	given, when, then := NewApplicationCommandStage(t)

	given.
		a_guild().and().
		an_application_command().and().
		the_command_is_created_in_guild()

	when.
		the_command_is_fetched()

	then.
		no_error_should_be_returned()
}

func TestApplicationCommand_Guild_Delete(t *testing.T) {
	given, when, then := NewApplicationCommandStage(t)

	given.
		a_guild().and().
		an_application_command().and().
		the_command_is_created_in_guild()

	when.
		the_command_is_deleted()

	then.
		no_error_should_be_returned()

	when.
		the_command_is_fetched()

	then.
		an_error_should_be_returned()
}
