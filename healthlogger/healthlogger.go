/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package healthlogger

import (
	"fmt"

	"github.com/Comcast/webpa-common/logging"
	hlog "github.com/InVisionApp/go-logger"
	"github.com/go-kit/kit/log"
)

type HealthLogger struct {
	log.Logger
	keyValPairs []interface{}
}

type Option func(*HealthLogger)

func NewHealthLogger(logger log.Logger) hlog.Logger {
	h := HealthLogger{logger, []interface{}{}}

	return &h
}

func (h *HealthLogger) Debug(msg ...interface{}) {
	logging.Debug(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Info(msg ...interface{}) {
	logging.Info(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Warn(msg ...interface{}) {
	logging.Warn(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Error(msg ...interface{}) {
	logging.Error(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Debugln(msg ...interface{}) {
	logging.Debug(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Infoln(msg ...interface{}) {
	logging.Info(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Warnln(msg ...interface{}) {
	logging.Warn(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Errorln(msg ...interface{}) {
	logging.Error(h, h.keyValPairs...).Log(logging.MessageKey(), msg)
}

func (h *HealthLogger) Debugf(format string, args ...interface{}) {
	logging.Debug(h, h.keyValPairs...).Log(logging.MessageKey(), fmt.Sprintf(format, args...))
}

func (h *HealthLogger) Infof(format string, args ...interface{}) {
	logging.Info(h, h.keyValPairs...).Log(logging.MessageKey(), fmt.Sprintf(format, args...))
}

func (h *HealthLogger) Warnf(format string, args ...interface{}) {
	logging.Warn(h, h.keyValPairs...).Log(logging.MessageKey(), fmt.Sprintf(format, args...))
}

func (h *HealthLogger) Errorf(format string, args ...interface{}) {
	logging.Error(h, h.keyValPairs...).Log(logging.MessageKey(), fmt.Sprintf(format, args...))
}

func (h *HealthLogger) WithFields(fields hlog.Fields) hlog.Logger {
	newKeyVals := h.keyValPairs
	for key, val := range fields {
		newKeyVals = append(newKeyVals, key, val)
	}
	return &HealthLogger{h, newKeyVals}
}
