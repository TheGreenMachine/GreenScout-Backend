from __future__ import print_function
import tbaapiv3client
from tbaapiv3client.rest import ApiException
from pprint import pprint
import sys
from datetime import datetime
import json
import os
# Defining the host is optional and defaults to https://www.thebluealliance.com/api/v3
# See configuration.py for a list of all supported configuration parameters.
configuration = tbaapiv3client.Configuration(
    host = "https://www.thebluealliance.com/api/v3"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: apiKey
configuration = tbaapiv3client.Configuration(
    host = "https://www.thebluealliance.com/api/v3",
    api_key = {
        'X-TBA-Auth-Key': sys.argv[1]
    }
)
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['X-TBA-Auth-Key'] = 'Bearer'

# Enter a context with an instance of the API client
with tbaapiv3client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = tbaapiv3client.EventApi(api_client)
    if(len(sys.argv) < 3) :
        event_key = ""
    else:
        event_key = sys.argv[2] # str | TBA Event Key, eg `2016nytr`
if_modified_since = 'if_modified_since_example' # str | Value of the `Last-Modified` header in the most recently cached response by the client. (optional)

events = {}

# Gets all events from this year, adds them to a dict and dumps to events.json
try:
        eventsSimple = api_instance.get_events_by_year_simple(datetime.now().year)
        for event in eventsSimple: 
              if event.country != "Israel": # Had to include this somewhere. Free Palestine! -Tag
                events[event.key] = event.name
        sorted_data = dict(sorted(events.items(), key=lambda item: item[1]))

        jsonStr = json.dumps(sorted_data, indent=4)

        file = open("events.json", "w")
        file.write(jsonStr)
        
except ApiException as e:
        print("ERR")