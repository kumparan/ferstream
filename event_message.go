package ferstream

import (
	"google.golang.org/protobuf/proto"

	"github.com/kumparan/go-utils"
	"github.com/kumparan/tapao"
	"github.com/pkg/errors"
)

type (
	// NatsEvent :nodoc:
	NatsEvent struct {
		ID       int64
		UserID   int64
		TenantID int64
		Subject  string // empty on publish
	}

	// NatsEventMessage :nodoc:
	NatsEventMessage struct {
		NatsEvent *NatsEvent
		Body      string
		OldBody   string
		Request   []byte
		Error     error
	}

	MessageParser interface {
		ParseFromBytes(data []byte) error
		AddSubject(subj string)
	}
)

// GetID :nodoc:
func (n *NatsEvent) GetID() int64 {
	if n == nil {
		return 0
	}
	return n.ID
}

// GetUserID :nodoc:
func (n *NatsEvent) GetUserID() int64 {
	if n == nil {
		return 0
	}
	return n.UserID
}

// GetTenantID :nodoc:
func (n *NatsEvent) GetTenantID() int64 {
	if n == nil {
		return 0
	}
	return n.TenantID
}

// GetSubject :nodoc:
func (n *NatsEvent) GetSubject() string {
	if n == nil {
		return ""
	}
	return n.Subject
}

// NewNatsEventMessage :nodoc:
func NewNatsEventMessage() *NatsEventMessage {
	return &NatsEventMessage{}
}

// Build :nodoc:
func (n *NatsEventMessage) Build() (data []byte, err error) {
	if n.Error != nil {
		return nil, n.Error
	}

	if n.NatsEvent == nil {
		n.wrapError(errors.New("empty nats nats event"))
		return nil, n.Error
	}

	message, err := tapao.Marshal(n)
	if err != nil {
		n.wrapError(err)
		return nil, n.Error
	}

	return message, nil
}

// WithEvent :nodoc:
func (n *NatsEventMessage) WithEvent(e *NatsEvent) *NatsEventMessage {
	if e.GetID() <= 0 {
		n.wrapError(errors.New("empty id"))
		return n
	}

	if e.GetUserID() == 0 {
		n.wrapError(errors.New("empty user id"))
		return n
	}

	n.NatsEvent = e
	return n
}

// WithBody :nodoc:
func (n *NatsEventMessage) WithBody(body interface{}) *NatsEventMessage {
	n.Body = utils.Dump(body)
	return n
}

// WithOldBody :nodoc:
func (n *NatsEventMessage) WithOldBody(body interface{}) *NatsEventMessage {
	n.OldBody = utils.Dump(body)
	return n
}

// WithRequest :nodoc:
func (n *NatsEventMessage) WithRequest(req proto.Message) *NatsEventMessage {
	b, err := tapao.Marshal(req, tapao.With(tapao.Protobuf))
	if err != nil {
		n.wrapError(err)
		return n
	}

	n.Request = b
	return n
}

func (n *NatsEventMessage) wrapError(err error) {
	if n.Error != nil {
		n.Error = errors.Wrap(n.Error, err.Error())
		return
	}
	n.Error = err
}

func (n *NatsEventMessage) ParseFromBytes(data []byte) (err error) {
	err = tapao.Unmarshal(data, &n, tapao.FallbackWith(tapao.JSON))
	if err != nil {
		n.Error = errors.Wrap(n.Error, err.Error())
		return err
	}
	return
}

func (n *NatsEventMessage) AddSubject(subj string) {
	n.NatsEvent.Subject = subj
}
