# Logging
### This project uses the GreenLogger, based on the logging wrapper of the same name used on team 1816's robots. 

Greenlogger comes with the ability to write output to the console, log files called **GSLogs**, and to slack.

Like the functions found in golang's fmt package, all greenlogger methods come with associated formatted versions.

# Logging a printout
``` 
greenlogger.LogMessage("Some message")
```

Outputs "Some Message" to the console

Outputs {Timestamp} : "Some message" to the current GSLog

# Logging an Error
``` 
greenlogger.LogError(error, "Problem somewhere")
```

Outputs "Problem somewhere : {error.error()}" to the console

Outputs "ERR AT {timestamp} : Problem somewhere: {error.error()} to the current GSLog

Sends "ERR: Problem somewhere: {error.error()}" to slack

# Logging a Fatal
``` 
greenlogger.FatalError(error, "Problem somewhere")
```

Outputs "FATAL: Problem somewhere : {error.error()}" to the console

Outputs "FATAL: ERR AT {timestamp} : Problem somewhere: {error.error()} to the current GSLog

Sends "FATAL:ERR: Problem somewhere: {error.error()}" to slack

Shuts down the program

## What does ELog mean?
It means exclusively log to the GSLog, not the console or slack.

## Why is there a getLogger() function?
It returns a logging interface that allows the http handler to write its errors in the same fashion as all others. 