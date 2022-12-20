package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RelayData - byte slice for storing json data that is relayed to a session subshell
var RelayData []byte

// ActiveSession - global copy of the currently active session for the polling library to track
var ActiveSession int

// Status check variable to know if user is in the chat CLI or not so messages can be displayed properly
var ActiveChat = false

// SessionRelay - function expected to be called by a goroutine in the context of the session subshell to check if any new data exists to be relayed
func SessionRelay() {
	Poll()
}

func ChatRelay() {

	Poll()
}

// Poll - Executes all polling functions starting with
func Poll() {

	reqString := AuthURL + "&polling=true"

	resp, err := WolfPackHTTP.Get(reqString)

	if err != nil {
		ErrorColorBold.Println("\nPolling connection could not reach Wolfpack server, server might be offline the error is:")
		ErrorColorUnderline.Println(err)
		WarningColorBold.Println("Trying again after 5 seconds...")
		time.Sleep(time.Second * 5)
		Poll()
	}

	defer resp.Body.Close()

	jsonData, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		//fmt.Println(err)
		return
	}

	RelayData = jsonData

	time.Sleep(time.Second * 1)

	if !ActiveChat {
		if CheckForNewSession(RelayData) {
			Poll()
		} else if CheckForSessionData(RelayData) {
			Poll()
		} else if CheckForFileDownload(RelayData) {
			Poll()
		}
	} else if ActiveChat {
		if CheckForNewChatData(RelayData) {
			Poll()
		}
	}

	Poll()

}

// CheckForNewSession - Checks the Wolfpack server to see if a new session has been established.
func CheckForNewSession(jsonData []byte) bool {

	// Parse the JSON response
	// We are expecting a JSON string that is totally dynamic here due to the nature of broadcasts, but for sessions we expect a "successMessage" and a "message" key in the JSON object.
	var coreResponse map[string]interface{}

	if string(jsonData) != "" {
		err := json.Unmarshal(jsonData, &coreResponse)

		if err != nil {
			//fmt.Println(err)
			return false
		}

		_, messageExists := coreResponse["message"]

		if messageExists {
			SuccessColorBold.Println("\n" + coreResponse["successMessage"].(string))
			fmt.Println(coreResponse["message"].(string))
			return true
		}
	}

	return false

}

// CheckForSessionData - polls the wolfpack server to see if an session's implant has returned any data
func CheckForSessionData(jsonData []byte) bool {

	// Clean up new lines from potentially JSON breaking output
	re := regexp.MustCompile(`\n`)
	jsonDataString := re.ReplaceAllString(string(jsonData), "\\n")

	jsonDataString = strings.Replace(jsonDataString, "\r", "\\r", -1)

	jsonCleanData := []byte(jsonDataString)

	// Parse the JSON response
	// We are expecting a JSON string that is totally dynamic here due to the nature of broadcasts, but for implant response data we expect a "data" key in the JSON object.
	var coreResponse map[string]interface{}

	if string(jsonData) != "" {
		err := json.Unmarshal(jsonCleanData, &coreResponse)

		if err != nil {
			//fmt.Println(err)
			return false
		}
		_, dataExists := coreResponse["data"]
		if dataExists {

			fmt.Println("\nSession " + strconv.Itoa(ActiveSession) + " returned:\n" + coreResponse["data"].(string))
			return true
		}
	}

	return false

}

// CheckForFileDownload - this polling function is only triggered by the session CLI's download subcommand, this polls the wolfpack server to see if an session's implant has returned any files data to download
func CheckForFileDownload(jsonData []byte) bool {

	// Parse the JSON response
	// We are expecting a JSON string that is totally dynamic here due to the nature of broadcasts, but for implant response data we expect a "data" key in the JSON object.
	var coreResponse map[string]interface{}

	if string(jsonData) != "" {
		err := json.Unmarshal(jsonData, &coreResponse)

		if err != nil {
			//fmt.Println(err)
			return false
		}
		_, dataExists := coreResponse["filename"]
		if dataExists {
			_, dataExists = coreResponse["file"]
			if dataExists {
				DownloadFile(coreResponse["filename"].(string), coreResponse["file"].(string))
				return true
			}
		}
	}

	return false

}

// CheckForNewChatData - polls the wolfpack server to see if any new chat data has come through
func CheckForNewChatData(jsonData []byte) bool {

	// Clean up new lines from potentially JSON breaking output
	re := regexp.MustCompile(`\n`)
	jsonDataString := re.ReplaceAllString(string(jsonData), "\\n")

	jsonCleanData := []byte(jsonDataString)

	// Parse the JSON response
	// We are expecting a JSON string that is totally dynamic here due to the nature of broadcasts, but for implant response data we expect a "data" key in the JSON object.
	var coreResponse map[string]interface{}

	if string(jsonData) != "" {
		err := json.Unmarshal(jsonCleanData, &coreResponse)

		if err != nil {
			//fmt.Println(err)
			return false
		}
		_, dataExists := coreResponse["chatData"]
		if dataExists {

			fmt.Println(coreResponse["chatData"].(string))
			return true
		}
	}

	return false

}
