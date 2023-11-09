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

func RetrieveAuthToken(httpClient *http.Client, logger *splunkLogger.Logger, messageId string) (string, error) {
	envErrors := initializeEnvironment()
	logger.Info(messageId, "RAT1.0", "Redis Milestone 1")
	if envErrors != nil {
		logger.Error(messageId, "RAT1.1", "Redis Milestone 1.1 "+envErrors.Error())
		return "", envErrors
	}

	redisClient := NewRedisClient()
	defer func(redisClient *RedisClient) {
		err := redisClient.Close()
		if err != nil {
			logger.Error(messageId, "RAT1.2", "Redis Milestone 1.2 "+err.Error())
		}
	}(redisClient)
	redisObject, err := redisClient.Get(accessTokenObject)
	logger.Info(messageId, "RAT1.3", "Redis Milestone 1.3")
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
	logger.Info(messageId, "RAT1.4", "Redis Milestone 1.4")
	req, err := http.NewRequest(post, apiURL, nil)
	if err != nil {
		logger.Error(messageId, "RAT1.5", "Redis Milestone 1.5 "+err.Error())
		return "", err
	}

	inputBytes := []byte(environment[clientId] + ":" + environment[clientSecret])
	encodedString := base64.StdEncoding.EncodeToString(inputBytes)
	authHeader := basic + encodedString
	req.Header.Set(authorization, authHeader)
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error(messageId, "RAT1.6", "Redis Milestone 1.6 "+err.Error())
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Info(messageId, "RAT1.7", "Redis Milestone 1.7")
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logger.Error(messageId, "RAT1.8", "Redis Milestone 1.8 Status Code:"+strconv.Itoa(resp.StatusCode))
		return "", nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(messageId, "RAT1.9", "Redis Milestone 1.9 "+err.Error())
		return "", err
	}
	var jsonData map[string]interface{}

	if err := json.Unmarshal([]byte(body), &jsonData); err != nil {
		logger.Error(messageId, "RAT1.10", "Redis Milestone 1.10 "+err.Error())
		return "", err
	}

	responseToken, foundAccessToken := jsonData[accessToken]
	expiresInToken, foundExpiresIn := jsonData[expiresIn]
	logger.Info(messageId, "RAT1.11", "Redis Milestone 1.11")
	if foundAccessToken && foundExpiresIn {
		responseToken := fmt.Sprintf("%v", responseToken)
		expiresInString := fmt.Sprintf("%v", expiresInToken)
		var tokenObject datamodels.TokenObject
		tokenObject.AccessToken = responseToken
		expiresIn, err := strconv.Atoi(expiresInString)
		if err != nil {
			logger.Error(messageId, "RAT1.12", "Redis Milestone 1.12 "+err.Error())
			panic(err)
		}
		tokenObject.ExpiresIn = expiresIn
		tokenObject.RetrievedDate = time.Now()
		jsonData, err := json.Marshal(tokenObject)
		if err != nil {
			logger.Info(messageId, "RAT1.13", "Redis Milestone 1.13 "+err.Error())
			return "", err
		}
		logger.Info(messageId, "RAT1.14", "Redis Milestone 1.14") // This log is indicative of successful token generation
		err = redisClient.Set(accessTokenObject, jsonData, 0)
		logger.Info(messageId, "RAT1.15", "Redis Milestone 1.15")
		return responseToken, nil
	}
	return "", nil
}
