# Slack CLI - an interactive and batch cross-platform scriptable client

Command line Slack client that uses the Slack library and a test token to open an RTM connection, display (with color) the various messages sent to the user and allows the user to send text and do all the supported Slack actions (invites, creation, archive, reactions, file stuff, etc.)
In the interactive mode, supports tab-completion for channels, groups, users, etc. and standard readline capability like ctrl-w, ctrl-a, ctrl-e, history, etc.

##Missing Features

- Very much still a work in progress with some commands still missing

##Install

If you have a go workplace setup and working you can simply do:

```go get -u -t -v github.com/demisto/slack/examples/scli```

Binaries for the various platforms will be added.

## List of commands

TODO - right now, just take a look at the completer.go / commands.go

##Usage

If you installed as specified in the step above:
```
$GOPATH/bin/scli -t token_received_from_slack
```

You can set configure the client by creating a JSON configuration file. The default location is ~/.scli but can be specified via -c path/to/location
The configuration file looks as follows:
```json
{
	"Token":          "The token to authenticate to Slack - can also be specified via command line -t",
	"CommandPrefix":  "For internal commands, what is the prefix we are expecting - default is !",
	"DefaultChannel": "The default channel we will post to at the start - default is 'general'",
	"Colors":         {"date": "blue", "user": "red", "channel": "yellow", "text": "white"}
}
```

###Batch usage
* Posting a message to default channel
```echo 'Hello there' | $GOPATH/bin/scli```
* Posting to a channel (same can be done with group or DM with G/D command)
```echo '!c general Hello there' | $GOPATH/bin/scli```
* Listing last 100 messages on a channel
```echo '!hist general 100' | $GOPATH/bin/scli```

##Authors

Written by `slavikm` to play with the Slack API.
