from __future__ import print_function

import tbaapiv3client
import sys
from tbaapiv3client.rest import ApiException
import json
import os

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

    event_key = sys.argv[2] # Arg is event name
    schedule_dir = sys.argv[3]
 
    Matches = {}

    # Gets the event matches, strips them of frc and adds them to a dict
    try:
        matchesRaw = api_instance.get_event_matches_simple(event_key) 
        for match in matchesRaw:
            BlueNumbers = []
            RedNumbers = []
            ex = ""
            for key in match.alliances.blue.team_keys:
                BlueNumbers.append(int(key.strip("frc")))

            for key in match.alliances.red.team_keys:
                RedNumbers.append(int(key.strip("frc")))
            
            Matches.update({match.match_number: {"Blue":BlueNumbers, "Red": RedNumbers}})                        

    except ApiException as e:
        print("Exception when calling EventApi->get_event_teams: %s\n" % e)

    # dumps the schedule to schedule.json
    sorted_json_str = json.dumps(Matches, indent=4, sort_keys=True)
    
    file = open(os.path.join(schedule_dir,"schedule.json"), "w")
    file.write(sorted_json_str)

    print("Finished Filling Out Match schedule!")
