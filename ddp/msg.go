package ddp

import (
	"encoding/json"
	"io"
	"log"
)

type message struct {
	Type string `json:"msg"`
}

type connectMessage struct {
	message
	Session string   `json:"session"`
	Version string   `json:"version"`
	Support []string `json:"support"`
}

type connectedMessage struct {
	message
	Session string `json:"session"`
}

func newConnectedMessage(session string) *connectedMessage {
	return &connectedMessage{
		message: message{session},
		Session: session,
	}
}

type failedMessage struct {
	message
	Version string `json:"version"`
}

func newFailedMessage(version string) *failedMessage {
	return &failedMessage{
		message: message{"failed"},
		Version: version,
	}
}

type pingMessage struct {
	message
	Id string `json:"id,omitempty"`
}

type pongMessage struct {
	message
	Id string `json:"id,omitempty"`
}

func newPongMessage(id string) *pongMessage {
	return &pongMessage{
		message: message{"pong"},
		Id:      id,
	}
}

type subMessage struct {
	message
	Id     string        `json:"id"`
	Name   string        `json:"name"`
	Params []interface{} `json:"params,omitempty"`
}

type unsubMessage struct {
	message
	Id string `json:"id"`
}

type nosubMessage struct {
	message
	Id    string    `json:"id"`
	Error *errorObj `json:"error,omitempty"`
}

type addedMessage struct {
	message
	Collection string                 `json:"collection"`
	Id         string                 `json:"id"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

type changedMessage struct {
	message
	Collection string                 `json:"collection"`
	Id         string                 `json:"id"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Cleared    []string               `json:"cleared,omitempty"`
}

type removedMessage struct {
	message
	Collection string `json:"collection"`
	Id         string `json:"id"`
}

type readyMessage struct {
	message
	Subs []string `json:"subs"`
}

type addedBeforeMessage struct {
	message
	Collection string                 `json:"collection"`
	Id         string                 `json:"id"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Before     *string                `json:"before"`
}

type movedBeforeMessage struct {
	message
	Collection string  `json:"collection"`
	Id         string  `json:"id"`
	Before     *string `json:"before"`
}

type methodMessage struct {
	message
	Method     string        `json:"method"`
	Params     []interface{} `json:"params,omitempty"`
	Id         string        `json:"id"`
	RandomSeed string        `json:"randomSeed,omitempty"`
}

type resultMessage struct {
	message
	Id     string      `json:"id"`
	Error  *errorObj   `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func newResultMessage(id string, error *errorObj, result interface{}) *resultMessage {
	return &resultMessage{
		message: message{"result"},
		Id:      id,
		Error:   error,
		Result:  result,
	}
}

type updatedMessage struct {
	message
	Methods []string `json:"methods"`
}

func newUpdatedMessage(methods []string) *updatedMessage {
	return &updatedMessage{
		message: message{"updated"},
		Methods: methods,
	}
}

type errorMessage struct {
	message
	Reason           string      `json:"reason"`
	OffendingMessage interface{} `json:"offendingMessage,omitempty"`
}

func newErrorMessage(reason string, offendingMessage interface{}) *errorMessage {
	return &errorMessage{
		message:          message{"error"},
		Reason:           reason,
		OffendingMessage: offendingMessage,
	}
}

func sendErrorMessage(w io.Writer, reason string, offendingMessage interface{}) {
	enc := json.NewEncoder(w)
	err := enc.Encode(newErrorMessage(reason, offendingMessage))
	if err != nil {
		log.Printf("encode ddp json: %v\n", err)
	}
}

type errorObj struct {
	Error     string `json:"error"`
	Reason    string `json:"reason,omitempty"`
	Message   string `json:"message,omitempty"`
	ErrorType string `json:"errorType"`
}
