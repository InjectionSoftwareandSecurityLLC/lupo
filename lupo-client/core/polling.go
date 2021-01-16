package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// CheckForNewSession - Checks the Wolfpack server to see if a new session has been established.
func CheckForNewSession() {

	reqString := AuthURL + "&polling=true"

	resp, err := WolfPackHTTP.Get(reqString)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	jsonData, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		//fmt.Println(err)
		return
	}

	// Parse the JSON response
	// We are expecting a JSON string that is totally dynamic here due to the nature of broadcasts, but for sessions we expect a "successMessage" and a "message" key in the JSON object.
	var coreResponse map[string]interface{}

	if string(jsonData) != "" {
		err = json.Unmarshal(jsonData, &coreResponse)

		if err != nil {
			//fmt.Println(err)
			return
		}

		SuccessColorBold.Println("\n" + coreResponse["successMessage"].(string))
		fmt.Println(coreResponse["message"].(string))
	}

	time.Sleep(time.Second * 1)
	CheckForNewSession()

}
