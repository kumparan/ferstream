package ferstream

import (
	"time"

	"github.com/kumparan/go-utils"
	"github.com/kumparan/tapao"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// NatsEventTimeFormat time format for NatsEvent 'time' field
const NatsEventTimeFormat = time.RFC3339Nano

type (
	// NatsEvent :nodoc:
	NatsEvent struct {
		ID       int64  `json:"id"`
		IDString string `json:"id_string"`
		UserID   int64  `json:"user_id"`
		TenantID int64  `json:"tenant_id"`
		Time     string `json:"time"`
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

	// NatsEventAuditLogMessage :nodoc:
	NatsEventAuditLogMessage struct {
		Subject        string    `json:"subject"` // empty on publish
		ServiceName    string    `json:"service_name"`
		UserID         int64     `json:"user_id"`
		AuditableType  string    `json:"auditable_type"`
		AuditableID    string    `json:"auditable_id"`
		Action         string    `json:"action"`
		AuditedChanges string    `json:"audited_changes"`
		OldData        string    `json:"old_data,omitempty"`
		NewData        string    `json:"new_data,omitempty"`
		CreatedAt      time.Time `json:"created_at"`
		Error          error     `json:"error"`
	}

	// MessageParser :nodoc:
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

// GetIDString :nodoc:
func (n *NatsEvent) GetIDString() string {
	if n == nil {
		return ""
	}
	return n.IDString
}

// GetTime :nodoc:
func (n *NatsEvent) GetTime() string {
	if n == nil {
		return ""
	}
	return n.Time
}

// IsTimeValid :nodoc:
func (n *NatsEvent) IsTimeValid() bool {
	_, err := time.Parse(NatsEventTimeFormat, n.GetTime())
	return err == nil
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
		n.wrapError(errors.New("empty nats event"))
		return nil, n.Error
	}

	message, err := tapao.Marshal(n, tapao.With(tapao.JSON))
	if err != nil {
		n.wrapError(err)
		return nil, n.Error
	}

	return message, nil
}

// WithEvent :nodoc:
func (n *NatsEventMessage) WithEvent(e *NatsEvent) *NatsEventMessage {
	if e.GetID() <= 0 && e.GetIDString() == "" {
		n.wrapError(errors.New("empty id"))
		return n
	}

	if e.GetUserID() == 0 {
		n.wrapError(errors.New("empty user id"))
		return n
	}

	switch e.GetTime() {
	case "":
		e.Time = time.Now().Format(NatsEventTimeFormat)
	default:
		if !e.IsTimeValid() {
			n.wrapError(errors.New("invalid time format"))
			return n
		}
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

// ParseFromBytes :nodoc:
func (n *NatsEventMessage) ParseFromBytes(data []byte) (err error) {
	err = tapao.Unmarshal(data, &n, tapao.With(tapao.JSON))
	if err != nil {
		n.Error = errors.Wrap(n.Error, err.Error())
		return err
	}
	return
}

// AddSubject :nodoc:
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

// Build :nodoc:
func (n *NatsEventAuditLogMessage) Build() (data []byte, err error) {
	if n.Error != nil {
		return nil, n.Error
	}

	message, err := tapao.Marshal(n, tapao.With(tapao.JSON))
	if err != nil {
		n.wrapError(err)
		return nil, n.Error
	}

	return message, nil
}

// ParseFromBytes :nodoc:
func (n *NatsEventAuditLogMessage) ParseFromBytes(data []byte) (err error) {
	err = tapao.Unmarshal(data, &n, tapao.With(tapao.JSON))
	if err != nil {
		n.Error = errors.Wrap(n.Error, err.Error())
		return err
	}
	return
}

// AddSubject :nodoc:
func (n *NatsEventAuditLogMessage) AddSubject(subj string) {
	n.Subject = subj
}

// ToJSONString marshal message to JSON string
func (n *NatsEventAuditLogMessage) ToJSONString() (string, error) {
	bt, err := tapao.Marshal(n, tapao.With(tapao.JSON))
	return string(bt), err
}

// ToJSONByte marshal message to JSON byte
func (n *NatsEventAuditLogMessage) ToJSONByte() ([]byte, error) {
	return tapao.Marshal(n, tapao.With(tapao.JSON))
}

func (n *NatsEventAuditLogMessage) wrapError(err error) {
	if n.Error != nil {
		n.Error = errors.Wrap(n.Error, err.Error())
		return
	}
	n.Error = err
}

// ParseJSON parse JSON into Nats event message
func ParseJSON(in string) (*NatsEventMessage, error) {
	msg := &NatsEventMessage{}
	err := tapao.Unmarshal([]byte(in), msg, tapao.With(tapao.JSON))
	return msg, err
}

// ParseNatsEventAuditLogMessageFromBytes :nodoc:
func ParseNatsEventAuditLogMessageFromBytes(in []byte) (*NatsEventAuditLogMessage, error) {
	msg := &NatsEventAuditLogMessage{}
	err := tapao.Unmarshal(in, msg, tapao.With(tapao.JSON))
	return msg, err
}
