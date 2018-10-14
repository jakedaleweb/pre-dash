# PRE Dashboard

## Purpose
Reports on key metrics for PRE squad

## Set-up

### Install necessary modules

`go get github.com/jakedaleweb/pre-dash`
`go get github.com/russellcardullo/go-pingdom/pingdom`
`go get github.com/subosito/gotenv`

### Add env variables

Create a `.env` file in the web root containing

```
export PINGDOM_EMAIL=whatever@whatever.com
export PINGDOM_PASSWORD=hunter12
export PINGDOM_TOKEN=tooooooooooooooooken
export FRESHDESK_URL=my.freshdesk.url
export FRESHDESK_TOKEN=verysecrettokendontshareever
export INCIDENT_VIEW_ID=206953
export MAC_VIEW_ID=1234
```

## Running

`cd $GOPATH/github.com/jakedaleweb/pre-dash`
`go install . && go run main.go pingdom.go freshdesk.go incidents.go toil.go types.go`

## Compiling for raspberry pi

`env GOOS=linux GOARCH=arm GOARM=5 go build`

Then scp the entire directory over, the templates and .env file are required to be in the same directory as the compiled binary. You can then run with `./pre-dash`.

