package reqres

import (
	"net/http"
)

// WriteFile() - helper for sending file data to frontend
//
// example:
//
// WriteFile(w,
// "attachment; filename=my-journey.png",
// "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
// fmt.Sprint(len(fileData.Content)),
// fileData.Content
// )
func WriteFile(w http.ResponseWriter, disposition string, contentType string, contentLength string, content []byte) error {
	// Set headers
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", contentLength)

	w.WriteHeader(http.StatusOK)

	_, err := w.Write(content)
	if err != nil {
		return err
	}

	return nil
}
