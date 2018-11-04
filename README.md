# Slack
This library implements the [Slack API](https://api.slack.com/) Web and Real Time Messaging parts.
A simple example utilizing most of the functionality can be seen under `examples/scli` which implements a full featured CLI client for Slack - either interactive or for batch usage.

## Implemented [Methods](https://api.slack.com/methods)

| *Method* | *Description* | *Support* |
|--------------------------------------------------------------------------|--------------------------------------------------------------------|-------|
| [api.test](https://api.slack.com/methods/api.test)                       | Checks API calling code                                            | false |
| [auth.test](https://api.slack.com/methods/auth.test)                     | Checks authentication & identity                                   | true  |
| [channels.archive](https://api.slack.com/methods/channels.archive)       | Archives a channel                                                 | true  |
| [channels.create](https://api.slack.com/methods/channels.create)         | Creates a channel                                                  | true  |
| [channels.history](https://api.slack.com/methods/channels.history)       | Fetches history of messages and events from a channel              | true  |
| [channels.info](https://api.slack.com/methods/channels.info)             | Gets information about a channel                                   | true  |
| [channels.invite](https://api.slack.com/methods/channels.invite)         | Invites a user to a channel                                        | true  |
| [channels.join](https://api.slack.com/methods/channels.join)             | Joins a channel, creating it if needed                             | true  |
| [channels.kick](https://api.slack.com/methods/channels.kick)             | Removes a user from a channel                                      | true  |
| [channels.leave](https://api.slack.com/methods/channels.leave)           | Leaves a channel                                                   | true  |
| [channels.list](https://api.slack.com/methods/channels.list)             | Lists all channels in a Slack team                                 | true  |
| [channels.mark](https://api.slack.com/methods/channels.mark)             | Sets the read cursor in a channel                                  | true  |
| [channels.rename](https://api.slack.com/methods/channels.rename)         | Renames a channel                                                  | true  |
| [channels.setPurpose](https://api.slack.com/methods/channels.setPurpose) | Sets the purpose for a channel                                     | true  |
| [channels.setTopic](https://api.slack.com/methods/channels.setTopic)     | Sets the topic for a channel                                       | true  |
| [channels.unarchive](https://api.slack.com/methods/channels.unarchive)   | Unarchives a channel                                               | true  |
| [chat.delete](https://api.slack.com/methods/chat.delete)                 | Deletes a message                                                  | true  |
| [chat.postMessage](https://api.slack.com/methods/chat.postMessage)       | Sends a message to a channel                                       | true  |
| [chat.update](https://api.slack.com/methods/chat.update)                 | Updates a message                                                  | false |
| [emoji.list](https://api.slack.com/methods/emoji.list)                   | Lists custom emoji for a team                                      | true  |
| [files.delete](https://api.slack.com/methods/files.delete)               | Deletes a file                                                     | true  |
| [files.info](https://api.slack.com/methods/files.info)                   | Gets information about a team file                                 | true  |
| [files.list](https://api.slack.com/methods/files.list)                   | Lists & filters team files                                         | true  |
| [files.upload](https://api.slack.com/methods/files.upload)               | Uploads or creates a file                                          | true  |
| [groups.archive](https://api.slack.com/methods/groups.archive)           | Archives a private group                                           | true  |
| [groups.close](https://api.slack.com/methods/groups.close)               | Closes a private group                                             | true  |
| [groups.create](https://api.slack.com/methods/groups.create)             | Creates a private group                                            | true  |
| [groups.createChild](https://api.slack.com/methods/groups.createChild)   | Clones and archives a private group                                | true  |
| [groups.history](https://api.slack.com/methods/groups.history)           | Fetches history of messages and events from a private group        | true  |
| [groups.info](https://api.slack.com/methods/groups.info)                 | Gets information about a private group                             | true  |
| [groups.invite](https://api.slack.com/methods/groups.invite)             | Invites a user to a private group                                  | true  |
| [groups.kick](https://api.slack.com/methods/groups.kick)                 | Removes a user from a private group                                | true  |
| [groups.leave](https://api.slack.com/methods/groups.leave)               | Leaves a private group                                             | true  |
| [groups.list](https://api.slack.com/methods/groups.list)                 | Lists private groups that the calling user has access to           | true  |
| [groups.mark](https://api.slack.com/methods/groups.mark)                 | Sets the read cursor in a private group                            | true  |
| [groups.open](https://api.slack.com/methods/groups.open)                 | Opens a private group                                              | true  |
| [groups.rename](https://api.slack.com/methods/groups.rename)             | Renames a private group                                            | true  |
| [groups.setPurpose](https://api.slack.com/methods/groups.setPurpose)     | Sets the purpose for a private group                               | true  |
| [groups.setTopic](https://api.slack.com/methods/groups.setTopic)         | Sets the topic for a private group                                 | true  |
| [groups.unarchive](https://api.slack.com/methods/groups.unarchive)       | Unarchives a private group                                         | true  |
| [im.close](https://api.slack.com/methods/im.close)                       | Close a direct message channel                                     | true  |
| [im.history](https://api.slack.com/methods/im.history)                   | Fetches history of messages and events from direct message channel | true  |
| [im.list](https://api.slack.com/methods/im.list)                         | Lists direct message channels for the calling user                 | true  |
| [im.mark](https://api.slack.com/methods/im.mark)                         | Sets the read cursor in a direct message channel                   | true  |
| [im.open](https://api.slack.com/methods/im.open)                         | Opens a direct message channel                                     | true  |
| [rtm.start](https://api.slack.com/methods/rtm.start)                     | Starts a Real Time Messaging session                               | true  |
| [search.all](https://api.slack.com/methods/search.all)                   | Searches for messages and files matching a query                   | false |
| [search.files](https://api.slack.com/methods/search.files)               | Searches for files matching a query                                | false |
| [search.messages](https://api.slack.com/methods/search.messages)         | Searches for messages matching a query                             | false |
| [stars.list](https://api.slack.com/methods/stars.list)                   | Lists stars for a user                                             | false |
| [team.accessLogs](https://api.slack.com/methods/team.accessLogs)         | Gets the access logs for the current team                          | false |
| [team.info](https://api.slack.com/methods/team.info)                     | Gets information about the current team                            | true  |
| [users.getPresence](https://api.slack.com/methods/users.getPresence)     | Gets user presence information                                     | false |
| [users.info](https://api.slack.com/methods/users.info)                   | Gets information about a user                                      | true  |
| [users.list](https://api.slack.com/methods/users.list)                   | Lists all users in a Slack team                                    | true  |
| [users.setActive](https://api.slack.com/methods/users.setActive)         | Marks a user as active                                             | false |
| [users.setPresence](https://api.slack.com/methods/users.setPresence)     | Manually sets user presence                                        | false |

## Missing Features

- All of the above with a `false` in the `support` column.
- Testing

## Install

If you have a go workplace setup and working you can simply do:

 ```go get -u -t -v github.com/demisto/slack```

## Usage

There are 2 ways to initiate the library, both using the various configuration functions slack.Set*:

* Either using a test token retrieved from [Slack](https://api.slack.com/web) and then setting the token
```go
s, err = slack.New(slack.SetToken("test token retrieved from Slack"))
```

* Using [OAuth](https://golang.org/x/oauth2) - see a simple example using [uuid](https://github.com/wayn3h0/go-uuid/random) for random state. For this to work, you need to register your application with Slack.
```go
// Start the OAuth process

// First, generate a random state
uuid, err := random.New()
if err != nil {
  panic(err)
}
conf := &oauth2.Config{
  ClientID:     "Your client ID",
  ClientSecret: "Your client secret",
  Scopes:       []string{"client"}, // the widest scope - can be others depending on requirement
  Endpoint: oauth2.Endpoint{
    AuthURL:  "https://slack.com/oauth/authorize",
    TokenURL: "https://slack.com/api/oauth.access",
  },
}
// Store state somewhere you can use later with timestamp
// ...
url := conf.AuthCodeURL(uuid.String())
// Redirect user to the OAuth Slack page
http.Redirect(w, r, url, http.StatusFound)
```

```go
// Now, handle the redirected URL after the successful authentication

state := r.FormValue("state")
code := r.FormValue("code")
errStr := r.FormValue("error")
if errStr != "" {
  WriteError(w, &Error{"oauth_err", 401, "Slack OAuth Error", errStr})
  return
}
if state == "" || code == "" {
  WriteError(w, ErrBadContentRequest)
  return
}
// Retrieve the state you saved in the first step and make sure it is not too old
// ...
token, err := slack.OAuthAccess("Your client ID", "Your client secret", code, "")
if err != nil {
  WriteError(w, &Error{"oauth_err", 401, "Slack OAuth Error", err.Error()})
  return
}
// Done - you have the token - you can save it for later use
s, err := slack.New(slack.SetToken(token.AccessToken))
if err != nil {
  panic(err)
}
// Get our own user id
test, err := s.AuthTest()
if err != nil {
  panic(err)
}
```

## Authors

The library was written by `slavikm` as a side project to play with Slack API for `demisto`.
