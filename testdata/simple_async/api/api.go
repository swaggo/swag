package api


type MyMessage struct {
	Message string
	MessageID int
}

// @asyncapi
// @server myServer mqtt mqtt://broker.hivemq.com
// @channel myChannel myServer "Channel to hold events"
func ConfigEventDrivenChannel() {
	// write your code
}

// @asyncapi
// @operation send myChannel MyMessage
func OnMessageReceived() {
	// write your code
}
