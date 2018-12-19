## checkmon

This is a super simple check for monitoring status emails based on gmail api examples


### Installation

Install Go https://golang.org/doc/install

Get gmail credentials https://developers.google.com/gmail/api/quickstart/go

clone this repo. 

`$ cd gcheck`

`$ go get`

Change the query string how you like:

`const queryStr = "from:Shinken-monitoring  newer_than:1d"`

`$ go build`

Run in the background on your server

`$ nohup ./checkmon &` 

Note that the first time the program runs, you will have to get an authentication token.
 
TODO: proper logging. 
