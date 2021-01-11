package core

import "fmt"

// PSK - global PSK for listeners to manage and set the server PSK
var PSK string

type Response struct {
	Response    string
	CurrentPSK  string
	Instruction string
}

func ManagePSK(psk string, isRandom bool, operator string) (response string, currentPSK string, instruction string) {
	if PSK == "" {
		if isRandom {
			LogData(operator + " executed: listener manage -r true")
			PSK = GeneratePSK()
			response := "Your new random PSK is:"
			currentPSK := PSK
			instruction := "Embed the PSK into any implants to connect to any listeners in this instance."
			return response, currentPSK, instruction
		} else {
			LogData(operator + " executed: listener manage")
			response := "Warning, you did not provide a PSK, this will keep the current PSK. You can ignore this if you did not want to update the PSK."
			PSK = DefaultPSK
			currentPSK := PSK
			instruction := ""
			return response, currentPSK, instruction
		}
	} else {
		LogData(operator + " executed: listener manage -k <redacted>")
		response := "Your new PSK is:"
		fmt.Println(PSK)
		PSK = psk
		currentPSK := PSK
		instruction := "Embed the PSK into any implants to connect to any listeners in this instance."
		return response, currentPSK, instruction
	}

}
