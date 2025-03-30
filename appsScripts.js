/**
 * Gets the scouters of a given match
 * 
 * @param {*} match The match number
 * @param {*} color The driverstation color (Red/Blue)
 * @param {*} driverStation The driverstation number (1-3)
 * @returns The scouters
 */
function GETSCOUTER(match, color, driverStation) {
    if (!Number.isInteger(match) || match < 1 ) {
      return "Please enter a valid match"
    }
  
    var colorAsString = color.toUpperCase();
  
    if (colorAsString != "RED" && colorAsString != "BLUE") {
      return "Please enter a valid color"
    }
  
    if (!Number.isInteger(driverStation) || driverStation < 1 || driverStation > 3) {
      return "Please enter a valid Driverstation"
    }
  
    var formData =`{"Match":${match}, "isBlue":${colorAsString == "BLUE"}, "DriverStation":${driverStation}}`;
  
  // Because payload is a JavaScript object, it is interpreted as
  // as form data. (No need to specify contentType; it automatically
  // defaults to either 'application/x-www-form-urlencoded'
  // or 'multipart/form-data')
  var options = {
    'method' : 'get',
    'payload' : formData
  };
  
  var response = UrlFetchApp.fetch('https://tagciccone.com/scouterLookup', options); //TODO: Probably change this
  return response.getContentText()
  }
  
  const quotes = [
    'Drink water.',
    'Slow down!',
    'Take a breather.',
    'Be nice to venue staff.',
    'Rithwik 2024',
    'Ask your local William Teskey about United States Presidents!',
    'Purdy is watching.',
    'Nerd.',
    'It is always funny to mess with Evan.',
    ':)',
    "I couldnt think of any more quotes",
    "No, I will not be telling you every quote I put in here.",
    "Are you cooked or are you cooking?",
    'Remind Aahil to do his webwork',
    'Remind Ethan to do his webwork',
    'Drew Cole 秃头书呆子',
    'Should you be looking at this, or doing strategy?',
    'Be like Usain Bolt wearing heelys.',
    'Did you lose the plot, or could it just not keep up with you?',
    'Monster energy is not a substitute for sleeping.',
    'Getting a buzzcut is a good life choice.',
    'Lock in.',
    'Use the :toocool: emote on slack more.',
    'Peace and Love.',
    'What year was the Year Without a Summer?',
    'What year did the second bank of the United States obtain its charter?',
    'Ryan McGoff',
    'Deodorant is a good choice to make.',
    'The Sun is Sunny.',
    "Compartmentalization is healthy if you don't think about it.",
    "At least you're not in the Duluth stands. Unless you are in which case tough I guess?",
    "Go Knicks!",
    "876 💙",
    "18! 16!",
    "Woolsey is wrong the halo show sucks",
    "Check out the newest project from Tag and Micheal: Currently unnamed study tool!",
    "Rithwik Barbados Barber",
    "Naz Reid.",
    ]
  
    /**
     * Gets a random motivational quote
     * @param {*} anything Pass in the contents of a commonly mutated range of cells to allow common refreshes. I commonly use RawData!B2:B
     * @returns A motivational quote
     */
  function GETMOTIVATIONALQUOTE(anything) {
    var quote =  quotes[(Math.floor(Math.random() * quotes.length))]
    return quote;
  }