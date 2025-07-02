package xrayclient

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"

	"xui-tg-admin/internal/config"
	"xui-tg-admin/internal/constants"
	"xui-tg-admin/internal/helpers"
	"xui-tg-admin/internal/models"
)

// Client represents an X-ray API client
type Client struct {
	httpClient   *resty.Client
	serverConfig config.ServerConfig
	cookieCache  *cache.Cache
	logger       *logrus.Logger
}

// XrayAPIResponse represents the response from the X-ray API
type XrayAPIResponse struct {
	Success bool        `json:"success"`
	Msg     string      `json:"msg"`
	Obj     interface{} `json:"obj"`
}

// NewClient creates a new X-ray API client
func NewClient(serverConfig config.ServerConfig, logger *logrus.Logger) *Client {
	httpClient := resty.New().
		SetTimeout(constants.DefaultTimeout * time.Second).
		SetRetryCount(constants.DefaultRetryCount).
		SetRetryWaitTime(constants.DefaultRetryWaitTime * time.Second).
		SetRetryMaxWaitTime(constants.DefaultRetryMaxWaitTime * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	return &Client{
		httpClient:   httpClient,
		serverConfig: serverConfig,
		cookieCache:  cache.New(constants.CacheExpiration*time.Minute, constants.CacheCleanupInterval*time.Minute),
		logger:       logger,
	}
}

// Login logs in to the X-ray API
func (c *Client) Login(ctx context.Context) error {
	// Check if we already have a valid session
	if _, found := c.cookieCache.Get("session"); found {
		return nil
	}

	c.logger.Infof("Logging in to X-ray API at %s", c.serverConfig.APIURL)
	c.logger.Debugf("Using username: %s", c.serverConfig.User)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"username": c.serverConfig.User,
			"password": c.serverConfig.Password,
		}).
		Post(fmt.Sprintf("%s/login", c.serverConfig.APIURL))

	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		c.logger.Errorf("Login failed - URL: %s/login, Status: %d, Response: %s",
			c.serverConfig.APIURL, resp.StatusCode(), string(resp.Body()))
		return fmt.Errorf("login failed with status code: %d, response: %s", resp.StatusCode(), string(resp.Body()))
	}

	var apiResp XrayAPIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("login failed: %s", apiResp.Msg)
	}

	// Store cookies for future requests
	cookies := resp.Cookies()
	if len(cookies) > 0 {
		c.cookieCache.Set("session", cookies, cache.DefaultExpiration)
		c.logger.Info("Successfully logged in to X-ray API")
		return nil
	}

	return errors.New("no session cookie received from server")
}

// GetInbounds gets the inbounds from the X-ray API
func (c *Client) GetInbounds(ctx context.Context) ([]models.Inbound, error) {
	if err := c.Login(ctx); err != nil {
		return nil, err
	}

	cookies, _ := c.cookieCache.Get("session")

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetCookies(cookies.([]*http.Cookie)).
		Get(fmt.Sprintf("%s/xui/API/inbounds", c.serverConfig.APIURL))

	if err != nil {
		return nil, fmt.Errorf("get inbounds request failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		if resp.StatusCode() == http.StatusUnauthorized {
			c.cookieCache.Delete("session")
			return c.GetInbounds(ctx)
		}
		c.logger.Errorf("Get inbounds failed - Status: %d, Response: %s", resp.StatusCode(), string(resp.Body()))
		return nil, fmt.Errorf("get inbounds failed with status code: %d, response: %s", resp.StatusCode(), string(resp.Body()))
	}

	var apiResp XrayAPIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse inbounds response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("get inbounds failed: %s", apiResp.Msg)
	}

	// Convert obj to JSON and then unmarshal to inbounds
	objJSON, err := json.Marshal(apiResp.Obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inbounds obj: %w", err)
	}

	var inbounds []models.Inbound
	if err := json.Unmarshal(objJSON, &inbounds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inbounds: %w", err)
	}

	return inbounds, nil
}

// AddClientToInbound adds a client to an inbound
func (c *Client) AddClientToInbound(ctx context.Context, inboundID int, client models.Client) error {
	if err := c.Login(ctx); err != nil {
		return err
	}

	cookies, _ := c.cookieCache.Get("session")

	// Create settings object with clients array
	settings := map[string]interface{}{
		"clients": []map[string]interface{}{client.ToDictionary()},
	}

	// Convert settings to JSON string
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		c.logger.Errorf("Failed to marshal settings: %v", err)
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Create request body
	requestBody := map[string]interface{}{
		"id":       inboundID,
		"settings": string(settingsJSON),
	}

	// Log request details
	c.logger.Infof("Adding client to inbound %d with email: %s", inboundID, client.Email)
	c.logger.Debugf("Request body: %+v", requestBody)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetCookies(cookies.([]*http.Cookie)).
		SetBody(requestBody).
		Post(fmt.Sprintf("%s/xui/API/inbounds/addClient", c.serverConfig.APIURL))

	if err != nil {
		c.logger.Errorf("Add client request failed: %v", err)
		return fmt.Errorf("add client request failed: %w", err)
	}

	// Log response details
	c.logger.Debugf("Response status: %d", resp.StatusCode())
	c.logger.Debugf("Response body: %s", string(resp.Body()))

	if resp.StatusCode() != http.StatusOK {
		// If unauthorized, try to login again
		if resp.StatusCode() == http.StatusUnauthorized {
			c.cookieCache.Delete("session")
			return c.AddClientToInbound(ctx, inboundID, client)
		}
		c.logger.Errorf("Add client failed with status code %d, response body: %s", resp.StatusCode(), string(resp.Body()))
		return fmt.Errorf("add client failed with status code: %d", resp.StatusCode())
	}

	// Check if response body is empty
	if len(resp.Body()) == 0 {
		c.logger.Errorf("Empty response body from server")
		return fmt.Errorf("empty response from server")
	}

	var apiResp XrayAPIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		c.logger.Errorf("Failed to parse add client response: %v, response body: %s", err, string(resp.Body()))
		return fmt.Errorf("failed to parse add client response: %w, body: %s", err, string(resp.Body()))
	}

	if !apiResp.Success {
		c.logger.Errorf("Add client failed with message: %s", apiResp.Msg)
		return fmt.Errorf("add client failed: %s", apiResp.Msg)
	}

	c.logger.Infof("Successfully added client %s to inbound %d", client.Email, inboundID)
	return nil
}

// RemoveClients removes clients from inbounds
func (c *Client) RemoveClients(ctx context.Context, emails []string) error {
	if err := c.Login(ctx); err != nil {
		return err
	}

	cookies, _ := c.cookieCache.Get("session")

	// Get all inbounds to find clients
	inbounds, err := c.GetInbounds(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inbounds: %w", err)
	}

	// Track deletion results
	var deletionErrors []string
	successfullyDeleted := false

	// For each email, find and delete from all inbounds
	for _, email := range emails {
		emailDeleted := false

		// Search through all inbounds
		for _, inbound := range inbounds {
			// Parse inbound settings to find client UUID
			var settings models.InboundSettings
			if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
				c.logger.Errorf("Failed to parse settings for inbound %d: %v", inbound.ID, err)
				continue
			}

			// Find client by email
			for _, client := range settings.Clients {
				// Ищем по базовому имени используя helper функцию
				if helpers.IsEmailMatchingBaseUsername(client.Email, email) {
					c.logger.Infof("Found matching client: %s in inbound %d", client.Email, inbound.ID)

					// Extract client UUID from client object
					// The client UUID is typically stored in the client object
					// We need to find the actual UUID field
					clientUUID := c.extractClientUUID(client, client.Email)
					if clientUUID == "" {
						c.logger.Errorf("Failed to extract UUID for client %s in inbound %d", client.Email, inbound.ID)
						continue
					}

					// Delete client using the correct API endpoint
					err := c.deleteClientFromInbound(ctx, cookies.([]*http.Cookie), inbound.ID, clientUUID)
					if err != nil {
						c.logger.Errorf("Failed to delete client %s from inbound %d: %v", client.Email, inbound.ID, err)
						deletionErrors = append(deletionErrors, fmt.Sprintf("Failed to delete %s from inbound %d: %v", client.Email, inbound.ID, err))
					} else {
						c.logger.Infof("Successfully deleted client %s from inbound %d", client.Email, inbound.ID)
						emailDeleted = true
						successfullyDeleted = true
					}
				}
			}
		}

		if !emailDeleted {
			c.logger.Warnf("Client %s not found in any inbound", email)
			deletionErrors = append(deletionErrors, fmt.Sprintf("Client %s not found in any inbound", email))
		}
	}

	// Return error if no clients were successfully deleted
	if !successfullyDeleted {
		c.logger.Errorf("No clients were successfully deleted. Errors: %s", strings.Join(deletionErrors, "; "))
		return fmt.Errorf("failed to delete any clients: %s", strings.Join(deletionErrors, "; "))
	}

	// Log warnings for any errors that occurred
	if len(deletionErrors) > 0 {
		c.logger.Warnf("Some deletion errors occurred: %s", strings.Join(deletionErrors, "; "))
	}

	c.logger.Infof("RemoveClients operation completed successfully")
	return nil
}

// deleteClientFromInbound deletes a client from a specific inbound using the correct API endpoint
func (c *Client) deleteClientFromInbound(ctx context.Context, cookies []*http.Cookie, inboundID int, clientUUID string) error {
	c.logger.Debugf("Deleting client with UUID %s from inbound %d", clientUUID, inboundID)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetCookies(cookies).
		Post(fmt.Sprintf("%s/xui/API/inbounds/%d/delClient/%s", c.serverConfig.APIURL, inboundID, clientUUID))

	if err != nil {
		return fmt.Errorf("delete client request failed: %w", err)
	}

	c.logger.Debugf("Delete client response status: %d, body: %s", resp.StatusCode(), string(resp.Body()))

	if resp.StatusCode() != http.StatusOK {
		// If unauthorized, try to login again
		if resp.StatusCode() == http.StatusUnauthorized {
			c.cookieCache.Delete("session")
			return c.deleteClientFromInbound(ctx, cookies, inboundID, clientUUID)
		}
		return fmt.Errorf("delete client failed with status code: %d, response: %s", resp.StatusCode(), string(resp.Body()))
	}

	var apiResp XrayAPIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		return fmt.Errorf("failed to parse delete client response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("delete client failed: %s", apiResp.Msg)
	}

	return nil
}

// extractClientUUID extracts the UUID from a client object
// This method needs to be implemented based on the actual structure of the client object
func (c *Client) extractClientUUID(client models.InboundClient, email string) string {
	// Use client.ID as UUID, as in the working C# implementation
	if client.ID != "" {
		return client.ID
	}

	// Fallback to SubID if ID is empty
	if client.SubID != "" {
		return client.SubID
	}

	// If both are empty, use email as fallback
	return email
}

// GetOnlineUsers gets the online users
func (c *Client) GetOnlineUsers(ctx context.Context) ([]string, error) {
	if err := c.Login(ctx); err != nil {
		return nil, err
	}

	cookies, _ := c.cookieCache.Get("session")

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetCookies(cookies.([]*http.Cookie)).
		Post(fmt.Sprintf("%s/xui/API/inbounds/onlines", c.serverConfig.APIURL))

	if err != nil {
		return nil, fmt.Errorf("get online users request failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		// If unauthorized, try to login again
		if resp.StatusCode() == http.StatusUnauthorized {
			c.cookieCache.Delete("session")
			return c.GetOnlineUsers(ctx)
		}
		return nil, fmt.Errorf("get online users failed with status code: %d", resp.StatusCode())
	}

	var apiResp XrayAPIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse online users response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("get online users failed: %s", apiResp.Msg)
	}

	// Convert obj to JSON and then unmarshal to string array
	objJSON, err := json.Marshal(apiResp.Obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal online users obj: %w", err)
	}

	var onlineUsers []string
	if err := json.Unmarshal(objJSON, &onlineUsers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal online users: %w", err)
	}

	return onlineUsers, nil
}

// ResetUserTraffic resets a user's traffic
func (c *Client) ResetUserTraffic(ctx context.Context, inboundID int, email string) error {
	if err := c.Login(ctx); err != nil {
		return err
	}

	cookies, _ := c.cookieCache.Get("session")

	c.logger.Debugf("Resetting traffic for client %s in inbound %d", email, inboundID)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetCookies(cookies.([]*http.Cookie)).
		Post(fmt.Sprintf("%s/xui/API/inbounds/%d/resetClientTraffic/%s", c.serverConfig.APIURL, inboundID, email))

	if err != nil {
		return fmt.Errorf("reset user traffic request failed: %w", err)
	}

	c.logger.Debugf("Reset traffic response status: %d, body: %s", resp.StatusCode(), string(resp.Body()))

	if resp.StatusCode() != http.StatusOK {
		// If unauthorized, try to login again
		if resp.StatusCode() == http.StatusUnauthorized {
			c.cookieCache.Delete("session")
			return c.ResetUserTraffic(ctx, inboundID, email)
		}
		return fmt.Errorf("reset user traffic failed with status code: %d, response: %s", resp.StatusCode(), string(resp.Body()))
	}

	var apiResp XrayAPIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		return fmt.Errorf("failed to parse reset user traffic response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("reset user traffic failed: %s", apiResp.Msg)
	}

	return nil
}

// GetSubscriptionURL gets a user's subscription URL
func (c *Client) GetSubscriptionURL(ctx context.Context, email string) (string, error) {
	if c.serverConfig.SubURLPrefix == "" {
		return "", errors.New("subscription URL prefix not configured for this server")
	}

	return fmt.Sprintf("%s/sub/%s", c.serverConfig.SubURLPrefix, email), nil
}
