# A Guide to setting up GreenScout's Backend

setup.go uses a LOT of recursion. It can be confusing, but I believe in you.

# Steps

1. To start the setup, run

    ```bash
    go run main.go setup
    ```
2. It will retrieve the configs from yaml. If they don't exist, it'll create a new object with default fields. If any of the following are already met and validated by what it reads, it will not ask for additional input on those fields.
3. It will verify the existence of the essential databases (auth and users). If you're a member of TheGreenMachine on github, clone GreenScout-Databases into this project. If not, **FUTURE DEVS WILL ADD TOOLS TO MAKE YOUR OWN**
4. It will ensure the existence of the configuration files neccecary for the google sheets API. A guide is provided [here](https://developers.google.com/sheets/api/quickstart/go#set_up_your_environment)
    - Make sure to publish your google cloud project. Otherwise, any generated tokens will expire very quickly.
5. It will ensure sqlite3 exists and is accessible by it. If you need to, download it [here](https://sqlite.org/download.html)
6. It will ensure the existence of the various InputtedJson directories, creating them if they don't exist.
7. It will ensure the existence of the RSA keys used for logging in, creating them if they don't exist
8. It will ensure the existence of scout.db, creating it if it doesn't exist.
9. It will always attempt to download the [Python TBA API](https://github.com/TBA-API/tba-api-client-python.git) in order to ensure it has access to it.
10. 
-   If it is in production mode, It will ensure there is a configured ipv4 address and corresponding domain name
-   If it is in testing mode, It will skip this step
11. It will validate the python driver it will use to activate its various python files. Most people just enter [python](https://www.python.org/downloads), though some will be different. On mac, the default is **python3**. If you have multiple python versions on the same machine, you must use whichever one is connected to pip and thus the TBA API package. for me, it was **python3.11**
12. It will ensure there is a valid blue alliance event key. If you need one, get it [here](https://www.thebluealliance.com/apidocs/v3)
13. 
-   If you enter in an event key recognized by TBA, it will accept that and move on, writing the event schedule and team list for that event to files.
-   If you enter a custom event key (begins with 'c'), it will accept that, but should pit scouting be enabled, require that you have a TeamLists file.
14. It will ensure there is a valid google sheets spreadsheet ID. THis is found between **/d/** and **/edit** in a google sheets link. If the account the token was generated for has no access to this sheet or cannot read from it, it will treat it as invalid. 
15. 
-   First, it will ask if the user would like to use slack or not. **It is highly recommended to use slack.**
-   If the user chose to use it, it will require a valid bot token and channel to a workspace it has access to and can write to. 
16. It will automatically configure logging. The only way to set logging configs is through YAML.
17. Finally, it will store these configurations in memory at constants.CachedConfigs and to the project at setup/greenscout.config.yaml

Now you can run
```bash
sudo go run main.go prod
```
to enter production mode! Don't forget, you can edit any of these configurations through the yaml file. 

DO NOT mess with any configuration called `Configured`. These are program-set only. 
