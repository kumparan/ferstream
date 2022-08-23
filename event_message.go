package ferstream

import (
	"github.com/kumparan/go-utils"
	"google.golang.org/protobuf/proto"

	"github.com/kumparan/tapao"
	"github.com/pkg/errors"
)

type (
	// NatsEvent :nodoc:
	NatsEvent struct {
		ID       int64  `json:"id"`
		IDStr    string `json:"id_str"`
		UserID   int64  `json:"user_id"`
		TenantID int64  `json:"tenant_id"`
		Subject  string `json:"subject"` // empty on publish
	}

	// NatsEventMessage :nodoc:
	NatsEventMessage struct {
		NatsEvent *NatsEvent
		Body      string `json:"body"`
		OldBody   string `json:"old_body"`
		Request   []byte `json:"request"`
		Error     error  `json:"error"`
	}

	MessageParser interface {
		ParseFromBytes(data []byte) error
		AddSubject(subj string)
		ToJSONString() (string, error)
		ToJSONByte() ([]byte, error)
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

// GetIDStr :nodoc:
func (n *NatsEvent) GetIDStr() string {
	if n == nil {
		return ""
	}
	return n.IDStr
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
	if e.GetID() <= 0 && e.GetIDStr() == "" {
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

// ToJSONString marshal message to JSON string
func (n *NatsEventMessage) ToJSONString() (string, error) {
	bt, err := tapao.Marshal(n, tapao.With(tapao.JSON))
	return string(bt), err
}

// ToJSONByte marshal message to JSON byte
func (n *NatsEventMessage) ToJSONByte() ([]byte, error) {
	return tapao.Marshal(n, tapao.With(tapao.JSON))
}

// ParseJSON parse JSON into message
func ParseJSON(in string) (*NatsEventMessage, error) {
	msg := &NatsEventMessage{}
	err := tapao.Unmarshal([]byte(in), msg, tapao.With(tapao.JSON))
	return msg, err
}
