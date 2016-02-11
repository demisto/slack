# A simple Slack frame

A web page / frame that you can embed into your web page and will show a scroll of messages from selected channels

##Install

If you have a go workplace setup and working you can simply do:

```go get -u -t -v github.com/demisto/slack/examples/frame```

Binaries for the various platforms will be added.

##Usage

If you installed as specified in the step above:
```
$GOPATH/bin/sframe -t token_received_from_slack
```

There a are a few optional parameters including:
* The list of channels (default is all that the token is subscribed to)
* The colors in the form of: -colors "date:blue,user:red,channel:yellow,text:white" - also accepts standard CSS hex colors

##Authors

Written by `slavikm` to play with the Slack API.
