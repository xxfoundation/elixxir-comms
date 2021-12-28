package mixmessages

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"strings"
)

func MakeNotificationsCSV(l []*NotificationData) string {
	output := make([][]string, len(l))
	for i, n := range l {
		output[i] = []string{base64.StdEncoding.EncodeToString(n.MessageHash),
			base64.StdEncoding.EncodeToString(n.IdentityFP)}
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.WriteAll(output); err != nil {
		jww.FATAL.Printf("Failed to make notificationsCSV: %+v", err)
	}
	return string(buf.Bytes())
}

func BuildNotificationCSV(ndList []*NotificationData, maxSize int) ([]byte, []*NotificationData) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	numWritten := 0

	for _, nd := range ndList {
		output := []string{base64.StdEncoding.EncodeToString(nd.MessageHash),
			base64.StdEncoding.EncodeToString(nd.IdentityFP)}

		if buf.Len()+len(output) > maxSize {
			break
		}

		if err := w.Write(output); err != nil {
			jww.FATAL.Printf("Failed to make notificationsCSV: %+v", err)
		}

		numWritten++
	}
	w.Flush()

	return buf.Bytes(), ndList[numWritten:]
}

func UpdateNotificationCSV(l *NotificationData, oldBuf *bytes.Buffer, maxSize int) bool {
	output := make([]string, 2)
	output = []string{base64.StdEncoding.EncodeToString(l.MessageHash),
		base64.StdEncoding.EncodeToString(l.IdentityFP)}

	addition := &bytes.Buffer{}

	w := csv.NewWriter(addition)
	if err := w.Write(output); err != nil {
		jww.FATAL.Printf("Failed to make notificationsCSV: %+v", err)
	}
	w.Flush()

	if addition.Len()+oldBuf.Len() >= maxSize {
		return false
	}
	if _, err := oldBuf.Write(addition.Bytes()); err != nil {
		jww.FATAL.Printf("Failed to append addition to CSV: %+v", err)
	}
	return true
}

func DecodeNotificationsCSV(data string) ([]*NotificationData, error) {
	r := csv.NewReader(strings.NewReader(data))
	read, err := r.ReadAll()
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to decode notifications CSV")
	}

	l := make([]*NotificationData, len(read))
	for i, touple := range read {
		messageHash, err := base64.StdEncoding.DecodeString(touple[0])
		if err != nil {
			return nil, errors.WithMessage(err, "Failed decode an element")
		}
		identityFP, err := base64.StdEncoding.DecodeString(touple[1])
		if err != nil {
			return nil, errors.WithMessage(err, "Failed decode an element")
		}
		l[i] = &NotificationData{
			EphemeralID: 0,
			IdentityFP:  identityFP,
			MessageHash: messageHash,
		}
	}
	return l, nil
}