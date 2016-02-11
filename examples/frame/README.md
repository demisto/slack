# A simple Slack frame

A web page / frame that you can embed into your web page and will show a scroll of messages from selected channels

##Install

If you have a go workplace setup and working you can simply do:

```
go get -u -t -v github.com/demisto/slack/examples/frame
```

To change the static web portion of the frame, you should do the folowing:
* Install node and npm
* cd <frame_dir>/static
* To build prod distribution - npm run prod
* To do development - npm start

To embed the static resources into the project and generate static.go:
* go get github.com/mjibson/esc
* From the frame directory - esc -o static.go -pkg main -prefix static/dist static/dist/

To see debug prints and also serve the static content from files instead of the embedded version, pass -debug as a flag.

##Usage

If you installed as specified in the step above:
```
$GOPATH/bin/frame -t token_received_from_slack
```

There a are a few optional parameters including:
* The list of channels (default is all that the token is subscribed to): -ch general,ch1,ch2
* The colors in the form of: -colors "date:blue,user:red,channel:#618081,text:black,background:white" - also accepts standard CSS hex colors
* The time format: -t as specified at https://golang.org/pkg/time
* Certificate and private key files: -cert "concatenated certificate file with the full chain" -key "private key file"
* Debug prints: -debug

##Authors

Written by `slavikm` to play with the Slack API.
