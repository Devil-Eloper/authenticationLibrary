package lib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Devil-Eloper/authenticationLibrary/lib/datamodels"
	splunkLogger "github.com/Devil-Eloper/splunkLogger/lib"
	"io"
	"net/http"
	"strconv"
	"time"
)

func RetrieveAuthToken(httpClient *http.Client, logger *splunkLogger.Logger) (string, error) {
	envErrors := initializeEnvironment()
	if envErrors != nil {
		logger.Info("", "RAT1.0", "")
		return "", envErrors
	}
	
	redisClient := NewRedisClient()
	defer func(redisClient *RedisClient) {
		err := redisClient.Close()
		if err != nil {

		}
	}(redisClient)
	redisObject, err := redisClient.Get(accessTokenObject)
	if redisObject != "" {
		var tokenObject datamodels.TokenObject
		err = json.Unmarshal([]byte(redisObject), &tokenObject)
		if err != nil {
			panic(err) // TODO
		}
		expiryDate := tokenObject.RetrievedDate.Add(time.Second * time.Duration(tokenObject.ExpiresIn))
		currentDate := time.Now()
		if currentDate.Before(expiryDate) {
			return tokenObject.AccessToken, nil
		}
	}
	apiURL := environment[authUrl]

	req, err := http.NewRequest(post, apiURL, nil)
	if err != nil {
		logger.Info("", "RAT1.1", "")
		return "", err
	}

	inputBytes := []byte(environment[clientId] + ":" + environment[clientSecret])
	encodedString := base64.StdEncoding.EncodeToString(inputBytes)
	authHeader := basic + encodedString
	req.Header.Set(authorization, authHeader)
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Info("", "RAT1.2", "")
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Info("", "RAT1.3", "")
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logger.Info("", "RAT1.4", "")
		return "", nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Info("", "RAT1.5", "")
		return "", err
	}
	var jsonData map[string]interface{}

	if err := json.Unmarshal([]byte(body), &jsonData); err != nil {
		logger.Info("", "RAT1.6", "")
		return "", err
	}

	responseToken, foundAccessToken := jsonData[accessToken]
	expiresInToken, foundExpiresIn := jsonData[expiresIn]
	if foundAccessToken && foundExpiresIn {
		responseToken := fmt.Sprintf("%v", responseToken)
		expiresInString := fmt.Sprintf("%v", expiresInToken)
		var tokenObject datamodels.TokenObject
		tokenObject.AccessToken = responseToken
		expiresIn, err := strconv.Atoi(expiresInString)
		if err != nil {
			panic(err)
		}
		tokenObject.ExpiresIn = expiresIn
		tokenObject.RetrievedDate = time.Now()
		jsonData, err := json.Marshal(tokenObject)
		if err != nil {
			fmt.Println("JSON marshaling error:", err)
			return "", err
		}
		logger.Info("", "RAT1.7", "") // This log is indicative of successful token generation
		err = redisClient.Set(accessTokenObject, jsonData, 0)
		return responseToken, nil
	}
	return "", nil
}
