from __future__ import print_function

import tbaapiv3client
import sys
from tbaapiv3client.rest import ApiException
# Define host
configuration = tbaapiv3client.Configuration(
    host = "https://www.thebluealliance.com/api/v3"
)

# Api ket 
configuration = tbaapiv3client.Configuration(
    host = "https://www.thebluealliance.com/api/v3",
    api_key = {
        'X-TBA-Auth-Key': sys.argv[1]
    }
)

# Enter context with api client
with tbaapiv3client.ApiClient(configuration) as api_client:
    api_instance = tbaapiv3client.EventApi(api_client)

    event_key = "2024mnmi2" # 2024 GCR my beloved

    try:
        api_instance.get_event_matches_simple(event_key)
    except ApiException as e:
        print("ERR")
