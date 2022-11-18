// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httpcli

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Error struct {
	// Code is the HTTP response status code
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Body    string      `json:"body,omitempty"`
	Header  http.Header `json:"header,omitempty"`
}

func (e *Error) Error() string {
	data, er := json.Marshal(e)
	if er == nil {
		return string(data)
	}

	return fmt.Sprintf(`{"code":%d,"message":%q}`,
		e.Code, e.Message+", "+er.Error())
}
