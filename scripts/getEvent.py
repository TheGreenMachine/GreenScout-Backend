from __future__ import print_function
import tbaapiv3client
from tbaapiv3client.rest import ApiException
from pprint import pprint
import sys
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

# Checks for the validity of an event 
try:
        api_response = api_instance.get_event(event_key)
        pprint(api_response.name)
except ApiException as e:
        print("ERR")