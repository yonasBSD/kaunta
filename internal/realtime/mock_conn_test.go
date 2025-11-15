package realtime

import "io"

type testConn struct {
	writeMessages []writeCall
	readMessages  []readCall
	closeCalls    int
}

type writeCall struct {
	messageType int
	payload     []byte
}

type readCall struct {
	messageType int
	payload     []byte
	err         error
}

func (c *testConn) WriteMessage(messageType int, data []byte) error {
	c.writeMessages = append(c.writeMessages, writeCall{
		messageType: messageType,
		payload:     append([]byte(nil), data...),
	})
	return nil
}

func (c *testConn) ReadMessage() (messageType int, p []byte, err error) {
	if len(c.readMessages) == 0 {
		return 0, nil, io.EOF
	}
	msg := c.readMessages[0]
	c.readMessages = c.readMessages[1:]
	return msg.messageType, append([]byte(nil), msg.payload...), msg.err
}

func (c *testConn) Close() error {
	c.closeCalls++
	return nil
}
