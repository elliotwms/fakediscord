# fakediscord

A highly experimental fake Discord server, intended to enable bot testing without calling the real Discord API,
analogous to [LocalStack](https://github.com/localstack/localstack).

## Features
Once completed, `fakediscord` will enable you to:

* Test offline, without interacting with the real Discord API
* Test with multiple simulated users, bot and non-bot
* Trigger server-side events to test commands etc., outside the 'official' bot flow
* Spin up a test instance with guilds and users preconfigured with YAML

Of course, you should also test your bot manually before releasing to the public, as there's a few things `fakediscord` **doesn't** intend to implement, including:

* Authorization - any action is allowed

## [Events](https://discord.com/developers/docs/topics/gateway-events)

As we develop `fakediscord` we will be aiming to implement each of the documented events along with their corresponding API interactions:

- [ ] [Hello](https://discord.com/developers/docs/topics/gateway-events#hello)
- [x] [Ready](https://discord.com/developers/docs/topics/gateway-events#ready)
- [ ] [Resumed](https://discord.com/developers/docs/topics/gateway-events#resumed)
- [ ] [Reconnect](https://discord.com/developers/docs/topics/gateway-events#reconnect)
- [ ] [Invalid Session](https://discord.com/developers/docs/topics/gateway-events#invalid-session)
- [ ] [Application Command Permissions Update](https://discord.com/developers/docs/topics/gateway-events#application-command-permissions-update)
- [ ] [Auto Moderation Rule Create](https://discord.com/developers/docs/topics/gateway-events#auto-moderation-rule-create)
- [ ] [Auto Moderation Rule Update](https://discord.com/developers/docs/topics/gateway-events#auto-moderation-rule-update)
- [ ] [Auto Moderation Rule Delete](https://discord.com/developers/docs/topics/gateway-events#auto-moderation-rule-delete)
- [ ] [Auto Moderation Action Execution](https://discord.com/developers/docs/topics/gateway-events#auto-moderation-action-execution)
- [x] [Channel Create](https://discord.com/developers/docs/topics/gateway-events#channel-create)
- [ ] [Channel Update](https://discord.com/developers/docs/topics/gateway-events#channel-update)
- [x] [Channel Delete](https://discord.com/developers/docs/topics/gateway-events#channel-delete)
- [x] [Channel Pins Update](https://discord.com/developers/docs/topics/gateway-events#channel-pins-update)
- [ ] [Thread Create](https://discord.com/developers/docs/topics/gateway-events#thread-create)
- [ ] [Thread Update](https://discord.com/developers/docs/topics/gateway-events#thread-update)
- [ ] [Thread Delete](https://discord.com/developers/docs/topics/gateway-events#thread-delete)
- [ ] [Thread List Sync](https://discord.com/developers/docs/topics/gateway-events#thread-list-sync)
- [ ] [Thread Member Update](https://discord.com/developers/docs/topics/gateway-events#thread-member-update)
- [ ] [Thread Members Update](https://discord.com/developers/docs/topics/gateway-events#thread-members-update)
- [x] [Guild Create](https://discord.com/developers/docs/topics/gateway-events#guild-create)
- [ ] [Guild Update](https://discord.com/developers/docs/topics/gateway-events#guild-update)
- [ ] [Guild Delete](https://discord.com/developers/docs/topics/gateway-events#guild-delete)
- [ ] [Guild Ban Add](https://discord.com/developers/docs/topics/gateway-events#guild-ban-add)
- [ ] [Guild Ban Remove](https://discord.com/developers/docs/topics/gateway-events#guild-ban-remove)
- [ ] [Guild Emojis Update](https://discord.com/developers/docs/topics/gateway-events#guild-emojis-update)
- [ ] [Guild Stickers Update](https://discord.com/developers/docs/topics/gateway-events#guild-stickers-update)
- [ ] [Guild Integrations Update](https://discord.com/developers/docs/topics/gateway-events#guild-integrations-update)
- [ ] [Guild Member Add](https://discord.com/developers/docs/topics/gateway-events#guild-member-add)
- [ ] [Guild Member Remove](https://discord.com/developers/docs/topics/gateway-events#guild-member-remove)
- [ ] [Guild Member Update](https://discord.com/developers/docs/topics/gateway-events#guild-member-update)
- [ ] [Guild Members Chunk](https://discord.com/developers/docs/topics/gateway-events#guild-members-chunk)
- [ ] [Guild Role Create](https://discord.com/developers/docs/topics/gateway-events#guild-role-create)
- [ ] [Guild Role Update](https://discord.com/developers/docs/topics/gateway-events#guild-role-update)
- [ ] [Guild Role Delete](https://discord.com/developers/docs/topics/gateway-events#guild-role-delete)
- [ ] [Guild Scheduled Event Create](https://discord.com/developers/docs/topics/gateway-events#guild-scheduled-event-create)
- [ ] [Guild Scheduled Event Update](https://discord.com/developers/docs/topics/gateway-events#guild-scheduled-event-update)
- [ ] [Guild Scheduled Event Delete](https://discord.com/developers/docs/topics/gateway-events#guild-scheduled-event-delete)
- [ ] [Guild Scheduled Event User Add](https://discord.com/developers/docs/topics/gateway-events#guild-scheduled-event-user-add)
- [ ] [Guild Scheduled Event User Remove](https://discord.com/developers/docs/topics/gateway-events#guild-scheduled-event-user-remove)
- [ ] [Integration Create](https://discord.com/developers/docs/topics/gateway-events#integration-create)
- [ ] [Integration Update](https://discord.com/developers/docs/topics/gateway-events#integration-update)
- [ ] [Integration Delete](https://discord.com/developers/docs/topics/gateway-events#integration-delete)
- [ ] [Interaction Create](https://discord.com/developers/docs/topics/gateway-events#interaction-create)
- [ ] [Invite Create](https://discord.com/developers/docs/topics/gateway-events#invite-create)
- [ ] [Invite Delete](https://discord.com/developers/docs/topics/gateway-events#invite-delete)
- [ ] [Message Create](https://discord.com/developers/docs/topics/gateway-events#message-create)
  - [x] Basic (via HTTP)
  - [ ] Embeds
  - [ ] Multipart
- [ ] [Message Update](https://discord.com/developers/docs/topics/gateway-events#message-update)
- [x] [Message Delete](https://discord.com/developers/docs/topics/gateway-events#message-delete)
- [ ] [Message Delete Bulk](https://discord.com/developers/docs/topics/gateway-events#message-delete-bulk)
- [x] [Message Reaction Add](https://discord.com/developers/docs/topics/gateway-events#message-reaction-add)
- [ ] [Message Reaction Remove](https://discord.com/developers/docs/topics/gateway-events#message-reaction-remove)
- [ ] [Message Reaction Remove All](https://discord.com/developers/docs/topics/gateway-events#message-reaction-remove-all)
- [ ] [Message Reaction Remove Emoji](https://discord.com/developers/docs/topics/gateway-events#message-reaction-remove-emoji)
- [ ] [Presence Update](https://discord.com/developers/docs/topics/gateway-events#presence-update)
- [ ] [Stage Instance Create](https://discord.com/developers/docs/topics/gateway-events#stage-instance-create)
- [ ] [Stage Instance Delete](https://discord.com/developers/docs/topics/gateway-events#stage-instance-delete)
- [ ] [Stage Instance Update](https://discord.com/developers/docs/topics/gateway-events#stage-instance-update)
- [ ] [Typing Start](https://discord.com/developers/docs/topics/gateway-events#typing-start)
- [ ] [User Update](https://discord.com/developers/docs/topics/gateway-events#user-update)
- [ ] [Voice State Update](https://discord.com/developers/docs/topics/gateway-events#voice-state-update)
- [ ] [Voice Server Update](https://discord.com/developers/docs/topics/gateway-events#voice-server-update)
- [ ] [Webhooks Update](https://discord.com/developers/docs/topics/gateway-events#webhooks-update)
