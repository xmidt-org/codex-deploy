#
#  Copyright 2019 Comcast Cable Communications Management, LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#

Feature: events
  In order to use the events api
  As an API user
  I need to bea able to get a list of events

  #set the global values for this feature file
  Background: Map of variables these tests use
    Given for the feature "get" the following environmental variables should be loaded "basicauthorization"
    And Gungnir is at "http://localhost:7000"

  Scenario: get single event
    Given the data:
      | deviceID         | birthdate           | metadata        | payload                                                                                              | type   | transaction_uuid                     |
      | mac:deadbeafcafe | 1578893303353430201 | {"/trust": "0"} | ewoJCSJpZCI6ICJtYWM6ZGVhZGJlYWZjYWZlIiwKCQkidHMiOiAiMjAyMC0wMS0xM1QwNToyODoyMy4zNTM0MzAyMDFaIgoJfQ== | online | a51fecfd-2dc8-421d-ab52-1eaa7d301513 |
    When I send "GET" request to "/events" with ID "mac:deadbeafcafe"
    Then the response code should be 200
    And the response should match json:
    """
    [
      {
        "msg_type": 4,
        "source": "dns:talaria",
        "dest": "event:device-status/mac:deadbeafcafe/online",
        "transaction_uuid": "a51fecfd-2dc8-421d-ab52-1eaa7d301513",
        "content_type": "json",
        "metadata": {
            "/trust": "0"
        },
        "payload": "ewoJCSJpZCI6ICJtYWM6ZGVhZGJlYWZjYWZlIiwKCQkidHMiOiAiMjAyMC0wMS0xM1QwNToyODoyMy4zNTM0MzAyMDFaIgoJfQ==",
        "birth_date": 1578893303353430201
      }
    ]
    """

  Scenario: get multiple events
    Given the data:
      | deviceID         | birthdate           | metadata                     | payload                                                                                                                                                                                                                                                                                                                                                                                                          | type    | transaction_uuid                     |
      | mac:deadbeafcafe | 1578893303353430201 | {"/trust": "0"}              | ewoJCSJpZCI6ICJtYWM6ZGVhZGJlYWZjYWZlIiwKCQkidHMiOiAiMjAyMC0wMS0xM1QwNToyODoyMy4zNTM0MzAyMDFaIgoJfQ==                                                                                                                                                                                                                                                                                                             | online  | a51fecfd-2dc8-421d-ab52-1eaa7d301513 |
      | mac:deadbeafcafe | 1578893303353440201 | {"/neat": "awesome"}         | ewoJCSJpZCI6ICJtYWM6ZGVhZGJlYWZjYWZlIiwKCQkidHMiOiAiMjAyMC0wMS0xM1QxNzowNDowMS41NjE0NzE0NjVaIiwKCQkiYnl0ZXMtc2VudCI6IDAsCgkJIm1lc3NhZ2VzLXNlbnQiOiAwLAoJCSJieXRlcy1yZWNlaXZlZCI6IDAsCgkJIm1lc3NhZ2VzLXJlY2VpdmVkIjogMCwKCQkiY29ubmVjdGVkLWF0IjogIjIwMjAtMDEtMTNUMTA6MzI6MzEuNTEyMDEzNjU1WiIsCgkJInVwLXRpbWUiOiAiNmgzMW0zMC4wNDk0NjMzMzlzIiwKCQkicmVhc29uLWZvci1jbG9zdXJlIjogIipubyBlcnJvcio6cmVhZGVycm9yIgoJfQ== | offline | a51fecfd-2dc8-421d-ab52-1eaa7d301514 |
      | mac:deadbeafcafe | 1578893303353450201 | {"/boot-time": "1578909505"} | ewoJCSJpZCI6ICJtYWM6ZGVhZGJlYWZjYWZlIiwKCQkidHMiOiAiMjAyMC0wMS0xM1QwNToyODoyMy4zNTM0MzAyMDFaIgoJfQ==                                                                                                                                                                                                                                                                                                             | online  | a51fecfd-2dc8-421d-ab52-1eaa7d301515 |

    When I send "GET" request to "/events" with ID "mac:deadbeafcafe"
    Then the response code should be 200
    And the response should match json:
    """
    [
        {
            "msg_type": 4,
            "source": "dns:talaria",
            "dest": "event:device-status/mac:deadbeafcafe/online",
            "transaction_uuid": "a51fecfd-2dc8-421d-ab52-1eaa7d301515",
            "content_type": "json",
            "metadata": {
                "/boot-time": "1578909505"
            },
            "payload": "ewoJCSJpZCI6ICJtYWM6ZGVhZGJlYWZjYWZlIiwKCQkidHMiOiAiMjAyMC0wMS0xM1QwNToyODoyMy4zNTM0MzAyMDFaIgoJfQ==",
            "birth_date": 1578893303353450201
        },
        {
            "msg_type": 4,
            "source": "dns:talaria",
            "dest": "event:device-status/mac:deadbeafcafe/offline",
            "transaction_uuid": "a51fecfd-2dc8-421d-ab52-1eaa7d301514",
            "content_type": "json",
            "metadata": {
                "/neat": "awesome"
            },
            "payload": "ewoJCSJpZCI6ICJtYWM6ZGVhZGJlYWZjYWZlIiwKCQkidHMiOiAiMjAyMC0wMS0xM1QxNzowNDowMS41NjE0NzE0NjVaIiwKCQkiYnl0ZXMtc2VudCI6IDAsCgkJIm1lc3NhZ2VzLXNlbnQiOiAwLAoJCSJieXRlcy1yZWNlaXZlZCI6IDAsCgkJIm1lc3NhZ2VzLXJlY2VpdmVkIjogMCwKCQkiY29ubmVjdGVkLWF0IjogIjIwMjAtMDEtMTNUMTA6MzI6MzEuNTEyMDEzNjU1WiIsCgkJInVwLXRpbWUiOiAiNmgzMW0zMC4wNDk0NjMzMzlzIiwKCQkicmVhc29uLWZvci1jbG9zdXJlIjogIipubyBlcnJvcio6cmVhZGVycm9yIgoJfQ==",
            "birth_date": 1578893303353440201
        },
        {
            "msg_type": 4,
            "source": "dns:talaria",
            "dest": "event:device-status/mac:deadbeafcafe/online",
            "transaction_uuid": "a51fecfd-2dc8-421d-ab52-1eaa7d301513",
            "content_type": "json",
            "metadata": {
                "/trust": "0"
            },
            "payload": "ewoJCSJpZCI6ICJtYWM6ZGVhZGJlYWZjYWZlIiwKCQkidHMiOiAiMjAyMC0wMS0xM1QwNToyODoyMy4zNTM0MzAyMDFaIgoJfQ==",
            "birth_date": 1578893303353430201
        }
    ]
    """