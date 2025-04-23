package linereversal

import "log"

type LineProcessor struct {
	SessionID      int
	Buffer         []byte
	DataChannel    chan []byte
	SessionChannel chan SessionMessage
}

func NewLineProcessor(sessionID int, sessionChannel chan SessionMessage, dataChannel chan []byte) *LineProcessor {
	lineProcessor := &LineProcessor{
		Buffer:         make([]byte, 0, 1024),
		DataChannel:    dataChannel,
		SessionChannel: sessionChannel,
	}
	// Begin the session listening
	go lineProcessor.ReadSessionData()
	return lineProcessor
}

func (lp *LineProcessor) ReadSessionData() {
	for {
		// Read the data from the session
		data, ok := <-lp.DataChannel
		if !ok {
			return
		}
		// Process the data
		lp.Buffer = append(lp.Buffer, data...)

		// Check to see if we have a new line in the buffer
		lp.ProcessData()
	}
}

func (lp *LineProcessor) ProcessData() {
	log.Printf("Processing data %s", string(lp.Buffer))
	for {
		if len(lp.Buffer) == 0 {
			return
		}

		newLineIndex := -1
		for i, b := range lp.Buffer {
			if b == '\n' {
				newLineIndex = i
				break
			}
		}

		if newLineIndex == -1 {
			return
		}
		// We have a new line, so we need to process the data
		lp.ReverseAndSend(lp.Buffer[:newLineIndex])
		// Remove the processed data from the buffer
		lp.Buffer = lp.Buffer[newLineIndex+1:]
	}
}

func (lp *LineProcessor) ReverseAndSend(data []byte) {
	// Reverse the data
	reversed := make([]byte, len(data)+1)
	for i, b := range data {
		reversed[len(data)-1-i] = b
	}
	// Add the new line character
	reversed[len(data)] = '\n'
	// Send the reversed data to the session
	lp.SessionChannel <- WriteMessage(reversed)
}
